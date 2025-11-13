# Terraform Technitium DNS Provider

The Technitium DNS provider for Terraform enables the management of individual DNS resource records on domains hosted by Technitium DNS servers. It provides granular control over DNS records while preserving existing records and tolerating external modifications.

## Features

- **Granular Record Management**: Manage individual DNS records (`A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, and more) without affecting the rest of the domain configuration.
- **Support for Diverse Record Types**: Covers all common DNS record types, including `A`, `AAAA`, `CNAME`, `MX`, `NS`, `TXT`, `SRV`, `PTR`, `NAPTR`, `CAA`, `ANAME`, `URI`, and `TLSA`. Special types like `FWD` and `APP` are also supported.
- **Preserve Existing Records**: Modifications made outside Terraform are respected; only managed records are updated or destroyed.
- **Incremental Updates**: Safely manage subsets of DNS records without disrupting others in the same zone.

## Getting Started

### Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    technitium = {
      source = "registry.terraform.io/kevynb/technitium"
    }
  }
}
```

### Authentication

The provider uses the `TECHNITIUM_API_URL` and `TECHNITIUM_API_TOKEN` environment variables for authentication.

Alternatively, credentials can be provided directly in the provider block:

```hcl
provider "technitium" {
   url   = "https://your-technitium-server/api"
   token = "your-api-token"
}
```

### Example Usage

```hcl
resource "technitium_dns_record" "example" {
  domain     = "test.example.com"
  type       = "A"
  ttl        = 3600
  ip_address = "192.168.1.1"
}
```

#### Advanced Example

```hcl
resource "technitium_dns_record" "srv_record" {
  zone        = "example.com"
  domain      = "service.example.com"
  type        = "SRV"
  ttl         = 3600
  priority    = 10
  weight      = 5
  port        = 443
  target      = "target.example.com"
}

resource "technitium_dns_record" "app_record" {
  zone       = "example.com"
  domain     = "app.example.com"
  type       = "APP"
  ttl        = 3600
  app_name   = "Split Horizon"
  class_path = "SplitHorizon.SimpleAddress"
  record_data = jsonencode({
    "tailscale": ["100.115.192.11", "fd7a:115c:a1e0:ab12:4843:ac32:9911:da02"],
    "private": ["192.168.1.50"]
  })
}
```

## Supported Record Types

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
- **SOA**: SOA records.
- **DNAME**: DNAME records.
- **DS**: DS records.
- **SSHFP**: SSHFP records.
- **SVCB**: SVCB records.
- **HTTPS**: HTTPS records.
- **FWD**: Technitium custom FWD records.
- **APP**: Technitium custom APP records.

## Differences from Other Providers

Compared to other DNS providers:

1. **Record-Focused Management**:
    - Only individual records are managed, not the entire domain.
    - Updates are incremental and non-destructive.
2. **Preserve Existing Records**:
    - Unmanaged records are untouched.
3. **Destroy Support**:
    - Deletes only the records explicitly managed by Terraform.

## Notes

This project was mostly whipped out overnight because I needed it for some homelab config. The code is far from perfect, I did not update the unit tests yet as I'm not using this in a business critical fashion.

Things to improve:

- [ ] Rewrite unit tests
- [ ] Rework the record resource so that each record type only allows the associated parameters (this would also allow for cleaner code everywhere as mapping operations could be more specific)
- [ ] Add some caching to the `read` mechanism to avoid refetching the list of records every time (overkill for now as I don't manage a lot of them, could be useful for other users)

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m 'Add new feature'`).
4. Push the branch (`git push origin feature/your-feature`).
5. Open a pull request.

## Thanks

This project is based on [terraform-provider-godaddy-dns](https://github.com/veksh/terraform-provider-godaddy-dns). Special thanks to the contributors for their foundational work.
