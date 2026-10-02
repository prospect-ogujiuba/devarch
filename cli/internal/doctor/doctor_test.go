package doctor

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prospect-ogujiuba/devarch/cli/internal/catalog"
	"github.com/prospect-ogujiuba/devarch/cli/internal/engine"
	"github.com/prospect-ogujiuba/devarch/cli/internal/hosts"
	"github.com/prospect-ogujiuba/devarch/cli/internal/runner"
	"github.com/prospect-ogujiuba/devarch/cli/internal/state"
	"github.com/prospect-ogujiuba/devarch/cli/internal/testutil"
)

func writeCert(t *testing.T, path string, names []string, notAfter time.Time) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}, DNSNames: names,
		NotBefore: notAfter.Add(-365 * 24 * time.Hour), NotAfter: notAfter}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644)
}

func setup(t *testing.T) (Env, *runner.Fake, string) {
	t.Helper()
	repo := t.TempDir()
	lib := filepath.Join(repo, "services-library")
	testutil.WriteFiles(t, lib, map[string]string{
		"proxy/nginx-proxy-manager/compose.yml": "x-devarch:\n  title: NPM\nservices:\n  nginx-proxy-manager:\n    image: npm:2\n    container_name: nginx-proxy-manager\n    ports: [\"127.0.0.1:80:80\"]\n",
		"backend/php/compose.yml":               "x-devarch:\n  title: PHP\nservices:\n  php:\n    image: php:8\n    container_name: php\n    ports: [\"127.0.0.1:8080:8080\"]\n",
	})
	writeCert(t, filepath.Join(lib, "proxy/nginx-proxy-manager/config/certs/local.crt"), []string{"*.test"}, time.Now().Add(400*24*time.Hour))
	c, _ := catalog.Load([]catalog.Path{{Dir: lib}})
	f := &runner.Fake{Responses: map[string]runner.Response{
		"podman --version": {Out: []byte("podman version 6.1.2\n")},
		"podman info":      {Out: []byte("true\n")},
		"podman ps":        {Out: []byte(`[{"Names":["nginx-proxy-manager"],"State":"running","Status":"Up","Labels":{"com.docker.compose.project":"nginx-proxy-manager","com.docker.compose.project.working_dir":"` + lib + `/proxy/nginx-proxy-manager"}}]`)},
		"loginctl":         {Out: []byte("Linger=no\n")},
	}}
	hostsFile := filepath.Join(t.TempDir(), "hosts")
	os.WriteFile(hostsFile, []byte("127.0.0.1 localhost\n"), 0o644)
	portStart := filepath.Join(t.TempDir(), "port_start")
	os.WriteFile(portStart, []byte("1024\n"), 0o644)
	e := &engine.Engine{Root: repo, Catalog: c, Runner: f, Settings: state.Defaults(repo), Hosts: &hosts.Manager{Platform: hosts.Unix, File: hostsFile, Runner: f}}
	return Env{Engine: e, RootHow: "config", PortStartFile: portStart, PortFree: func(string, int) bool { return true }}, f, lib
}

func byName(cs []Check) map[string]Check {
	m := map[string]Check{}
	for _, c := range cs {
		m[c.Name] = c
	}
	return m
}

func TestDoctor(t *testing.T) {
	env, _, _ := setup(t)
	env.PortFree = func(ip string, port int) bool { return port != 8080 }
	checks := byName(Run(context.Background(), env))
	expect := map[string]Status{
		"checkout": OK, "catalog": OK, "podman": OK, "compose provider": OK, "rootless": OK,
		"network": OK, "proxy": OK, "host ports": Warn, "certificate": OK, "hosts": Warn,
	}
	for name, st := range expect {
		if checks[name].Status != st {
			t.Errorf("%s = %+v, want %s", name, checks[name], st)
		}
	}
	if !strings.Contains(checks["host ports"].Detail, "8080 (php)") {
		t.Errorf("host ports detail: %s", checks["host ports"].Detail)
	}
	if c, ok := checks["ports 80/443"]; ok && c.Status != Fail {
		t.Errorf("port start 1024 must fail: %+v", c)
	}
	if c, ok := checks["lingering"]; ok && (c.Status != Warn || !strings.Contains(c.Fix, "enable-linger")) {
		t.Errorf("lingering: %+v", c)
	}
}

