# 🍔 stacker

Build Docker Compose stacks from reusable YAML service templates.

## Install

```bash
go install github.com/MarcelArt/stacker@latest
```

Or build from source:

```bash
go build -o stacker .
```

## Add services

```bash
stacker add postgres
stacker add n8n
stacker -f ./deploy/compose.yml -n mynet add cloudflared
```

The 17 bundled YAML templates are: `cloudflared`, `gitea`,
`grafana`, `infisical`, `jenkins`, `keycloak`, `kimeru`, `n8n`, `nakama`, `nats`, `ollama`,
`open-web-ui`, `portainer`, `postgres`, `redis`, `stirling`, and `surrealdb`.
They work offline. Use `stacker add --help` to see the available names.

The old service commands have been replaced: use `stacker add postgres` in place
of `stacker postgres`.

You can also add a template from an HTTPS URL, including a Gist's **Raw** URL:

```bash
stacker add 'https://gist.githubusercontent.com/USER/ID/raw/REVISION/template.yml'
```

Use the URL to the YAML content rather than the Gist web page. Local template
paths and HTTP URLs are unsupported. A template can contain several services and
their networks, volumes, configs, and secrets. Stacker fetches the template and
writes configuration; setup instructions are printed for the user to follow.

## Environment setup

Templates describe the variables their services need. Stacker prints these
descriptions and appends missing entries with empty values to `.env` beside the
Compose file:

```env
# Cloudflare Tunnel token.
TUNNEL_TOKEN=
```

Fill in the values before starting the service. Existing entries and values are
preserved, including when overwriting a service. Stacker does not substitute
secret values into the Compose file or prompt for them.

With `-f ./deploy/compose.yml`, the environment file is `./deploy/.env`. Infisical
and Redis also use that file through `env_file: .env`.

Missing service dependencies remain advisory. For example, adding n8n without
PostgreSQL succeeds and tells you to run `stacker add postgres`. Add the missing
dependencies before starting the stack.

## Existing configurations

```bash
stacker add cloudflared --overwrite
stacker ls
stacker rm -c cloudflared
stacker config
```

An existing service name causes an error unless `--overwrite` is supplied. That
flag replaces the whole service definition. Identical networks, volumes, configs,
and secrets are reused; conflicting resource definitions cause an error even
with `--overwrite`. Invalid YAML or resource conflicts are reported before either
output file changes.

`add` and `rm` preserve unrelated Compose attributes, including `network_mode`,
long-form ports, mapped networks and dependencies, and custom extensions. YAML
formatting and comments may change when the file is rewritten.

## Networks

`-n` selects the shared network, defaulting to `net`. Templates explicitly mark
which network is shared; their additional networks retain their own definitions.

`stacker config` saves network defaults to `stacker.toml` in the current working
directory. Use `stacker -n mynet config` to select a particular network when there
are several. Subsequent additions honor the saved network name and external
setting; `-n` overrides the name. External shared networks are referenced with
`external: true` and must already exist before starting containers.

Templates can use `network_mode: host`. Stacker preserves that mode and does not
attach the service to the shared network. A service cannot combine `network_mode`
with `networks`, or host mode with published `ports`.

## Create a template

Add a YAML file under [`templates/`](templates/) and rebuild Stacker, or publish it
at an HTTPS URL. Service definitions use standard Compose YAML; `x-stacker`
contains descriptions, shared-network metadata, and setup instructions.

See the [template authoring guide](docs/service-templates.md) for the format and a
host-network example. No per-service Go command is needed.

## Development

```bash
go test ./...
go vet ./...
go build -o stacker .
```

## License

See [LICENSE](LICENSE).
