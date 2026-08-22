package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                  = &SettingsResource{}
	_ resource.ResourceWithConfigure     = &SettingsResource{}
	_ datasource.DataSource              = &SettingsDataSource{}
	_ datasource.DataSourceWithConfigure = &SettingsDataSource{}
)

// SettingsResource defines the implementation of Technitium DNS server settings
type SettingsResource struct {
	client   model.DNSApiClient
	reqMutex *sync.Mutex
}

func SettingsResourceFactory(m *sync.Mutex) func() resource.Resource {
	return func() resource.Resource {
		return &SettingsResource{reqMutex: m}
	}
}

func (r *SettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
}

func settingsSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			MarkdownDescription: "Identifier for the settings (always `settings`).",
			Computed:            true,
		},

		// General
		"dns_server_domain": rschema.StringAttribute{
			MarkdownDescription: "The primary domain name used by this DNS Server to identify itself.",
			Optional:            true,
			Computed:            true,
		},
		"dns_server_local_end_points": rschema.ListAttribute{
			MarkdownDescription: "Local end points (IP:port) the DNS Server listens on.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_server_ipv4_source_addresses": rschema.ListAttribute{
			MarkdownDescription: "IPv4 source addresses for outbound DNS requests.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_server_ipv6_source_addresses": rschema.ListAttribute{
			MarkdownDescription: "IPv6 source addresses for outbound DNS requests.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"default_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for records.",
			Optional:            true,
			Computed:            true,
		},
		"default_ns_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for NS records.",
			Optional:            true,
			Computed:            true,
		},
		"default_soa_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for SOA records.",
			Optional:            true,
			Computed:            true,
		},
		"default_responsible_person": rschema.StringAttribute{
			MarkdownDescription: "The default SOA Responsible Person email address.",
			Optional:            true,
			Computed:            true,
		},
		"use_soa_serial_date_scheme": rschema.BoolAttribute{
			MarkdownDescription: "Use date scheme for SOA serial.",
			Optional:            true,
			Computed:            true,
		},
		"min_soa_refresh": rschema.Int64Attribute{
			MarkdownDescription: "Minimum SOA Refresh interval in seconds.",
			Optional:            true,
			Computed:            true,
		},
		"min_soa_retry": rschema.Int64Attribute{
			MarkdownDescription: "Minimum SOA Retry interval in seconds.",
			Optional:            true,
			Computed:            true,
		},
		"zone_transfer_allowed_networks": rschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to perform zone transfer.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"notify_allowed_networks": rschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to send notify for secondary zones.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_apps_enable_automatic_update": rschema.BoolAttribute{
			MarkdownDescription: "Allow DNS server to automatically update DNS Apps.",
			Optional:            true,
			Computed:            true,
		},

		// Network
		"prefer_ipv6": rschema.BoolAttribute{
			MarkdownDescription: "Use IPv6 for querying whenever possible.",
			Optional:            true,
			Computed:            true,
		},
		"enable_udp_socket_pool": rschema.BoolAttribute{
			MarkdownDescription: "Enable UDP socket pool for outbound requests.",
			Optional:            true,
			Computed:            true,
		},
		"udp_payload_size": rschema.Int64Attribute{
			MarkdownDescription: "Maximum EDNS UDP payload size (512-4096).",
			Optional:            true,
			Computed:            true,
		},

		// DNSSEC
		"dnssec_validation": rschema.BoolAttribute{
			MarkdownDescription: "Enable DNSSEC validation.",
			Optional:            true,
			Computed:            true,
		},
		"edns_client_subnet": rschema.BoolAttribute{
			MarkdownDescription: "Enable EDNS Client Subnet.",
			Optional:            true,
			Computed:            true,
		},
		"edns_client_subnet_ipv4_prefix_length": rschema.Int64Attribute{
			MarkdownDescription: "EDNS Client Subnet IPv4 prefix length.",
			Optional:            true,
			Computed:            true,
		},
		"edns_client_subnet_ipv6_prefix_length": rschema.Int64Attribute{
			MarkdownDescription: "EDNS Client Subnet IPv6 prefix length.",
			Optional:            true,
			Computed:            true,
		},
		"edns_client_subnet_ipv4_override": rschema.StringAttribute{
			MarkdownDescription: "IPv4 network address override for ECS.",
			Optional:            true,
			Computed:            true,
		},
		"edns_client_subnet_ipv6_override": rschema.StringAttribute{
			MarkdownDescription: "IPv6 network address override for ECS.",
			Optional:            true,
			Computed:            true,
		},

		// QPM Limits
		"qpm_limit_sample_minutes": rschema.Int64Attribute{
			MarkdownDescription: "Client query stats sample size in minutes for QPM limit.",
			Optional:            true,
			Computed:            true,
		},
		"qpm_limit_udp_truncation_percentage": rschema.Int64Attribute{
			MarkdownDescription: "Percentage of requests responded with TC when QPM limit exceeds (0-100).",
			Optional:            true,
			Computed:            true,
		},
		"qpm_limit_bypass_list": rschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to bypass QPM limit.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},

		// Timeouts
		"client_timeout": rschema.Int64Attribute{
			MarkdownDescription: "Timeout in milliseconds before responding ServerFailure (1000-10000).",
			Optional:            true,
			Computed:            true,
		},
		"tcp_send_timeout": rschema.Int64Attribute{
			MarkdownDescription: "Max TCP send timeout in milliseconds (1000-90000).",
			Optional:            true,
			Computed:            true,
		},
		"tcp_receive_timeout": rschema.Int64Attribute{
			MarkdownDescription: "Max TCP receive timeout in milliseconds (1000-90000).",
			Optional:            true,
			Computed:            true,
		},
		"quic_idle_timeout": rschema.Int64Attribute{
			MarkdownDescription: "QUIC idle connection timeout in milliseconds (1000-90000).",
			Optional:            true,
			Computed:            true,
		},
		"quic_max_inbound_streams": rschema.Int64Attribute{
			MarkdownDescription: "Max inbound bidirectional streams per QUIC connection (1-1000).",
			Optional:            true,
			Computed:            true,
		},
		"listen_backlog": rschema.Int64Attribute{
			MarkdownDescription: "Maximum pending inbound connections.",
			Optional:            true,
			Computed:            true,
		},
		"max_concurrent_resolutions_per_core": rschema.Int64Attribute{
			MarkdownDescription: "Max concurrent async outbound resolutions per CPU core.",
			Optional:            true,
			Computed:            true,
		},

		// Web Service
		"web_service_local_addresses": rschema.ListAttribute{
			MarkdownDescription: "Network interface IP addresses the web service listens on.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"web_service_http_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for the web console HTTP.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_enable_tls": rschema.BoolAttribute{
			MarkdownDescription: "Enable HTTPS for the web service.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_enable_http3": rschema.BoolAttribute{
			MarkdownDescription: "Enable HTTP/3 for the web service.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_http_to_tls_redirect": rschema.BoolAttribute{
			MarkdownDescription: "Enable HTTP to HTTPS redirection.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_use_self_signed_tls_certificate": rschema.BoolAttribute{
			MarkdownDescription: "Use self-signed TLS certificate.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_tls_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for the web console HTTPS.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_tls_certificate_path": rschema.StringAttribute{
			MarkdownDescription: "Path to PKCS #12 certificate (.pfx) for web HTTPS.",
			Optional:            true,
			Computed:            true,
		},
		"web_service_tls_certificate_password": rschema.StringAttribute{
			MarkdownDescription: "Password for the web service TLS certificate.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
		},
		"web_service_real_ip_header": rschema.StringAttribute{
			MarkdownDescription: "HTTP header for reading client IP behind a reverse proxy.",
			Optional:            true,
			Computed:            true,
		},

		// DNS-over-X protocols
		"enable_dns_over_udp_proxy": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-UDP-PROXY requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_tcp_proxy": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-TCP-PROXY requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_http": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-HTTP requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_tls": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-TLS requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_https": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-HTTPS requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_http3": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-HTTP/3 requests.",
			Optional:            true,
			Computed:            true,
		},
		"enable_dns_over_quic": rschema.BoolAttribute{
			MarkdownDescription: "Accept DNS-over-QUIC requests.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_udp_proxy_port": rschema.Int64Attribute{
			MarkdownDescription: "UDP port for DNS-over-UDP-PROXY.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_tcp_proxy_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for DNS-over-TCP-PROXY.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_http_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for DNS-over-HTTP.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_tls_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for DNS-over-TLS.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_https_port": rschema.Int64Attribute{
			MarkdownDescription: "TCP port for DNS-over-HTTPS.",
			Optional:            true,
			Computed:            true,
		},
		"dns_over_quic_port": rschema.Int64Attribute{
			MarkdownDescription: "UDP port for DNS-over-QUIC.",
			Optional:            true,
			Computed:            true,
		},

		// Reverse Proxy & TLS for DNS
		"reverse_proxy_network_acl": rschema.ListAttribute{
			MarkdownDescription: "ACL for reverse proxy access for DNS-over-UDP-PROXY, DNS-over-TCP-PROXY, and DNS-over-HTTP protocols.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_tls_certificate_path": rschema.StringAttribute{
			MarkdownDescription: "Path to PKCS #12 certificate (.pfx) for DNS-over-TLS/HTTPS.",
			Optional:            true,
			Computed:            true,
		},
		"dns_tls_certificate_password": rschema.StringAttribute{
			MarkdownDescription: "Password for the DNS TLS certificate.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
		},
		"dns_over_http_real_ip_header": rschema.StringAttribute{
			MarkdownDescription: "HTTP header for reading client IP for DNS-over-HTTP behind a reverse proxy.",
			Optional:            true,
			Computed:            true,
		},

		// Recursion
		"recursion": rschema.StringAttribute{
			MarkdownDescription: "Recursion policy. Valid values: `Deny`, `Allow`, `AllowOnlyForPrivateNetworks`, `UseSpecifiedNetworkACL`.",
			Optional:            true,
			Computed:            true,
		},
		"recursion_network_acl": rschema.ListAttribute{
			MarkdownDescription: "Access Control List for recursion (used when recursion is `UseSpecifiedNetworkACL`).",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"randomize_name": rschema.BoolAttribute{
			MarkdownDescription: "Enable QNAME randomization (0x20).",
			Optional:            true,
			Computed:            true,
		},
		"qname_minimization": rschema.BoolAttribute{
			MarkdownDescription: "Enable QNAME minimization.",
			Optional:            true,
			Computed:            true,
		},
		"resolver_retries": rschema.Int64Attribute{
			MarkdownDescription: "Number of retries for the recursive resolver.",
			Optional:            true,
			Computed:            true,
		},
		"resolver_timeout": rschema.Int64Attribute{
			MarkdownDescription: "Timeout in milliseconds for the recursive resolver.",
			Optional:            true,
			Computed:            true,
		},
		"resolver_concurrency": rschema.Int64Attribute{
			MarkdownDescription: "Number of concurrent requests for the recursive resolver.",
			Optional:            true,
			Computed:            true,
		},
		"resolver_max_stack_count": rschema.Int64Attribute{
			MarkdownDescription: "Max stack count for the recursive resolver.",
			Optional:            true,
			Computed:            true,
		},

		// Cache
		"save_cache": rschema.BoolAttribute{
			MarkdownDescription: "Save DNS cache on disk when the server stops.",
			Optional:            true,
			Computed:            true,
		},
		"serve_stale": rschema.BoolAttribute{
			MarkdownDescription: "Enable serve stale feature.",
			Optional:            true,
			Computed:            true,
		},
		"serve_stale_ttl": rschema.Int64Attribute{
			MarkdownDescription: "TTL in seconds for expired cached records (max 7 days).",
			Optional:            true,
			Computed:            true,
		},
		"serve_stale_answer_ttl": rschema.Int64Attribute{
			MarkdownDescription: "TTL in seconds for stale response records (0-300).",
			Optional:            true,
			Computed:            true,
		},
		"serve_stale_reset_ttl": rschema.Int64Attribute{
			MarkdownDescription: "TTL in seconds to reset stale record TTL (10-900).",
			Optional:            true,
			Computed:            true,
		},
		"serve_stale_max_wait_time": rschema.Int64Attribute{
			MarkdownDescription: "Max wait time in milliseconds before serving stale (0-1800).",
			Optional:            true,
			Computed:            true,
		},
		"cache_maximum_entries": rschema.Int64Attribute{
			MarkdownDescription: "Maximum entries in cache.",
			Optional:            true,
			Computed:            true,
		},
		"cache_minimum_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "Minimum TTL for cached records.",
			Optional:            true,
			Computed:            true,
		},
		"cache_maximum_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "Maximum TTL for cached records.",
			Optional:            true,
			Computed:            true,
		},
		"cache_negative_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "Negative TTL for cached records.",
			Optional:            true,
			Computed:            true,
		},
		"cache_failure_record_ttl": rschema.Int64Attribute{
			MarkdownDescription: "Failure TTL for cached records.",
			Optional:            true,
			Computed:            true,
		},
		"cache_prefetch_eligibility": rschema.Int64Attribute{
			MarkdownDescription: "Minimum initial TTL for prefetch eligibility.",
			Optional:            true,
			Computed:            true,
		},
		"cache_prefetch_trigger": rschema.Int64Attribute{
			MarkdownDescription: "TTL trigger value to initiate prefetch. Set 0 to disable.",
			Optional:            true,
			Computed:            true,
		},
		"cache_prefetch_sample_interval_in_minutes": rschema.Int64Attribute{
			MarkdownDescription: "Sample interval in minutes for auto prefetch.",
			Optional:            true,
			Computed:            true,
		},
		"cache_prefetch_sample_eligibility_hits_per_hour": rschema.Int64Attribute{
			MarkdownDescription: "Minimum hits per hour for auto prefetch eligibility.",
			Optional:            true,
			Computed:            true,
		},

		// Blocking
		"enable_blocking": rschema.BoolAttribute{
			MarkdownDescription: "Enable domain name blocking.",
			Optional:            true,
			Computed:            true,
		},
		"allow_txt_blocking_report": rschema.BoolAttribute{
			MarkdownDescription: "Respond with TXT records containing blocked domain report.",
			Optional:            true,
			Computed:            true,
		},
		"blocking_bypass_list": rschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to bypass blocking.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"blocking_type": rschema.StringAttribute{
			MarkdownDescription: "Blocking response type. Valid values: `AnyAddress`, `NxDomain`, `CustomAddress`.",
			Optional:            true,
			Computed:            true,
		},
		"blocking_answer_ttl": rschema.Int64Attribute{
			MarkdownDescription: "TTL in seconds for blocking responses.",
			Optional:            true,
			Computed:            true,
		},
		"custom_blocking_addresses": rschema.ListAttribute{
			MarkdownDescription: "Custom blocking IP addresses (used when blocking_type is `CustomAddress`).",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"block_list_urls": rschema.ListAttribute{
			MarkdownDescription: "Block list URLs for automatic download.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"block_list_update_interval_hours": rschema.Int64Attribute{
			MarkdownDescription: "Interval in hours for block list auto-update.",
			Optional:            true,
			Computed:            true,
		},

		// Proxy
		"proxy_type": rschema.StringAttribute{
			MarkdownDescription: "Proxy protocol type. Valid values: `None`, `Http`, `Socks5`.",
			Optional:            true,
			Computed:            true,
		},
		"proxy_address": rschema.StringAttribute{
			MarkdownDescription: "Proxy server hostname or IP.",
			Optional:            true,
			Computed:            true,
		},
		"proxy_port": rschema.Int64Attribute{
			MarkdownDescription: "Proxy server port.",
			Optional:            true,
			Computed:            true,
		},
		"proxy_username": rschema.StringAttribute{
			MarkdownDescription: "Proxy server username.",
			Optional:            true,
			Computed:            true,
		},
		"proxy_password": rschema.StringAttribute{
			MarkdownDescription: "Proxy server password.",
			Optional:            true,
			Computed:            true,
			Sensitive:           true,
		},
		"proxy_bypass": rschema.StringAttribute{
			MarkdownDescription: "Comma separated bypass list for proxy.",
			Optional:            true,
			Computed:            true,
		},

		// Forwarders
		"forwarders": rschema.ListAttribute{
			MarkdownDescription: "List of forwarder addresses. Set empty to do recursive resolution.",
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"forwarder_protocol": rschema.StringAttribute{
			MarkdownDescription: "Forwarder DNS transport protocol. Valid values: `Udp`, `Tcp`, `Tls`, `Https`, `Quic`.",
			Optional:            true,
			Computed:            true,
		},
		"concurrent_forwarding": rschema.BoolAttribute{
			MarkdownDescription: "Query multiple forwarders concurrently.",
			Optional:            true,
			Computed:            true,
		},
		"forwarder_retries": rschema.Int64Attribute{
			MarkdownDescription: "Number of retries for forwarder DNS client.",
			Optional:            true,
			Computed:            true,
		},
		"forwarder_timeout": rschema.Int64Attribute{
			MarkdownDescription: "Timeout in milliseconds for forwarder DNS client.",
			Optional:            true,
			Computed:            true,
		},
		"forwarder_concurrency": rschema.Int64Attribute{
			MarkdownDescription: "Number of concurrent requests for forwarder DNS client.",
			Optional:            true,
			Computed:            true,
		},

		// Logging
		"logging_type": rschema.StringAttribute{
			MarkdownDescription: "Logging type. Valid values: `None`, `File`, `Console`, `FileAndConsole`.",
			Optional:            true,
			Computed:            true,
		},
		"ignore_resolver_logs": rschema.BoolAttribute{
			MarkdownDescription: "Stop logging domain name resolution errors.",
			Optional:            true,
			Computed:            true,
		},
		"log_queries": rschema.BoolAttribute{
			MarkdownDescription: "Log every query and response.",
			Optional:            true,
			Computed:            true,
		},
		"use_local_time": rschema.BoolAttribute{
			MarkdownDescription: "Use local time instead of UTC for logging.",
			Optional:            true,
			Computed:            true,
		},
		"log_folder": rschema.StringAttribute{
			MarkdownDescription: "Folder path for log files.",
			Optional:            true,
			Computed:            true,
		},
		"max_log_file_days": rschema.Int64Attribute{
			MarkdownDescription: "Max days to keep log files. Set 0 to disable auto delete.",
			Optional:            true,
			Computed:            true,
		},
		"enable_in_memory_stats": rschema.BoolAttribute{
			MarkdownDescription: "Enable in-memory stats (only Last Hour on Dashboard).",
			Optional:            true,
			Computed:            true,
		},
		"max_stat_file_days": rschema.Int64Attribute{
			MarkdownDescription: "Max days to keep stat files. Set 0 to disable auto delete.",
			Optional:            true,
			Computed:            true,
		},
	}
}

func (r *SettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages the Technitium DNS Server settings. This is a singleton resource — only one instance should exist.",
		Attributes:          settingsSchemaAttributes(),
	}
}

func (r *SettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(model.DNSApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Internal error: expected model.DNSApiClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData tfDNSSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "settings: create (apply)")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	apiSettings := tfSettings2model(planData)

	result, err := r.client.SetSettings(ctx, &apiSettings)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to set DNS settings: %s", err))
		return
	}

	stateData := modelSettings2tf(*result)
	// Preserve sensitive values from plan since the API returns masked passwords
	preserveSensitiveSettings(&stateData, planData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *SettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfDNSSettings
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "settings: read")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	result, err := r.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read DNS settings: %s", err))
		return
	}

	newState := modelSettings2tf(*result)
	// Preserve sensitive values from existing state since the API returns masked passwords
	preserveSensitiveSettings(&newState, stateData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *SettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData tfDNSSettings
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "settings: update")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	apiSettings := tfSettings2model(planData)

	result, err := r.client.SetSettings(ctx, &apiSettings)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to update DNS settings: %s", err))
		return
	}

	stateData := modelSettings2tf(*result)
	preserveSensitiveSettings(&stateData, planData)
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}

