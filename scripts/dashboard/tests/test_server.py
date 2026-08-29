import importlib.util
from functools import partial
from http.server import ThreadingHTTPServer
import json
from pathlib import Path
import tempfile
import threading
import unittest
from unittest.mock import Mock, patch
from urllib.error import HTTPError
from urllib.request import urlopen


SERVER_PATH = Path(__file__).parents[1] / "server.py"
spec = importlib.util.spec_from_file_location("devarch_dashboard_server", SERVER_PATH)
server = importlib.util.module_from_spec(spec)
spec.loader.exec_module(server)


class InventoryTests(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.TemporaryDirectory()
        self.root = Path(self.temp_dir.name)
        (self.root / "apps").mkdir()
        (self.root / "services-library").mkdir()

    def tearDown(self):
        self.temp_dir.cleanup()

    def test_discovers_projects_and_detects_common_types(self):
        wordpress = self.root / "apps" / "client-site"
        wordpress.mkdir()
        (wordpress / "wp-config.php").write_text("<?php\n")

        next_app = self.root / "apps" / "storefront"
        next_app.mkdir()
        (next_app / "package.json").write_text(
            json.dumps({"dependencies": {"next": "latest"}})
        )
        (self.root / "apps" / ".devarch-backups").mkdir()

        projects = server.discover_projects(self.root)

        self.assertEqual([project["name"] for project in projects], ["client-site", "storefront"])
        self.assertEqual(projects[0]["kind"], "WordPress")
        self.assertEqual(projects[0]["url"], "https://client-site.test")
        self.assertEqual(projects[1]["kind"], "Next.js")

    def test_discovers_catalog_services_in_deterministic_order(self):
        for service_id in ("database/postgres", "ai/ollama"):
            compose = self.root / "services-library" / service_id / "compose.yml"
            compose.parent.mkdir(parents=True)
            compose.write_text("services: {}\n")

        services = server.discover_services(self.root)

        self.assertEqual([item["id"] for item in services], ["ai/ollama", "database/postgres"])
        self.assertEqual(
            services[1]["command"],
            "cd services-library/database/postgres && podman compose up -d",
        )

    def test_normalizes_running_podman_containers_without_labels_or_commands(self):
        result = Mock(
            returncode=0,
            stdout=json.dumps(
                [
                    {
                        "Id": "abcdef1234567890",
                        "Names": ["mailpit"],
                        "Image": "docker.io/axllent/mailpit:latest",
                        "State": "running",
                        "Status": "Up 2 hours (healthy)",
                        "Ports": [
                            {
                                "host_ip": "127.0.0.1",
                                "host_port": 8025,
                                "container_port": 8025,
                                "protocol": "tcp",
                            }
                        ],
                        "Labels": {"contains": "untrusted metadata"},
                        "Command": ["secret", "argument"],
                    }
                ]
            ),
            stderr="",
        )

        containers, error = server.read_containers(run=lambda *args, **kwargs: result)

        self.assertIsNone(error)
        self.assertEqual(containers[0]["name"], "mailpit")
        self.assertEqual(containers[0]["id"], "abcdef123456")
        self.assertEqual(containers[0]["openUrl"], "http://127.0.0.1:8025")
        self.assertNotIn("labels", containers[0])
        self.assertNotIn("command", containers[0])

    def test_podman_socket_inventory_avoids_cli_namespace_requirements(self):
        raw = [{"Id": "1234567890abcdef", "Names": ["redis"], "Image": "redis:latest", "State": "running", "Status": "running", "Ports": []}]
        cli = Mock(side_effect=AssertionError("CLI should not be used"))
        with patch.object(server, "_read_podman_socket", return_value=(raw, None)):
            containers, error = server.read_containers(
                run=cli,
                socket_path=Path("/run/user/1000/podman/podman.sock"),
            )

        self.assertIsNone(error)
        self.assertEqual(containers[0]["name"], "redis")
        cli.assert_not_called()

    def test_missing_podman_is_reported_without_failing_inventory(self):
        def missing(*args, **kwargs):
            raise FileNotFoundError("podman")

        containers, error = server.read_containers(run=missing)

        self.assertEqual(containers, [])
        self.assertEqual(error, "Podman is not installed or is not available on PATH.")

    def test_host_allowlist_accepts_canonical_and_local_hosts_only(self):
        self.assertTrue(server.is_allowed_host("devarch.test"))
        self.assertTrue(server.is_allowed_host("devarch.test:443"))
        self.assertTrue(server.is_allowed_host("127.0.0.1:7411"))
        self.assertFalse(server.is_allowed_host("192.168.1.50:7411"))
        self.assertFalse(server.is_allowed_host("attacker.example"))

    def test_supported_browser_routes_return_dashboard_shell(self):
        (self.root / "index.html").write_text("<title>DevArch route shell</title>")
        handler = partial(server.DashboardHandler, directory=str(self.root))
        httpd = ThreadingHTTPServer(("127.0.0.1", 0), handler)
        thread = threading.Thread(target=httpd.serve_forever, daemon=True)
        thread.start()
        try:
            for route in (
                "/apps",
                "/apps/client-site",
                "/projects",
                "/projects/client-site",
                "/containers",
                "/services",
                "/services/database/postgres",
            ):
                with urlopen(f"http://127.0.0.1:{httpd.server_port}{route}") as response:
                    self.assertIn(b"DevArch route shell", response.read())
            with self.assertRaises(HTTPError) as missing:
                urlopen(f"http://127.0.0.1:{httpd.server_port}/unsupported")
            self.assertEqual(missing.exception.code, 404)
            missing.exception.close()
        finally:
            httpd.shutdown()
            httpd.server_close()
            thread.join()

    def test_inventory_api_returns_json_without_http_caching(self):
        server.DashboardHandler.repo_root = self.root
        handler = partial(server.DashboardHandler, directory=self.temp_dir.name)
        httpd = ThreadingHTTPServer(("127.0.0.1", 0), handler)
        thread = threading.Thread(target=httpd.serve_forever, daemon=True)
        thread.start()
        try:
            with patch.object(server, "read_containers", return_value=([], None)):
                with urlopen(
                    f"http://127.0.0.1:{httpd.server_port}/api/inventory"
                ) as response:
                    payload = json.load(response)
                    self.assertEqual(response.headers["Cache-Control"], "no-store")
                    self.assertEqual(response.headers.get_content_type(), "application/json")
            self.assertEqual(payload["projects"], [])
            self.assertEqual(payload["services"], [])
            self.assertIsNone(payload["runtimeError"])
        finally:
            httpd.shutdown()
            httpd.server_close()
            thread.join()


if __name__ == "__main__":
    unittest.main()
