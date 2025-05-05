# Organization

The `Organization` object represents a Numerous organization.

### Getting organization info

An App that is deployed to Numerous can obtain basic organization info
using `organization_from_env`.
Organization `slug` property can be changed if an organization is renamed.
So the value is cached to avoid extra API calls.

```py
from numerous.organization import organization_from_env

organization = organization_from_env()
print(organization.id, organization.slug)
```

## API reference

See the [API reference](reference/numerous/organization/index.md) for details.