package cmd

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/MarcelArt/stacker/internal/scaffold"
)

func runCommand(args ...string) (string, error) {
	root := newRootCommand()
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs(args)
	err := root.Execute()
	return output.String(), err
}

func TestRootHelpNeedsNoComposeFile(t *testing.T) {
	t.Chdir(t.TempDir())
	output, err := runCommand()
	if err != nil || !strings.Contains(output, "add") {
		t.Fatalf("root help failed: %v; %s", err, output)
	}
}

func TestAddEnvironmentConflictsAndDependencyGuidance(t *testing.T) {
	folder := t.TempDir()
	t.Chdir(folder)
	if err := os.Mkdir("deploy", 0755); err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join("deploy", ".env")
	if err := os.WriteFile(envPath, []byte("POSTGRES_PASSWORD=existing-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := runCommand("-f", "deploy/compose.yml", "-n", "custom", "add", "n8n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "stacker add postgres") || !strings.Contains(output, "deploy/.env") || strings.Contains(output, "existing-secret") {
		t.Fatalf("incorrect setup guidance: %s", output)
	}
	env, _ := os.ReadFile(envPath)
	if !strings.Contains(string(env), "POSTGRES_PASSWORD=existing-secret\n") || strings.Count(string(env), "POSTGRES_PASSWORD=") != 1 {
		t.Fatal("existing environment entry was changed")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Fatal("environment file was written into the working directory")
	}
	before, _ := os.ReadFile("deploy/compose.yml")
	if _, err := runCommand("-f", "deploy/compose.yml", "-n", "custom", "add", "n8n"); err == nil {
		t.Fatal("repeated add must require --overwrite")
	}
	after, _ := os.ReadFile("deploy/compose.yml")
	if !bytes.Equal(before, after) {
		t.Fatal("conflict changed Compose")
	}
	if _, err := runCommand("-f", "deploy/compose.yml", "-n", "custom", "add", "n8n", "--overwrite"); err != nil {
		t.Fatal(err)
	}
	secondEnv, _ := os.ReadFile(envPath)
	if !bytes.Equal(env, secondEnv) {
		t.Fatal("overwriting a service duplicated environment entries")
	}
}

func TestManagementCommandsPreserveRichComposeAndHostMode(t *testing.T) {
	t.Chdir(t.TempDir())
	data := []byte(`
name: rich-stack
x-preserve: {arbitrary: data}
services:
  host:
    image: example/host
    network_mode: host
    cap_add: [NET_ADMIN]
  app:
    image: example/app
    environment: [A=B]
    depends_on: {host: {condition: service_started}}
    networks: {custom: {aliases: [app-alias]}}
    deploy: {replicas: 2}
    ports: [{target: 80, published: "8080", protocol: tcp}]
  remove: {image: example/remove}
networks:
  custom: {external: true, name: custom, labels: {owner: stacker}}
volumes: {keep: {labels: {owner: stacker}}}
`)
	if err := os.WriteFile("compose.yml", data, 0644); err != nil {
		t.Fatal(err)
	}
	output, err := runCommand("-f", "compose.yml", "ls")
	if err != nil || !strings.Contains(output, "8080:80/tcp") || !strings.Contains(output, "host") {
		t.Fatalf("list failed: %v; %s", err, output)
	}
	if _, err := runCommand("-f", "compose.yml", "config"); err != nil {
		t.Fatal(err)
	}
	// The default is read from stacker.toml. External networks omit driver.
	if _, err := runCommand("-f", "compose.yml", "add", "cloudflared"); err == nil || !strings.Contains(err.Error(), "conflicting networks") {
		t.Fatal("expected a resource conflict for the labelled network")
	}
	if _, err := runCommand("-f", "compose.yml", "rm", "-c", "remove"); err != nil {
		t.Fatal(err)
	}
	expected, _ := scaffold.ParseCompose(data)
	delete(expected.Section("services"), "remove")
	actual, err := scaffold.ReadCompose("compose.yml", false)
	if err != nil || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("remove lost fields or host mode: %v", err)
	}
}

