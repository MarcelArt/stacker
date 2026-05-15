# AGENTS.md

## Project

`stacker` 🍔 — Go CLI that builds Docker Compose configs one service at a time. Built with Cobra + `gopkg.in/yaml.v3`.

## Commands

- `go build -o stacker .` — build binary
- `go run main.go <subcommand>` — run directly
- `go vet ./...` — static analysis
- No tests exist yet

## Architecture

- `cmd/root.go` — root Cobra command, parses compose file and holds global `dockerCompose` state. Flags: `-f` (compose file, default `docker-compose.yml`), `-n` (network name, default `net`).
- `cmd/<service>.go` — one file per service subcommand. Each command mutates the shared `dockerCompose` variable then marshals/writes YAML to disk.
- `internal/models/compose.model.go` — `Compose`, `Network`, `Volume`, `Service`, `Healthcheck` structs with `Set*` helpers that lazily init maps.

Pattern for adding a new service: create `cmd/<service>.go`, define a Cobra command that calls `dockerCompose.SetNetwork`, `SetVolume`, `SetService`, then `yaml.Marshal` + `os.WriteFile`. Register with `rootCmd.AddCommand` in `init()`.

## Known issue

`internal/models/compose.model.go:12` — `Volume.External` tag is `yam` instead of `yaml`, so volume external flag will not serialize correctly.
