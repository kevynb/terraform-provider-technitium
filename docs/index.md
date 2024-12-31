---
page_title: "Provider: Technitium DNS"
description: |-
  Manage DNS resource records for domains hosted on Technitium DNS servers
---

# Technitium DNS Management Plugin: `technitium`

The Technitium DNS provider allows for management of individual DNS resource records for domains hosted on Technitium DNS servers. It is designed for fine-grained DNS record management while preserving existing records and tolerating external modifications.

## Configuration

Provider configuration is simple and usually empty if the authentication information is set in the environment variables `TECHNITIUM_API_URL` and `TECHNITIUM_API_TOKEN`. Alternatively, you can specify these directly in the provider block.

Example:

```terraform
provider "technitium" {
  api_url  = "https://your-technitium-server/api"
  api_token = "your-api-token"
}
```

## Schema

### Optional

- `api_url` (String): Technitium API base URL.
- `api_token` (String, Sensitive): Technitium API token for authentication.

## DNS Record Resource: `technitium_record`

DNS entries are managed as instances of the `technitium_record` resource.

The provider supports the following DNS record types:
- **A, AAAA**: Address records for IPv4 and IPv6.
- **CNAME**: Canonical name records.
- **MX**: Mail exchange records.
- **NS**: Name server records.
- **TXT**: Text records.
- **SRV**: Service locator records.
- **PTR**: Pointer records for reverse DNS.
- **NAPTR**: Naming authority pointer records.
- **CAA**: Certification Authority Authorization records.
- **ANAME**: Alias records.
- **URI**: URI records.
- **TLSA**: TLS authentication records.
- **SOA**: Start of authority for a DNS zone.
- **DNAME**: Redirects a subtree to another domain.
- **DS**: Links parent and child zones for DNSSEC.
- **SSHFP**: Stores SSH key fingerprints.
- **SVCB**: Defines parameters for a service.
- **HTTPS**: Optimized SVCB for HTTPS services.
- **FWD**: Forwards queries to another server.
- **APP**: Application-specific DNS entries.

### Example Usage

Create a simple CNAME record:

```terraform
resource "technitium_record" "my_cname" {
  domain = "alias.example.com"
  type   = "CNAME"
  cname  = "target.example.com"
  ttl    = 3600
}
```

Create multiple records using `for_each`:

```terraform
terraform {
  required_providers {
    technitium-dns = {
      source = "registry.terraform.io/your_username/technitium-dns"
    }
  }
}

provider "technitium-dns" {}

locals {
  records = {
    "srv" = {
      type     = "SRV",
      domain   = "_service._tcp.mydomain.com",
      priority = 10,
      weight   = 5,
      port     = 443,
      target   = "target.mydomain.com",
    },
    "txt" = {
      type   = "TXT",
      domain = "verification.mydomain.com",
      text   = "sample-verification-string",
    },
  }
}

resource "technitium_record" "multiple_records" {
  for_each = local.records
  domain   = each.value.domain
  type     = each.value.type
  ttl      = lookup(each.value, "ttl", 3600)
  priority = lookup(each.value, "priority", null)
  weight   = lookup(each.value, "weight", null)
  port     = lookup(each.value, "port", null)
  target   = lookup(each.value, "target", null)
  text     = lookup(each.value, "text", null)
}
```

### Advanced Example

```terraform
resource "technitium_record" "app_record" {
  domain     = "app.example.com"
  type       = "APP"
  ttl        = 3600
  app_name   = "Split Horizon"
  class_path = "SplitHorizon.SimpleAddress"
  record_data = jsonencode({
    "tailscale": ["100.115.205.32", "fd7a:115c:a1e0:ab12:4843:cd96:6273:cd20"],
    "private": ["192.168.88.50"]
  })
}
```

Explore the Technitium DNS provider to simplify and automate your DNS management. Let us know if you encounter any issues or need additional features!
