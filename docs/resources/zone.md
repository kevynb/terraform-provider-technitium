---
# technitium_zone

Manages a DNS zone in Technitium DNS Server.

## Example Usage

```hcl
resource "technitium_zone" "example" {
  name = "example.com"
  type = "Primary"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain name for the DNS zone.
* `type` - (Required) The type of zone to create. Valid values are `Primary`, `Secondary`, `Stub`, `Forwarder`, `SecondaryForwarder`, `Catalog`, `SecondaryCatalog`.
* `catalog` - (Optional) The name of the catalog zone to become its member zone. Valid only for `Primary`, `Stub`, and `Forwarder` zones.
* `use_soa_serial_date_scheme` - (Optional) Set to `true` to enable using date scheme for SOA serial. Valid only with `Primary`, `Forwarder`, and `Catalog` zones.
* `primary_name_server_addresses` - (Optional) List of comma separated IP addresses or domain names of the primary name server. Required for `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones.
* `zone_transfer_protocol` - (Optional) The zone transfer protocol to be used by `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones. Valid values are `Tcp`, `Tls`, `Quic`.
* `tsig_key_name` - (Optional) The TSIG key name to be used by `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones.
* `validate_zone` - (Optional) Set to `true` to enable ZONEMD validation. Valid only for `Secondary` zones.
* `initialize_forwarder` - (Optional) Set to `true` to initialize the Conditional Forwarder zone with an FWD record. Valid for Conditional Forwarder zones.
* `protocol` - (Optional) The DNS transport protocol to be used by the Conditional Forwarder zone. Valid values are `Udp`, `Tcp`, `Tls`, `Https`, `Quic`.
* `forwarder` - (Optional) The address of the DNS server to be used as a forwarder. Required for Conditional Forwarder zones.
* `dnssec_validation` - (Optional) Set to `true` to enable DNSSEC validation. Valid for Conditional Forwarder zones.
* `proxy_type` - (Optional) The type of proxy to be used for conditional forwarding. Valid values are `NoProxy`, `DefaultProxy`, `Http`, `Socks5`.
* `proxy_address` - (Optional) The proxy server address.
* `proxy_port` - (Optional) The proxy server port.
* `proxy_username` - (Optional) The proxy server username.
* `proxy_password` - (Optional) The proxy server password.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the zone (same as `name`).

## Import

Zones can be imported using the zone name:

```shell
terraform import technitium_zone.example example.com