func TestDoctorStopsWithoutPodman(t *testing.T) {
	env, f, _ := setup(t)
	f.Responses["podman --version"] = runner.Response{Err: errors.New("executable file not found")}
	checks := Run(context.Background(), env)
	last := checks[len(checks)-1]
	if last.Name != "podman" || last.Status != Fail {
		t.Fatalf("got %+v", checks)
	}
}

func TestCertificateStates(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "services-library/proxy/nginx-proxy-manager/config/certs/local.crt")
	now := time.Now()
	if c := certificate(root, now); c.Status != Fail {
		t.Fatalf("missing cert: %+v", c)
	}
	writeCert(t, path, []string{"*.test"}, now.Add(10*24*time.Hour))
	if c := certificate(root, now); c.Status != Warn || !strings.Contains(c.Detail, "expires") {
		t.Fatalf("expiring: %+v", c)
	}
	writeCert(t, path, []string{"*.test"}, now.Add(-time.Hour))
	if c := certificate(root, now); c.Status != Fail {
		t.Fatalf("expired: %+v", c)
	}
	writeCert(t, path, []string{"*.dev"}, now.Add(400*24*time.Hour))
	if c := certificate(root, now); c.Status != Warn {
		t.Fatalf("wrong names: %+v", c)
	}
}

func TestDoctorUsesSettings(t *testing.T) {
	env, f, _ := setup(t)
	env.Engine.Settings.Network = "dev-net"
	env.Engine.Settings.HostsManage = false
	f.Responses["podman network exists dev-net"] = runner.Response{Err: &runner.ExitError{Code: 1}}
	checks := byName(Run(context.Background(), env))
	if c := checks["network"]; c.Status != Warn || !strings.Contains(c.Fix, "podman network create dev-net") {
		t.Errorf("network: %+v", c)
	}
	if c := checks["hosts"]; c.Status != OK || !strings.Contains(c.Detail, "not managed") {
		t.Errorf("hosts: %+v", c)
	}
}

func TestDoctorDocker(t *testing.T) {
	env, f, lib := setup(t)
	env.Engine.Settings.Runtime = "docker"
	f.Responses["docker --version"] = runner.Response{Out: []byte("Docker version 28.1.1\n")}
	f.Responses["docker info"] = runner.Response{Err: errors.New("permission denied while trying to connect to the Docker daemon socket")}
	f.Responses["docker ps"] = runner.Response{Out: []byte(`{"Names":"nginx-proxy-manager","State":"running","Status":"Up","Labels":"com.docker.compose.project=nginx-proxy-manager,com.docker.compose.project.working_dir=` + lib + `/proxy/nginx-proxy-manager","Ports":""}` + "\n")}
	env.DockerGroup = func() (bool, bool) { return true, false }
	checks := byName(Run(context.Background(), env))
	if c := checks["docker"]; c.Status != OK || !strings.Contains(c.Detail, "Docker version") {
		t.Errorf("docker: %+v", c)
	}
	if c := checks["docker access"]; c.Status != Fail || !strings.Contains(c.Fix, "usermod -aG docker") {
		t.Errorf("docker access: %+v", c)
	}
	if c := checks["proxy"]; c.Status != OK {
		t.Errorf("proxy: %+v", c)
	}
	for _, podmanOnly := range []string{"podman", "rootless", "ports 80/443", "lingering"} {
		if _, ok := checks[podmanOnly]; ok {
			t.Errorf("docker doctor ran the %s check", podmanOnly)
		}
	}
	for _, l := range f.Lines() {
		if strings.HasPrefix(l, "podman ") {
			t.Errorf("ran %s", l)
		}
	}

	f.Responses["docker info"] = runner.Response{Out: []byte("28.1.1\n")}
	env.DockerGroup = func() (bool, bool) { return true, true }
	if c := byName(Run(context.Background(), env))["docker access"]; c.Status != OK {
		t.Errorf("member: %+v", c)
	}
	env.DockerGroup = func() (bool, bool) { return false, false }
	if c := byName(Run(context.Background(), env))["docker access"]; c.Status != OK {
		t.Errorf("Docker Desktop without a docker group: %+v", c)
	}
}
