# 🍔 stacker

**Stack up your Docker Compose infrastructure one service at a time.**

`stacker` is a Go CLI that builds Docker Compose configs by adding preconfigured services on demand. Instead of maintaining a monolithic compose file, you pick only the services you need and `stacker` generates the YAML with the correct networks, volumes, environment variables, and dependencies.

## Install

```bash
go install github.com/MarcelArt/stacker@latest
```

Or build from source:

```bash
git clone https://github.com/MarcelArt/stacker.git
cd stacker
go build -o stacker .
```

## Usage

```bash
stacker [command] [flags]
```

### Global Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--file` | `-f` | `docker-compose.yml` | Target compose file |
| `--network` | `-n` | `net` | Network name shared by all services |

### Available Services

| Command | Service | Image | Requires |
|---|---|---|---|
| `postgres` | PostgreSQL 17 | `postgres:17` | — |
| `cloudflared` | Cloudflare Tunnel | `cloudflare/cloudflared:latest` | — |
| `grafana` | Grafana OTel LGTM | `grafana/otel-lgtm:latest` | — |
| `portainer` | Portainer CE | `portainer/portainer-ce:latest` | — |
| `ollama` | Ollama LLM | `ollama/ollama:latest` | — |
| `n8n` | n8n Workflow Automation | `docker.n8n.io/n8nio/n8n` | `postgres` |
| `keycloak` | Keycloak SSO | `quay.io/keycloak/keycloak:latest` | `postgres` |
| `nakama` | Nakama Game Server | `registry.heroiclabs.com/heroiclabs/nakama:3.22.0` | `postgres` |
| `open-web-ui` | Open WebUI | `ghcr.io/open-webui/open-webui:main` | `ollama` |

> 🍔 **Tip:** Run services with dependencies first (e.g. `postgres` before `n8n`). Each command adds one tasty layer to your stack.

### Examples

Build a stack with PostgreSQL and n8n — layer by layer: 🍔

```bash
stacker postgres
stacker n8n
```

Use a custom compose file and network:

```bash
stacker -f compose.yml -n mynet postgres
stacker -f compose.yml -n mynet keycloak
```

### Environment Variables

Services that require secrets read from a `.env` file. After adding a service, `stacker` prints the required variables. Common ones:

```env
# PostgreSQL
POSTGRES_USER=<your-pg-user>
POSTGRES_PASSWORD=<your-pg-password>

# Keycloak
KEYCLOAK_ADMIN=<your-keycloak-admin>
KEYCLOAK_ADMIN_PASSWORD=<your-keycloak-admin-password>
KC_HOSTNAME=<your-keycloak-hostname>

# Cloudflare Tunnel
TUNNEL_TOKEN=<your-tunnel-token>
```

## How It Works

Each subcommand reads the existing compose file, adds its service definition (along with any required networks and volumes), and writes the updated YAML back to disk. Services are additive — stack your burger one layer at a time. 🍔

## Adding a New Service

1. Create `cmd/<service>.go`
2. Define a Cobra command that calls `dockerCompose.SetNetwork`, `dockerCompose.SetVolume`, and `dockerCompose.SetService`
3. Marshal with `yaml.Marshal` and write with `os.WriteFile`
4. Register with `rootCmd.AddCommand` in `init()`

## License

See [LICENSE](LICENSE).
