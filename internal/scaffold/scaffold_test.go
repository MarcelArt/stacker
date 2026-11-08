package scaffold

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func mustTemplate(t *testing.T, source string) *Template {
	t.Helper()
	template, err := ParseTemplate([]byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return template
}

func TestBuiltinsAssembleWithCompleteEnvironmentMetadata(t *testing.T) {
	if len(Builtins()) != 17 {
		t.Fatalf("expected all 17 built-in templates, got %v", Builtins())
	}
	doc := Document{}
	for _, name := range Builtins() {
		template, err := LoadTemplate(context.Background(), name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := template.BindNetwork("custom", false); err != nil {
			t.Fatal(err)
		}
		doc, err = Merge(doc, template.Compose, false)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
	if len(doc.Section("services")) != 17 {
		t.Fatalf("unexpected services: %v", doc.Section("services"))
	}
	if missing := MissingDependencies(doc, doc); len(missing) != 0 {
		t.Fatalf("built-in dependencies not satisfied: %v", missing)
	}
}

func TestHostModeAndMappedNetworksInOneBundle(t *testing.T) {
	template := mustTemplate(t, `
x-stacker:
  shared_network: shared
services:
  host:
    image: example/host
    network_mode: host
    cap_add: [NET_ADMIN]
  app:
    image: example/app
    networks:
      shared:
        aliases: [application]
      private: {}
networks:
  shared: {driver: bridge}
  private: {internal: true}
`)
	if err := template.BindNetwork("public", true); err != nil {
		t.Fatal(err)
	}
	host := template.Compose.Section("services")["host"].(map[string]any)
	if host["network_mode"] != "host" || host["networks"] != nil {
		t.Fatalf("host mode changed: %v", host)
	}
	network := template.Compose.Section("networks")["public"].(map[string]any)
	if network["external"] != true || network["driver"] != nil {
		t.Fatalf("external network must not include driver options: %v", network)
	}
	app := template.Compose.Section("services")["app"].(map[string]any)
	refs := app["networks"].(map[string]any)
	if refs["public"] == nil || refs["private"] == nil || refs["shared"] != nil {
		t.Fatalf("network mapping not preserved: %v", refs)
	}
	path := filepath.Join(t.TempDir(), "compose.yml")
	if err := WriteCompose(path, template.Compose); err != nil {
		t.Fatal(err)
	}
	roundtrip, err := ReadCompose(path, false)
	if err != nil || !reflect.DeepEqual(roundtrip, template.Compose) {
		t.Fatalf("networking lost during write: %v", err)
	}
}

func TestHostOnlyDoesNotAddUnusedSharedNetwork(t *testing.T) {
	template := mustTemplate(t, `
x-stacker: {shared_network: shared}
services:
  host: {image: example/host, network_mode: host}
networks:
  shared: {driver: bridge}
`)
	if err := template.BindNetwork("public", false); err != nil {
		t.Fatal(err)
	}
	if len(template.Compose.Section("networks")) != 0 {
		t.Fatal("host-only template created an unused shared network")
	}
}

func TestInvalidTemplates(t *testing.T) {
	cases := []string{
		"<html>not YAML</html>",
		"services: {app: {image: nginx}}\n---\nservices: {}",
		"services: {app: {image: nginx}, app: {image: redis}}",
		"services: {app: {image: nginx, network_mode: host, networks: [default]}}",
		"services: {app: {image: nginx, network_mode: host, ports: ['80:80']}}",
		"services: {app: {image: nginx, networks: wrong}}",
		"services: {app: {image: nginx, command: '${UNDECLARED}'}}",
		"services: {app: {image: nginx}}\nx-stacker: {environment: {BAD-NAME: {description: invalid}}}",
		"services: {app: {image: nginx}}\nx-stacker: {environmnt: {}}",
	}
	for _, source := range cases {
		if _, err := ParseTemplate([]byte(source)); err == nil {
			t.Errorf("accepted invalid template: %s", source)
		}
	}
	// Container shell variables escaped with $$ are not Compose inputs.
	mustTemplate(t, "services: {app: {image: nginx, command: 'echo $$CONTAINER_VALUE'}}")
}

func TestMergePreservesUnknownFieldsAndRejectsResourceConflicts(t *testing.T) {
	existing, err := ParseCompose([]byte(`
name: custom-stack
x-custom: {keep: true}
services:
  original:
    image: example/original
    deploy: {replicas: 2}
    environment: [A=B]
    ports: [{target: 80, published: "8080", protocol: tcp}]
networks:
  net: {driver: bridge, external: false, name: net}
volumes:
  data:
`))
	if err != nil {
		t.Fatal(err)
	}
	incoming := mustTemplate(t, `
services: {added: {image: example/added}}
networks: {net: {driver: bridge, name: net}}
volumes: {data: {}}
`)
	merged, err := Merge(existing, incoming.Compose, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(merged["x-custom"], existing["x-custom"]) || !reflect.DeepEqual(merged.Section("services")["original"], existing.Section("services")["original"]) {
		t.Fatal("existing fields were changed")
	}
	if _, exists := existing.Section("services")["added"]; exists {
		t.Fatal("merge mutated its input")
	}
	if _, err := Merge(merged, incoming.Compose, false); err == nil {
		t.Fatal("existing services must require --overwrite, even when identical")
	}
	if _, err := Merge(merged, incoming.Compose, true); err != nil {
		t.Fatal(err)
	}
	incoming.Compose.Section("networks")["net"] = map[string]any{"external": true}
	if _, err := Merge(existing, incoming.Compose, true); err == nil {
		t.Fatal("--overwrite must not replace conflicting resources")
	}
}

func TestEnvironmentPreservesValuesAndQuotedMultilineContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	original := "# existing\r\nexport TOKEN=existing-secret\r\nEMPTY=\"\"\r\nUNSET\r\nMULTILINE='first\nNEW=inside-string\nlast'"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	requirements := map[string]EnvironmentRequirement{
		"TOKEN": {Description: "Keep the existing token."},
		"EMPTY": {Description: "An empty entry."},
		"NEW":   {Description: "First line.\nSecond line."},
		"UNSET": {Description: "Preserve a declared unset entry."},
	}
	data, pending, err := PrepareEnvironment(path, requirements)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), original+"\n") || !strings.HasSuffix(string(data), "# First line.\n# Second line.\nNEW=\n") {
		t.Fatalf("unexpected environment content: %s", data)
	}
	if !reflect.DeepEqual(pending, []string{"EMPTY", "NEW", "UNSET"}) {
		t.Fatalf("unexpected empty variables: %v", pending)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	second, _, err := PrepareEnvironment(path, requirements)
	if err != nil || string(second) != string(data) {
		t.Fatalf("repeated preparation changed existing entries: %v", err)
	}
	if err := os.WriteFile(path, []byte("BROKEN='unterminated"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := PrepareEnvironment(path, requirements); err == nil {
		t.Fatal("must not append inside an unterminated value")
	}
}

func TestHTTPSLoader(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/raw":
			fmt.Fprint(w, "services: {host: {image: nginx, network_mode: host}}")
		case "/redirect":
			http.Redirect(w, r, "/raw", http.StatusFound)
		case "/downgrade":
			http.Redirect(w, r, "http://example.com/raw", http.StatusFound)
		case "/oversize":
			fmt.Fprint(w, strings.Repeat("x", maxTemplateBytes+1))
		case "/html":
			fmt.Fprint(w, "<html>Gist web page</html>")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	for _, path := range []string{"/raw", "/redirect"} {
		if _, err := loadRemote(context.Background(), server.URL+path, server.Client()); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
	for _, path := range []string{"/downgrade", "/oversize", "/html", "/missing"} {
		if _, err := loadRemote(context.Background(), server.URL+path, server.Client()); err == nil {
			t.Errorf("accepted %s", path)
		}
	}
	for _, source := range []string{"http://example.com/raw", "./local.yml", "https://user:password@example.com/raw"} {
		if _, err := LoadTemplate(context.Background(), source); err == nil {
			t.Errorf("accepted disallowed source %s", source)
		}
	}
}

func TestWriteStackStagesBothFiles(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, "compose.yml")
	original := []byte("services: {existing: {image: nginx}}\n")
	if err := os.WriteFile(path, original, 0640); err != nil {
		t.Fatal(err)
	}
	// An unusable .env path must not leave the Compose file changed.
	if err := os.Mkdir(filepath.Join(folder, ".env"), 0700); err != nil {
		t.Fatal(err)
	}
	doc := mustTemplate(t, "services: {added: {image: nginx}}")
	if err := WriteStack(path, doc.Compose, []byte("TOKEN=\n")); err == nil {
		t.Fatal("expected environment staging error")
	}
	actual, err := os.ReadFile(path)
	if err != nil || string(actual) != string(original) {
		t.Fatal("Compose file changed despite environment failure")
	}
	matches, _ := filepath.Glob(filepath.Join(folder, ".stacker-*"))
	if len(matches) != 0 {
		t.Fatalf("left temporary files behind: %v", matches)
	}
}