func (r *SettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// DNS settings cannot be "deleted" — removing from state is sufficient.
	// The server will keep its current settings.
	tflog.Info(ctx, "settings: delete (removing from state only)")
	resp.Diagnostics.Append(resp.State.RemoveResource(ctx)...)
}

// preserveSensitiveSettings copies password fields from plan/state to the new state
// because the API returns masked values (e.g. "************") for sensitive fields.
func preserveSensitiveSettings(target *tfDNSSettings, source tfDNSSettings) {
	if !source.WebServiceTlsCertificatePassword.IsNull() && !source.WebServiceTlsCertificatePassword.IsUnknown() {
		target.WebServiceTlsCertificatePassword = source.WebServiceTlsCertificatePassword
	}
	if !source.DnsTlsCertificatePassword.IsNull() && !source.DnsTlsCertificatePassword.IsUnknown() {
		target.DnsTlsCertificatePassword = source.DnsTlsCertificatePassword
	}
	if !source.ProxyPassword.IsNull() && !source.ProxyPassword.IsUnknown() {
		target.ProxyPassword = source.ProxyPassword
	}
}

// SettingsDataSource defines the data source implementation for settings
type SettingsDataSource struct {
	client   model.DNSApiClient
	reqMutex *sync.Mutex
}

func SettingsDataSourceFactory(m *sync.Mutex) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &SettingsDataSource{reqMutex: m}
	}
}

