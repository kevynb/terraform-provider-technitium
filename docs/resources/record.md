---
page_title: "technitium_record Resource - terraform-provider-technitium"
subcategory: ""
description: |-
  Manage DNS resource records on Technitium DNS servers
---

# technitium_record (Resource)

The `technitium_record` resource allows you to manage individual DNS resource records on domains hosted by Technitium DNS servers.

## Example Usage

### Create a CNAME Record

```terraform
resource "technitium_record" "my_cname" {
  zone   = "example.com"
  domain = "alias.example.com"
  type   = "CNAME"
  cname  = "target.example.com"
  ttl    = 3600
}
```

### Create an APP Record

```terraform
resource "technitium_record" "app_record" {
  zone       = "example.com"
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

## Schema

### Required

- **`type`** (String): DNS record type. Supported values:
    - `A`, `AAAA`: Address records for IPv4 and IPv6.
    - `CNAME`: Canonical name records.
    - `MX`: Mail exchange records.
    - `NS`: Name server records.
    - `TXT`: Text records.
    - `SRV`: Service locator records.
    - `PTR`: Pointer records for reverse DNS.
    - `NAPTR`: Naming authority pointer records.
    - `CAA`: Certification Authority Authorization records.
    - `ANAME`: Alias records.
    - `URI`: URI records.
    - `TLSA`: TLS authentication records.
    - `SOA`, `DNAME`, `DS`, `SSHFP`, `SVCB`, `HTTPS`, `FWD`, `APP`: Advanced and custom records.

- **`domain`** (String): Fully qualified domain name (FQDN) for the record (e.g., `sub.example.com`).

- **`ttl`** (Number): Record time-to-live in seconds. Must be between 600 and 604800 (1 week). Defaults to `3600`.

### Optional, based on record type.

- **`preference`** (Number): Priority for `MX` records. Lower values indicate higher priority.
- **`priority`** (Number): Priority for `SRV` records. Lower values indicate higher priority.
- **`weight`** (Number): Weight for `SRV` records to influence load balancing.
- **`port`** (Number): Port number for `SRV` records.
- **`target`** (String): Target for `CNAME`, `SRV`, or similar records.
- **`text`** (String): Text content for `TXT` records.
- **`app_name`** (String): Application name for `APP` records.
- **`class_path`** (String): Class path for `APP` records.
- **`record_data`** (String): JSON-encoded data for `APP` records.
- **`ptr`** (Boolean): Create a PTR record for `A` or `AAAA` records (default: `false`).
- **`create_ptr_zone`** (Boolean): Automatically create a PTR zone for `A` or `AAAA` records (default: `false`).
- **`update_svcb_hints`** (Boolean): Update SVCB hints for `A` or `AAAA` records (default: `false`).
- **`exchange`** (String): Exchange server for `MX` records.
- **`cname`** (String): Canonical name for `CNAME` records.
- **`ip_address`** (String): IPv4 or IPv6 address for `A` or `AAAA` records.

Additional optional fields may be supported based on record type. Their name should mirror what you see in the technitium UI for each field.

### Examples for Specific Record Types

1. **MX Record**:
   ```terraform
   resource "technitium_record" "mx_record" {
     domain    = "mail.example.com"
     type      = "MX"
     ttl       = 3600
     preference  = 10
     exchange  = "mailserver.example.com"
   }
   ```

2. **SRV Record**:
   ```terraform
   resource "technitium_record" "srv_record" {
     domain    = "_sip._tcp.example.com"
     type      = "SRV"
     ttl       = 3600
     priority  = 10
     weight    = 5
     port      = 5060
     target    = "sipserver.example.com"
   }
   ```

3. **TXT Record**:
   ```terraform
   resource "technitium_record" "txt_record" {
     domain = "verification.example.com"
     type   = "TXT"
     ttl    = 3600
     text   = "sample-verification-code"
   }
   ```

## Import

Import is currently not supported. This will need to be added. PRs welcome.
