---
status: accepted
---

# Replace service commands with YAML service templates

Replace the individual service commands with one `stacker add` entrypoint and convert every existing service definition to a YAML service template. Templates may come from this repository or an internet source such as a Gist and may include multiple services, networks, and volumes. This lets contributors add service definitions without adding Go commands, in exchange for maintaining a template format and a shared loader.

Built-in templates are bundled into the binary and selected by name, so they work offline. Remote templates are fetched from HTTPS URLs. Local template paths are outside the agreed scope.

Templates use ordinary Compose YAML with an `x-stacker` extension for descriptions and environment requirements. The format must remain compatible with Docker Compose. Templates explicitly reference the shared network selected through `-n` or `stacker.toml`, and may also define additional networks.

Adding a template explains its environment requirements and appends missing variables with empty values and descriptive comments to `.env`, without prompting for values or replacing existing entries. The user supplies values afterward. The environment file lives beside the target Compose file, including when `-f` points into another directory. Migrated templates standardize on this file; Infisical's `.env.infisical` reference changes to `.env`.

Existing service names cause an error unless `--overwrite` is supplied to replace those services. Identical networks and volumes are reused; conflicting definitions are rejected. Unrelated Compose fields are preserved.

Missing service dependencies remain advisory, matching existing behavior: add the service and explain which dependencies the user needs to add. Do not automatically fetch or add dependencies. A template may explicitly bundle related services.

Update `rm`, `ls`, and `config` to work with the richer Compose YAML supported by `add`; in particular, removing a service must preserve unrelated Compose fields. Preserve the existing shared network and external-network configuration behavior.

The complete design was confirmed in the design interview, with the addition that templates must support `network_mode: host`. Services using a network mode retain it and receive no shared-network attachment. Reject `network_mode` combined with `networks`, and host mode combined with published ports, before writing files.

## Template contract

`x-stacker` is Compose extension metadata: it declares the template description, the network key to bind to the user's shared network, environment variable descriptions, and optional setup instructions. Docker Compose ignores `x-` extension fields, as documented in the [Compose extension reference](https://docs.docker.com/reference/compose-file/extension/). The loader consumes this metadata without adding it to the destination Compose file. Template instructions are printed, never executed.

```yaml
x-stacker:
  description: Cloudflare Tunnel
  shared_network: shared
  environment:
    TUNNEL_TOKEN:
      description: Cloudflare Tunnel token; fill this in before starting the service.

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

For `stacker -n mynet add cloudflared`, bind the declared `shared` network key and its service references to `mynet`, honoring the configured external-network setting. Preserve additional template networks. Append `TUNNEL_TOKEN=` and its description only when the variable is absent from the destination `.env`; preserve existing values and report empty entries that still need user input.

Validate templates and resource conflicts before changing files. Reject unreadable or malformed existing Compose files rather than replacing their content. Preserve arbitrary valid Compose fields through the add/remove workflow; exact formatting and comment preservation are outside the contract. See the [authoring guide](../service-templates.md) for the implemented format.
