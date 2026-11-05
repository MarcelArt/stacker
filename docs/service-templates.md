# Service templates

A service template is ordinary Docker Compose YAML with optional `x-stacker`
metadata. Built-ins live in `templates/*.yml` and are embedded when Stacker is
built. Remote templates must be available as raw YAML over HTTPS.

## Example

```yaml
x-stacker:
  description: Cloudflare Tunnel
  shared_network: shared
  environment:
    TUNNEL_TOKEN:
      description: Cloudflare Tunnel token; fill this in before starting.
  instructions:
    - Configure the tunnel in Cloudflare before starting this service.

services:
  cloudflared:
    image: cloudflare/cloudflared:latest
    container_name: cloudflared
    command: tunnel run
    restart: unless-stopped
    environment:
      TUNNEL_TOKEN: ${TUNNEL_TOKEN}
    networks:
      - shared

networks:
  shared:
    driver: bridge
```

Docker Compose explicitly ignores `x-` extension fields, so metadata does not
invalidate a template. See the [Docker Compose extension reference](https://docs.docker.com/reference/compose-file/extension/).
Stacker consumes `x-stacker` without copying it into the destination file.

## Metadata

| Field | Meaning |
| --- | --- |
| `description` | Optional description printed after adding the template. |
| `shared_network` | Optional key under `networks` to bind to the selected shared network. |
| `environment` | Map of variable names to objects containing a nonempty `description`. |
| `instructions` | Optional list of setup instructions printed after adding services. |

Every Compose interpolation variable in the template must have an environment
description, including variables in commands, entrypoints, and volume paths.
Both `${VARIABLE}` and `$VARIABLE` are recognized. Use `$$VARIABLE` when the
container's shell should expand the variable instead of Compose. Variables used
only through `env_file` must also be declared in metadata if users need to fill
them in. Stacker cannot discover inputs hidden inside a container image.

Descriptions become comments in `.env`, followed by `VARIABLE=`. Values are
always empty for newly appended entries. Existing assignments, including empty,
exported, quoted, and multiline values, remain unchanged. No defaults or secret
values are stored in template metadata. All migrated templates use `.env` beside
the target Compose file.

Metadata keys are checked for typos. A template must contain at least one service
and exactly one YAML document. Downloads have a 20-second timeout and a 1 MiB
size limit; redirects must remain on HTTPS. Instructions are text, never scripts
executed by Stacker.

## Shared networks

In the example, `stacker -n mynet add cloudflared` renames the `shared` network key
and its service references to `mynet`. Both list and mapping forms of service
networks work, including mapping attributes such as aliases. Additional networks
retain their definitions. Omit `shared_network` when no network binding is needed.

The saved external-network setting applies to the shared network only. External
shared networks contain `name` and `external: true`; driver and IPAM settings are
omitted because Compose expects the network to exist already. Volumes retain
their explicit template settings independently of the shared network.

## Host networking

```yaml
x-stacker:
  description: A service using the host network

services:
  host-service:
    image: alpine:3
    command: ["sleep", "infinity"]
    network_mode: host
    restart: unless-stopped
```

A host-network service receives no shared-network attachment. Templates can mix
host-network services and services using the shared network. If no services use
the declared shared network, Stacker omits that unused shared network from the
addition.

Compose rejects `network_mode` combined with `networks`. Host networking also
cannot use published `ports`; Stacker reports these combinations before writing
files. Other Compose network modes are preserved too. See the [Compose service
reference](https://docs.docker.com/reference/compose-file/services/#network_mode).

## Bundles and conflicts

Define multiple services under `services` to create a bundle. Standard
`depends_on` declarations supply dependency guidance. Missing dependencies are
reported after adding services; Stacker never downloads another template
automatically.

An existing service requires `--overwrite`, even if its definition is identical.
Overwriting replaces the entire service. Networks, volumes, configs, and secrets
must be identical to existing definitions; `--overwrite` does not bypass resource
conflicts. Empty resource definitions and omitted default `external: false`
are treated as equivalent. Other conflicting
top-level fields are rejected as well.

Unrelated Compose fields survive additions and removals. This includes extension
fields and valid alternative shapes such as array environments, mapped
`depends_on`, long-form ports, and `network_mode`. Stacker validates YAML,
metadata, resource conflicts, and network-mode combinations; Docker Compose
remains the validator for the complete Compose schema and runtime requirements.
YAML formatting, comments, and anchor notation may change when rewritten.

For a complete stack, check the resulting file before starting containers:

```bash
docker compose -f compose.yml config --quiet
```

A partial stack with missing dependencies may fail Compose consistency checks
until those dependencies have been added.