func TestMalformedComposeIsNeverOverwritten(t *testing.T) {
	t.Chdir(t.TempDir())
	data := []byte("services: [broken\n")
	if err := os.WriteFile("compose.yml", data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCommand("-f", "compose.yml", "add", "postgres"); err == nil {
		t.Fatal("malformed input should fail")
	}
	after, _ := os.ReadFile("compose.yml")
	if !bytes.Equal(after, data) {
		t.Fatal("malformed Compose was overwritten")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Fatal("failed add created .env")
	}
}

func TestAddPreservesHostModeAndUsesSavedExternalNetwork(t *testing.T) {
	t.Chdir(t.TempDir())
	data := []byte(`
services:
  host:
    image: alpine:3
    network_mode: host
    command: [sleep, infinity]
networks:
  custom: {name: custom, external: true}
`)
	if err := os.WriteFile("compose.yml", data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := runCommand("-f", "compose.yml", "config"); err != nil {
		t.Fatal(err)
	}
	if _, err := runCommand("-f", "compose.yml", "add", "redis"); err != nil {
		t.Fatal(err)
	}
	doc, err := scaffold.ReadCompose("compose.yml", false)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := scaffold.ParseCompose(data)
	if !reflect.DeepEqual(doc.Section("services")["host"], before.Section("services")["host"]) {
		t.Fatal("adding a service changed host networking")
	}
	network := doc.Section("networks")["custom"].(map[string]any)
	if network["external"] != true || network["driver"] != nil {
		t.Fatalf("incorrect external shared network: %v", network)
	}
	redis := doc.Section("services")["redis"].(map[string]any)
	if !reflect.DeepEqual(redis["networks"], []any{"custom"}) {
		t.Fatalf("saved network was not applied: %v", redis)
	}
	if _, err := os.Stat(".env"); err != nil {
		t.Fatal("env_file: .env needs a file even when there are no declared inputs")
	}
}

func TestHTTPSBundleWithHostNetworking(t *testing.T) {
	t.Chdir(t.TempDir())
	const source = `
x-stacker:
  description: Mixed host and shared networking
  shared_network: shared
  environment:
    APP_TOKEN: {description: Application token.}
services:
  host:
    image: alpine:3
    command: [sleep, infinity]
    network_mode: host
  app:
    image: nginx:alpine
    environment: {APP_TOKEN: "${APP_TOKEN}"}
    networks:
      shared: {aliases: [application]}
networks:
  shared: {driver: bridge}
`
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, source)
	}))
	defer server.Close()
	// Trust only this test certificate; exercise the real HTTPS command path.
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	originalTransport := http.DefaultTransport
	transport := originalTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: roots}
	http.DefaultTransport = transport
	defer func() {
		http.DefaultTransport = originalTransport
		transport.CloseIdleConnections()
	}()
	if _, err := runCommand("-n", "custom", "add", server.URL+"/template.yml"); err != nil {
		t.Fatal(err)
	}
	doc, err := scaffold.ReadCompose("docker-compose.yml", false)
	if err != nil {
		t.Fatal(err)
	}
	host := doc.Section("services")["host"].(map[string]any)
	if host["network_mode"] != "host" || host["networks"] != nil {
		t.Fatalf("HTTPS host service was changed: %v", host)
	}
	if _, present := doc["x-stacker"]; present {
		t.Fatal("Stacker metadata leaked into the assembled Compose file")
	}
	env, _ := os.ReadFile(".env")
	if string(env) != "# Application token.\nAPP_TOKEN=\n" {
		t.Fatalf("incorrect environment file: %s", env)
	}
	// The core suite works without Docker; use the real Compose validator when available.
	if _, err := exec.LookPath("docker"); err == nil {
		if err := exec.Command("docker", "compose", "version").Run(); err == nil {
			output, err := exec.Command("docker", "compose", "-f", "docker-compose.yml", "config", "--quiet").CombinedOutput()
			if err != nil {
				t.Fatalf("Docker Compose rejected the HTTPS host-mode bundle: %v\n%s", err, output)
			}
		}
	}
}