func (d *SettingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_settings"
}

func settingsDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			MarkdownDescription: "Identifier for the settings (always `settings`).",
			Computed:            true,
		},

		// General
		"dns_server_domain": dschema.StringAttribute{
			MarkdownDescription: "The primary domain name used by this DNS Server.",
			Computed:            true,
		},
		"dns_server_local_end_points": dschema.ListAttribute{
			MarkdownDescription: "Local end points (IP:port) the DNS Server listens on.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_server_ipv4_source_addresses": dschema.ListAttribute{
			MarkdownDescription: "IPv4 source addresses for outbound DNS requests.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_server_ipv6_source_addresses": dschema.ListAttribute{
			MarkdownDescription: "IPv6 source addresses for outbound DNS requests.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"default_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for records.",
			Computed:            true,
		},
		"default_ns_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for NS records.",
			Computed:            true,
		},
		"default_soa_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "The default TTL value for SOA records.",
			Computed:            true,
		},
		"default_responsible_person": dschema.StringAttribute{
			MarkdownDescription: "The default SOA Responsible Person email address.",
			Computed:            true,
		},
		"use_soa_serial_date_scheme": dschema.BoolAttribute{
			MarkdownDescription: "Use date scheme for SOA serial.",
			Computed:            true,
		},
		"min_soa_refresh": dschema.Int64Attribute{
			MarkdownDescription: "Minimum SOA Refresh interval in seconds.",
			Computed:            true,
		},
		"min_soa_retry": dschema.Int64Attribute{
			MarkdownDescription: "Minimum SOA Retry interval in seconds.",
			Computed:            true,
		},
		"zone_transfer_allowed_networks": dschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to perform zone transfer.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"notify_allowed_networks": dschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to send notify.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_apps_enable_automatic_update": dschema.BoolAttribute{
			MarkdownDescription: "Allow DNS server to automatically update DNS Apps.",
			Computed:            true,
		},

		// Network
		"prefer_ipv6": dschema.BoolAttribute{
			MarkdownDescription: "Use IPv6 for querying whenever possible.",
			Computed:            true,
		},
		"enable_udp_socket_pool": dschema.BoolAttribute{
			MarkdownDescription: "Enable UDP socket pool.",
			Computed:            true,
		},
		"udp_payload_size": dschema.Int64Attribute{
			MarkdownDescription: "Maximum EDNS UDP payload size.",
			Computed:            true,
		},

		// DNSSEC
		"dnssec_validation": dschema.BoolAttribute{
			MarkdownDescription: "Enable DNSSEC validation.",
			Computed:            true,
		},
		"edns_client_subnet": dschema.BoolAttribute{
			MarkdownDescription: "Enable EDNS Client Subnet.",
			Computed:            true,
		},
		"edns_client_subnet_ipv4_prefix_length": dschema.Int64Attribute{
			MarkdownDescription: "EDNS Client Subnet IPv4 prefix length.",
			Computed:            true,
		},
		"edns_client_subnet_ipv6_prefix_length": dschema.Int64Attribute{
			MarkdownDescription: "EDNS Client Subnet IPv6 prefix length.",
			Computed:            true,
		},
		"edns_client_subnet_ipv4_override": dschema.StringAttribute{
			MarkdownDescription: "IPv4 network address override for ECS.",
			Computed:            true,
		},
		"edns_client_subnet_ipv6_override": dschema.StringAttribute{
			MarkdownDescription: "IPv6 network address override for ECS.",
			Computed:            true,
		},

		// QPM Limits
		"qpm_limit_sample_minutes": dschema.Int64Attribute{
			MarkdownDescription: "Client query stats sample size in minutes.",
			Computed:            true,
		},
		"qpm_limit_udp_truncation_percentage": dschema.Int64Attribute{
			MarkdownDescription: "Percentage of TC responses when QPM limit exceeds.",
			Computed:            true,
		},
		"qpm_limit_bypass_list": dschema.ListAttribute{
			MarkdownDescription: "IP/network addresses allowed to bypass QPM limit.",
			Computed:            true,
			ElementType:         types.StringType,
		},

		// Timeouts
		"client_timeout": dschema.Int64Attribute{
			MarkdownDescription: "Timeout in milliseconds before ServerFailure response.",
			Computed:            true,
		},
		"tcp_send_timeout": dschema.Int64Attribute{
			MarkdownDescription: "Max TCP send timeout in milliseconds.",
			Computed:            true,
		},
		"tcp_receive_timeout": dschema.Int64Attribute{
			MarkdownDescription: "Max TCP receive timeout in milliseconds.",
			Computed:            true,
		},
		"quic_idle_timeout": dschema.Int64Attribute{
			MarkdownDescription: "QUIC idle connection timeout in milliseconds.",
			Computed:            true,
		},
		"quic_max_inbound_streams": dschema.Int64Attribute{
			MarkdownDescription: "Max inbound streams per QUIC connection.",
			Computed:            true,
		},
		"listen_backlog": dschema.Int64Attribute{
			MarkdownDescription: "Maximum pending inbound connections.",
			Computed:            true,
		},
		"max_concurrent_resolutions_per_core": dschema.Int64Attribute{
			MarkdownDescription: "Max concurrent resolutions per CPU core.",
			Computed:            true,
		},

		// Web Service
		"web_service_local_addresses": dschema.ListAttribute{
			MarkdownDescription: "Web service listen addresses.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"web_service_http_port": dschema.Int64Attribute{
			MarkdownDescription: "Web console HTTP port.",
			Computed:            true,
		},
		"web_service_enable_tls": dschema.BoolAttribute{
			MarkdownDescription: "HTTPS enabled for web service.",
			Computed:            true,
		},
		"web_service_enable_http3": dschema.BoolAttribute{
			MarkdownDescription: "HTTP/3 enabled for web service.",
			Computed:            true,
		},
		"web_service_http_to_tls_redirect": dschema.BoolAttribute{
			MarkdownDescription: "HTTP to HTTPS redirection enabled.",
			Computed:            true,
		},
		"web_service_use_self_signed_tls_certificate": dschema.BoolAttribute{
			MarkdownDescription: "Self-signed TLS certificate in use.",
			Computed:            true,
		},
		"web_service_tls_port": dschema.Int64Attribute{
			MarkdownDescription: "Web console HTTPS port.",
			Computed:            true,
		},
		"web_service_tls_certificate_path": dschema.StringAttribute{
			MarkdownDescription: "Path to web service TLS certificate.",
			Computed:            true,
		},
		"web_service_real_ip_header": dschema.StringAttribute{
			MarkdownDescription: "HTTP header for client IP behind reverse proxy.",
			Computed:            true,
		},

		// DNS-over-X protocols
		"enable_dns_over_udp_proxy": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-UDP-PROXY enabled.",
			Computed:            true,
		},
		"enable_dns_over_tcp_proxy": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-TCP-PROXY enabled.",
			Computed:            true,
		},
		"enable_dns_over_http": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-HTTP enabled.",
			Computed:            true,
		},
		"enable_dns_over_tls": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-TLS enabled.",
			Computed:            true,
		},
		"enable_dns_over_https": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-HTTPS enabled.",
			Computed:            true,
		},
		"enable_dns_over_http3": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-HTTP/3 enabled.",
			Computed:            true,
		},
		"enable_dns_over_quic": dschema.BoolAttribute{
			MarkdownDescription: "DNS-over-QUIC enabled.",
			Computed:            true,
		},
		"dns_over_udp_proxy_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-UDP-PROXY port.",
			Computed:            true,
		},
		"dns_over_tcp_proxy_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-TCP-PROXY port.",
			Computed:            true,
		},
		"dns_over_http_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-HTTP port.",
			Computed:            true,
		},
		"dns_over_tls_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-TLS port.",
			Computed:            true,
		},
		"dns_over_https_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-HTTPS port.",
			Computed:            true,
		},
		"dns_over_quic_port": dschema.Int64Attribute{
			MarkdownDescription: "DNS-over-QUIC port.",
			Computed:            true,
		},

		// Reverse Proxy & TLS for DNS
		"reverse_proxy_network_acl": dschema.ListAttribute{
			MarkdownDescription: "ACL for reverse proxy access.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"dns_tls_certificate_path": dschema.StringAttribute{
			MarkdownDescription: "Path to DNS TLS certificate.",
			Computed:            true,
		},
		"dns_over_http_real_ip_header": dschema.StringAttribute{
			MarkdownDescription: "HTTP header for DNS-over-HTTP client IP.",
			Computed:            true,
		},

		// Recursion
		"recursion": dschema.StringAttribute{
			MarkdownDescription: "Recursion policy.",
			Computed:            true,
		},
		"recursion_network_acl": dschema.ListAttribute{
			MarkdownDescription: "ACL for recursion.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"randomize_name": dschema.BoolAttribute{
			MarkdownDescription: "QNAME randomization enabled.",
			Computed:            true,
		},
		"qname_minimization": dschema.BoolAttribute{
			MarkdownDescription: "QNAME minimization enabled.",
			Computed:            true,
		},
		"resolver_retries": dschema.Int64Attribute{
			MarkdownDescription: "Recursive resolver retries.",
			Computed:            true,
		},
		"resolver_timeout": dschema.Int64Attribute{
			MarkdownDescription: "Recursive resolver timeout.",
			Computed:            true,
		},
		"resolver_concurrency": dschema.Int64Attribute{
			MarkdownDescription: "Recursive resolver concurrency.",
			Computed:            true,
		},
		"resolver_max_stack_count": dschema.Int64Attribute{
			MarkdownDescription: "Recursive resolver max stack count.",
			Computed:            true,
		},

		// Cache
		"save_cache": dschema.BoolAttribute{
			MarkdownDescription: "Save DNS cache on disk.",
			Computed:            true,
		},
		"serve_stale": dschema.BoolAttribute{
			MarkdownDescription: "Serve stale enabled.",
			Computed:            true,
		},
		"serve_stale_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Serve stale TTL in seconds.",
			Computed:            true,
		},
		"serve_stale_answer_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Serve stale answer TTL.",
			Computed:            true,
		},
		"serve_stale_reset_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Serve stale reset TTL.",
			Computed:            true,
		},
		"serve_stale_max_wait_time": dschema.Int64Attribute{
			MarkdownDescription: "Serve stale max wait time.",
			Computed:            true,
		},
		"cache_maximum_entries": dschema.Int64Attribute{
			MarkdownDescription: "Cache maximum entries.",
			Computed:            true,
		},
		"cache_minimum_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Cache minimum record TTL.",
			Computed:            true,
		},
		"cache_maximum_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Cache maximum record TTL.",
			Computed:            true,
		},
		"cache_negative_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Cache negative record TTL.",
			Computed:            true,
		},
		"cache_failure_record_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Cache failure record TTL.",
			Computed:            true,
		},
		"cache_prefetch_eligibility": dschema.Int64Attribute{
			MarkdownDescription: "Cache prefetch eligibility.",
			Computed:            true,
		},
		"cache_prefetch_trigger": dschema.Int64Attribute{
			MarkdownDescription: "Cache prefetch trigger.",
			Computed:            true,
		},
		"cache_prefetch_sample_interval_in_minutes": dschema.Int64Attribute{
			MarkdownDescription: "Cache prefetch sample interval.",
			Computed:            true,
		},
		"cache_prefetch_sample_eligibility_hits_per_hour": dschema.Int64Attribute{
			MarkdownDescription: "Cache prefetch eligibility hits per hour.",
			Computed:            true,
		},

		// Blocking
		"enable_blocking": dschema.BoolAttribute{
			MarkdownDescription: "Domain blocking enabled.",
			Computed:            true,
		},
		"allow_txt_blocking_report": dschema.BoolAttribute{
			MarkdownDescription: "TXT blocking report enabled.",
			Computed:            true,
		},
		"blocking_bypass_list": dschema.ListAttribute{
			MarkdownDescription: "Blocking bypass list.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"blocking_type": dschema.StringAttribute{
			MarkdownDescription: "Blocking response type.",
			Computed:            true,
		},
		"blocking_answer_ttl": dschema.Int64Attribute{
			MarkdownDescription: "Blocking answer TTL.",
			Computed:            true,
		},
		"custom_blocking_addresses": dschema.ListAttribute{
			MarkdownDescription: "Custom blocking addresses.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"block_list_urls": dschema.ListAttribute{
			MarkdownDescription: "Block list URLs.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"block_list_update_interval_hours": dschema.Int64Attribute{
			MarkdownDescription: "Block list update interval in hours.",
			Computed:            true,
		},

		// Proxy
		"proxy_type": dschema.StringAttribute{
			MarkdownDescription: "Proxy type.",
			Computed:            true,
		},
		"proxy_address": dschema.StringAttribute{
			MarkdownDescription: "Proxy server address.",
			Computed:            true,
		},
		"proxy_port": dschema.Int64Attribute{
			MarkdownDescription: "Proxy server port.",
			Computed:            true,
		},
		"proxy_username": dschema.StringAttribute{
			MarkdownDescription: "Proxy server username.",
			Computed:            true,
		},
		"proxy_bypass": dschema.StringAttribute{
			MarkdownDescription: "Proxy bypass list.",
			Computed:            true,
		},

		// Forwarders
		"forwarders": dschema.ListAttribute{
			MarkdownDescription: "Forwarder addresses.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"forwarder_protocol": dschema.StringAttribute{
			MarkdownDescription: "Forwarder protocol.",
			Computed:            true,
		},
		"concurrent_forwarding": dschema.BoolAttribute{
			MarkdownDescription: "Concurrent forwarding enabled.",
			Computed:            true,
		},
		"forwarder_retries": dschema.Int64Attribute{
			MarkdownDescription: "Forwarder retries.",
			Computed:            true,
		},
		"forwarder_timeout": dschema.Int64Attribute{
			MarkdownDescription: "Forwarder timeout.",
			Computed:            true,
		},
		"forwarder_concurrency": dschema.Int64Attribute{
			MarkdownDescription: "Forwarder concurrency.",
			Computed:            true,
		},

		// Logging
		"logging_type": dschema.StringAttribute{
			MarkdownDescription: "Logging type.",
			Computed:            true,
		},
		"ignore_resolver_logs": dschema.BoolAttribute{
			MarkdownDescription: "Ignore resolver logs.",
			Computed:            true,
		},
		"log_queries": dschema.BoolAttribute{
			MarkdownDescription: "Log queries enabled.",
			Computed:            true,
		},
		"use_local_time": dschema.BoolAttribute{
			MarkdownDescription: "Use local time for logging.",
			Computed:            true,
		},
		"log_folder": dschema.StringAttribute{
			MarkdownDescription: "Log folder path.",
			Computed:            true,
		},
		"max_log_file_days": dschema.Int64Attribute{
			MarkdownDescription: "Max log file days.",
			Computed:            true,
		},
		"enable_in_memory_stats": dschema.BoolAttribute{
			MarkdownDescription: "In-memory stats enabled.",
			Computed:            true,
		},
		"max_stat_file_days": dschema.Int64Attribute{
			MarkdownDescription: "Max stat file days.",
			Computed:            true,
		},
	}
}

func (d *SettingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Retrieves the current Technitium DNS Server settings.",
		Attributes:          settingsDataSourceSchemaAttributes(),
	}
}

func (d *SettingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(model.DNSApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Internal error: expected model.DNSApiClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *SettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	d.reqMutex.Lock()
	defer d.reqMutex.Unlock()

	result, err := d.client.GetSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read DNS settings: %s", err))
		return
	}

	stateData := modelSettings2tf(*result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
}
