# AGENTS.md

`stacker` 🍔 builds Docker Compose stacks from bundled or HTTPS YAML service templates.

## Validation

- `go test ./...` — CLI, template, merge, environment, and HTTPS regression tests
- `go vet ./...` — static analysis
- `go build -o stacker .` — build binary

## Architecture and conventions

- `cmd/` constructs Cobra commands around per-invocation options; `add` is the service entrypoint.
- `internal/scaffold/` loads templates, preserves generic Compose mappings, checks conflicts, and stages Compose/environment writes.
- `templates/*.yml` are embedded built-ins. To add a service or bundle, follow [docs/service-templates.md](docs/service-templates.md); keep service definitions in YAML.
- `internal/models/config.model.go` represents network defaults in `stacker.toml`.
- [CONTEXT.md](CONTEXT.md) defines domain terms; [docs/adr/](docs/adr/) records architectural decisions.

Preserve arbitrary Compose fields in add/remove operations. Keep missing dependencies advisory, existing environment entries intact, and new `.env` values empty. Preserve `network_mode` services without attaching a shared network; reject incompatible network and port settings before writing files.
