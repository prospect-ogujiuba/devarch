#!/usr/bin/env python3
"""Small, read-only localhost dashboard for a DevArch checkout."""

from __future__ import annotations

import argparse
from datetime import datetime, timezone
from functools import partial
import http.client
from http.server import SimpleHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import socket
import subprocess
from typing import Any, Callable
from urllib.parse import urlsplit


DASHBOARD_DIR = Path(__file__).resolve().parent
DEFAULT_REPO_ROOT = DASHBOARD_DIR.parents[1]
STATIC_DIR = DASHBOARD_DIR / "static"
ALLOWED_HOSTS = {"devarch.test", "dashboard.test", "localhost", "127.0.0.1", "::1"}


def is_allowed_host(host_header: str) -> bool:
    host = host_header.strip().lower()
    if host.startswith("[") and "]" in host:
        hostname = host[1 : host.index("]")]
    elif host.count(":") == 1:
        hostname = host.rsplit(":", 1)[0]
    else:
        hostname = host
    return hostname in ALLOWED_HOSTS


def is_browser_route(path: str) -> bool:
    parts = [part for part in path.strip("/").split("/") if part]
    if not parts:
        return True
    if parts in (["apps"], ["projects"], ["containers"], ["services"]):
        return True
    if len(parts) == 2 and parts[0] in {"apps", "projects"}:
        return True
    return len(parts) == 3 and parts[0] == "services"


def _project_kind(path: Path) -> str:
    if (path / "wp-config.php").is_file() or (path / "wp-content").is_dir():
        return "WordPress"
    if (path / "artisan").is_file():
        return "Laravel"

    package_file = path / "package.json"
    if package_file.is_file():
        try:
            package = json.loads(package_file.read_text(encoding="utf-8"))
            dependencies = {
                **package.get("dependencies", {}),
                **package.get("devDependencies", {}),
            }
        except (OSError, json.JSONDecodeError, TypeError):
            return "JavaScript"

        frameworks = (
            ("next", "Next.js"),
            ("nuxt", "Nuxt"),
            ("@sveltejs/kit", "SvelteKit"),
            ("astro", "Astro"),
            ("@angular/core", "Angular"),
            ("@react-router/node", "React Router"),
            ("@builder.io/qwik", "Qwik"),
            ("vite", "Vite"),
        )
        for dependency, label in frameworks:
            if dependency in dependencies:
                return label
        return "JavaScript"

    if (path / "composer.json").is_file():
        return "PHP"
    return "Project"


def discover_projects(repo_root: Path) -> list[dict[str, str]]:
    apps_root = repo_root / "apps"
    if not apps_root.is_dir():
        return []

    projects = []
    for path in sorted(apps_root.iterdir(), key=lambda item: item.name.casefold()):
        if not path.is_dir() or path.name.startswith("."):
            continue
        projects.append(
            {
                "name": path.name,
                "kind": _project_kind(path),
                "path": str(path.resolve()),
                "relativePath": path.relative_to(repo_root).as_posix(),
                "url": f"https://{path.name}.test",
            }
        )
    return projects


def discover_services(repo_root: Path) -> list[dict[str, str]]:
    catalog_root = repo_root / "services-library"
    if not catalog_root.is_dir():
        return []

    services = []
    for compose_file in sorted(catalog_root.glob("*/*/compose.y*ml")):
        relative_dir = compose_file.parent.relative_to(repo_root).as_posix()
        category, name = compose_file.parent.relative_to(catalog_root).parts
        services.append(
            {
                "id": f"{category}/{name}",
                "name": name,
                "category": category,
                "path": str(compose_file.parent.resolve()),
                "relativePath": relative_dir,
                "command": f"cd {relative_dir} && podman compose up -d",
            }
        )
    return services


def _container_name(value: Any) -> str:
    if isinstance(value, list):
        return str(value[0]) if value else "Unnamed"
    return str(value or "Unnamed")


def _normalize_ports(value: Any) -> list[dict[str, Any]]:
    if not isinstance(value, list):
        return []

    ports = []
    for item in value:
        if not isinstance(item, dict):
            continue
        host_port = item.get("host_port", item.get("hostPort"))
        container_port = item.get("container_port", item.get("containerPort"))
        if not host_port:
            continue
        host = str(item.get("host_ip", item.get("hostIP", "127.0.0.1")))
        if host in {"0.0.0.0", "::", ""}:
            host = "127.0.0.1"
        ports.append(
            {
                "host": host,
                "hostPort": int(host_port),
                "containerPort": int(container_port) if container_port else None,
                "protocol": str(item.get("protocol", "tcp")),
                "url": f"http://{host}:{int(host_port)}",
            }
        )
    return ports


class UnixHTTPConnection(http.client.HTTPConnection):
    def __init__(self, socket_path: Path, timeout: float = 3) -> None:
        super().__init__("localhost", timeout=timeout)
        self.socket_path = socket_path

    def connect(self) -> None:
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.settimeout(self.timeout)
        self.sock.connect(str(self.socket_path))


