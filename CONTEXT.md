# Stacker

Stacker assembles Docker Compose configurations from reusable service templates.

## Language

**Service template**:
A reusable definition of one or more services, their supporting networks and volumes, and the environment variables and setup guidance they require.
_Avoid_: Service command, scaffold configuration

**Compose configuration**:
The assembled definition of the user's stack, containing services and their supporting resources.
_Avoid_: Service template

**Environment requirement**:
A variable a service template declares together with guidance for supplying its value in the stack's environment file.

**Built-in template**:
A service template distributed with Stacker and selected by name.

**Remote template**:
A service template retrieved from an HTTPS URL.

**Shared network**:
The stack network selected by the user that templates can reference alongside their own additional networks.

**Service dependency**:
Another service referenced by a service's `depends_on` declaration. A missing dependency is reported as setup guidance and does not prevent adding the service.