def podman_socket_path() -> Path:
    runtime_dir = os.environ.get("XDG_RUNTIME_DIR", f"/run/user/{os.getuid()}")
    return Path(runtime_dir) / "podman" / "podman.sock"


def _read_podman_socket(socket_path: Path) -> tuple[list[Any] | None, str | None]:
    connection = UnixHTTPConnection(socket_path)
    try:
        connection.request("GET", "/v1.0.0/libpod/containers/json?all=false")
        response = connection.getresponse()
        body = response.read()
        if response.status != 200:
            return None, f"Podman API returned HTTP {response.status}."
        payload = json.loads(body)
        if not isinstance(payload, list):
            return None, "Podman API returned an unexpected response."
        return payload, None
    except (OSError, http.client.HTTPException, json.JSONDecodeError) as error:
        return None, f"Podman API unavailable: {error}."
    finally:
        connection.close()


def _normalize_containers(raw_containers: Any) -> list[dict[str, Any]]:
    containers = []
    for raw in raw_containers if isinstance(raw_containers, list) else []:
        if not isinstance(raw, dict):
            continue
        ports = _normalize_ports(raw.get("Ports"))
        identifier = str(raw.get("Id", raw.get("ID", "")))
        containers.append(
            {
                "id": identifier[:12],
                "name": _container_name(raw.get("Names", raw.get("Name"))),
                "image": str(raw.get("Image", "Unknown image")),
                "state": str(raw.get("State", "unknown")),
                "status": str(raw.get("Status", "")),
                "ports": ports,
                "openUrl": ports[0]["url"] if ports else None,
            }
        )
    containers.sort(key=lambda item: item["name"].casefold())
    return containers


def read_containers(
    run: Callable[..., subprocess.CompletedProcess[str]] = subprocess.run,
    socket_path: Path | None = None,
) -> tuple[list[dict[str, Any]], str | None]:
    if socket_path is not None:
        raw_containers, socket_error = _read_podman_socket(socket_path)
        if raw_containers is not None:
            return _normalize_containers(raw_containers), None
    else:
        socket_error = None

    try:
        result = run(
            ["podman", "ps", "--format", "json"],
            capture_output=True,
            text=True,
            check=False,
            timeout=10,
        )
    except FileNotFoundError:
        return [], socket_error or "Podman is not installed or is not available on PATH."
    except subprocess.TimeoutExpired:
        return [], socket_error or "Podman did not respond within 10 seconds."
    except OSError as error:
        return [], socket_error or f"Podman could not be started: {error}."

    if result.returncode != 0:
        detail = (result.stderr or "Podman returned an error.").strip().splitlines()[0]
        return [], socket_error or f"Podman inventory unavailable: {detail}"

    try:
        raw_containers = json.loads(result.stdout or "[]")
    except json.JSONDecodeError:
        return [], socket_error or "Podman returned invalid JSON."

    return _normalize_containers(raw_containers), None


def build_inventory(repo_root: Path) -> dict[str, Any]:
    containers, runtime_error = read_containers(socket_path=podman_socket_path())
    return {
        "generatedAt": datetime.now(timezone.utc).isoformat(),
        "projects": discover_projects(repo_root),
        "containers": containers,
        "services": discover_services(repo_root),
        "runtimeError": runtime_error,
    }


class DashboardHandler(SimpleHTTPRequestHandler):
    repo_root = DEFAULT_REPO_ROOT

    def do_GET(self) -> None:
        if not is_allowed_host(self.headers.get("Host", "")):
            self.send_error(421, "This local dashboard is available at devarch.test.")
            return
        request_path = urlsplit(self.path).path
        if request_path.rstrip("/") == "/api/inventory":
            payload = json.dumps(build_inventory(self.repo_root)).encode("utf-8")
            self.send_response(200)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
            return
        if is_browser_route(request_path):
            self.path = "/index.html"
        super().do_GET()

    def log_message(self, format: str, *args: Any) -> None:
        print(f"dashboard: {format % args}")


def main() -> None:
    parser = argparse.ArgumentParser(description="Serve the DevArch discovery dashboard.")
    parser.add_argument("--host", default="127.0.0.1")
    parser.add_argument("--port", default=7411, type=int)
    parser.add_argument("--repo-root", default=DEFAULT_REPO_ROOT, type=Path)
    args = parser.parse_args()

    DashboardHandler.repo_root = args.repo_root.resolve()
    handler = partial(DashboardHandler, directory=str(STATIC_DIR))
    server = ThreadingHTTPServer((args.host, args.port), handler)
    print(f"DevArch dashboard: http://{args.host}:{args.port}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        pass
    finally:
        server.server_close()


if __name__ == "__main__":
    main()
