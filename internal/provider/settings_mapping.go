package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

// tfDNSSettings is the Terraform state representation of DNS settings.
type tfDNSSettings struct {
	ID types.String `tfsdk:"id"`

	// General
	DnsServerDomain              types.String `tfsdk:"dns_server_domain"`
	DnsServerLocalEndPoints      types.List   `tfsdk:"dns_server_local_end_points"`
	DnsServerIPv4SourceAddresses types.List   `tfsdk:"dns_server_ipv4_source_addresses"`
	DnsServerIPv6SourceAddresses types.List   `tfsdk:"dns_server_ipv6_source_addresses"`
	DefaultRecordTtl             types.Int64  `tfsdk:"default_record_ttl"`
	DefaultNsRecordTtl           types.Int64  `tfsdk:"default_ns_record_ttl"`
	DefaultSoaRecordTtl          types.Int64  `tfsdk:"default_soa_record_ttl"`
	DefaultResponsiblePerson     types.String `tfsdk:"default_responsible_person"`
	UseSoaSerialDateScheme       types.Bool   `tfsdk:"use_soa_serial_date_scheme"`
	MinSoaRefresh                types.Int64  `tfsdk:"min_soa_refresh"`
	MinSoaRetry                  types.Int64  `tfsdk:"min_soa_retry"`
	ZoneTransferAllowedNetworks  types.List   `tfsdk:"zone_transfer_allowed_networks"`
	NotifyAllowedNetworks        types.List   `tfsdk:"notify_allowed_networks"`
	DnsAppsEnableAutomaticUpdate types.Bool   `tfsdk:"dns_apps_enable_automatic_update"`

	// Network
	PreferIPv6          types.Bool  `tfsdk:"prefer_ipv6"`
	EnableUdpSocketPool types.Bool  `tfsdk:"enable_udp_socket_pool"`
	UdpPayloadSize      types.Int64 `tfsdk:"udp_payload_size"`

	// DNSSEC
	DnssecValidation                 types.Bool   `tfsdk:"dnssec_validation"`
	EDnsClientSubnet                 types.Bool   `tfsdk:"edns_client_subnet"`
	EDnsClientSubnetIPv4PrefixLength types.Int64  `tfsdk:"edns_client_subnet_ipv4_prefix_length"`
	EDnsClientSubnetIPv6PrefixLength types.Int64  `tfsdk:"edns_client_subnet_ipv6_prefix_length"`
	EDnsClientSubnetIpv4Override     types.String `tfsdk:"edns_client_subnet_ipv4_override"`
	EDnsClientSubnetIpv6Override     types.String `tfsdk:"edns_client_subnet_ipv6_override"`

	// QPM Limits
	QpmLimitSampleMinutes           types.Int64 `tfsdk:"qpm_limit_sample_minutes"`
	QpmLimitUdpTruncationPercentage types.Int64 `tfsdk:"qpm_limit_udp_truncation_percentage"`
	QpmLimitBypassList              types.List  `tfsdk:"qpm_limit_bypass_list"`

	// Timeouts
	ClientTimeout                   types.Int64 `tfsdk:"client_timeout"`
	TcpSendTimeout                  types.Int64 `tfsdk:"tcp_send_timeout"`
	TcpReceiveTimeout               types.Int64 `tfsdk:"tcp_receive_timeout"`
	QuicIdleTimeout                 types.Int64 `tfsdk:"quic_idle_timeout"`
	QuicMaxInboundStreams           types.Int64 `tfsdk:"quic_max_inbound_streams"`
	ListenBacklog                   types.Int64 `tfsdk:"listen_backlog"`
	MaxConcurrentResolutionsPerCore types.Int64 `tfsdk:"max_concurrent_resolutions_per_core"`

	// Web Service
	WebServiceLocalAddresses              types.List   `tfsdk:"web_service_local_addresses"`
	WebServiceHttpPort                    types.Int64  `tfsdk:"web_service_http_port"`
	WebServiceEnableTls                   types.Bool   `tfsdk:"web_service_enable_tls"`
	WebServiceEnableHttp3                 types.Bool   `tfsdk:"web_service_enable_http3"`
	WebServiceHttpToTlsRedirect           types.Bool   `tfsdk:"web_service_http_to_tls_redirect"`
	WebServiceUseSelfSignedTlsCertificate types.Bool   `tfsdk:"web_service_use_self_signed_tls_certificate"`
	WebServiceTlsPort                     types.Int64  `tfsdk:"web_service_tls_port"`
	WebServiceTlsCertificatePath          types.String `tfsdk:"web_service_tls_certificate_path"`
	WebServiceTlsCertificatePassword      types.String `tfsdk:"web_service_tls_certificate_password"`
	WebServiceRealIpHeader                types.String `tfsdk:"web_service_real_ip_header"`

	// DNS-over-X protocols
	EnableDnsOverUdpProxy types.Bool  `tfsdk:"enable_dns_over_udp_proxy"`
	EnableDnsOverTcpProxy types.Bool  `tfsdk:"enable_dns_over_tcp_proxy"`
	EnableDnsOverHttp     types.Bool  `tfsdk:"enable_dns_over_http"`
	EnableDnsOverTls      types.Bool  `tfsdk:"enable_dns_over_tls"`
	EnableDnsOverHttps    types.Bool  `tfsdk:"enable_dns_over_https"`
	EnableDnsOverHttp3    types.Bool  `tfsdk:"enable_dns_over_http3"`
	EnableDnsOverQuic     types.Bool  `tfsdk:"enable_dns_over_quic"`
	DnsOverUdpProxyPort   types.Int64 `tfsdk:"dns_over_udp_proxy_port"`
	DnsOverTcpProxyPort   types.Int64 `tfsdk:"dns_over_tcp_proxy_port"`
	DnsOverHttpPort       types.Int64 `tfsdk:"dns_over_http_port"`
	DnsOverTlsPort        types.Int64 `tfsdk:"dns_over_tls_port"`
	DnsOverHttpsPort      types.Int64 `tfsdk:"dns_over_https_port"`
	DnsOverQuicPort       types.Int64 `tfsdk:"dns_over_quic_port"`

	// Reverse Proxy & TLS for DNS
	ReverseProxyNetworkACL    types.List   `tfsdk:"reverse_proxy_network_acl"`
	DnsTlsCertificatePath     types.String `tfsdk:"dns_tls_certificate_path"`
	DnsTlsCertificatePassword types.String `tfsdk:"dns_tls_certificate_password"`
	DnsOverHttpRealIpHeader   types.String `tfsdk:"dns_over_http_real_ip_header"`

	// Recursion
	Recursion             types.String `tfsdk:"recursion"`
	RecursionNetworkACL   types.List   `tfsdk:"recursion_network_acl"`
	RandomizeName         types.Bool   `tfsdk:"randomize_name"`
	QnameMinimization     types.Bool   `tfsdk:"qname_minimization"`
	ResolverRetries       types.Int64  `tfsdk:"resolver_retries"`
	ResolverTimeout       types.Int64  `tfsdk:"resolver_timeout"`
	ResolverConcurrency   types.Int64  `tfsdk:"resolver_concurrency"`
	ResolverMaxStackCount types.Int64  `tfsdk:"resolver_max_stack_count"`

	// Cache
	SaveCache                                 types.Bool  `tfsdk:"save_cache"`
	ServeStale                                types.Bool  `tfsdk:"serve_stale"`
	ServeStaleTtl                             types.Int64 `tfsdk:"serve_stale_ttl"`
	ServeStaleAnswerTtl                       types.Int64 `tfsdk:"serve_stale_answer_ttl"`
	ServeStaleResetTtl                        types.Int64 `tfsdk:"serve_stale_reset_ttl"`
	ServeStaleMaxWaitTime                     types.Int64 `tfsdk:"serve_stale_max_wait_time"`
	CacheMaximumEntries                       types.Int64 `tfsdk:"cache_maximum_entries"`
	CacheMinimumRecordTtl                     types.Int64 `tfsdk:"cache_minimum_record_ttl"`
	CacheMaximumRecordTtl                     types.Int64 `tfsdk:"cache_maximum_record_ttl"`
	CacheNegativeRecordTtl                    types.Int64 `tfsdk:"cache_negative_record_ttl"`
	CacheFailureRecordTtl                     types.Int64 `tfsdk:"cache_failure_record_ttl"`
	CachePrefetchEligibility                  types.Int64 `tfsdk:"cache_prefetch_eligibility"`
	CachePrefetchTrigger                      types.Int64 `tfsdk:"cache_prefetch_trigger"`
	CachePrefetchSampleIntervalInMinutes      types.Int64 `tfsdk:"cache_prefetch_sample_interval_in_minutes"`
	CachePrefetchSampleEligibilityHitsPerHour types.Int64 `tfsdk:"cache_prefetch_sample_eligibility_hits_per_hour"`

	// Blocking
	EnableBlocking               types.Bool   `tfsdk:"enable_blocking"`
	AllowTxtBlockingReport       types.Bool   `tfsdk:"allow_txt_blocking_report"`
	BlockingBypassList           types.List   `tfsdk:"blocking_bypass_list"`
	BlockingType                 types.String `tfsdk:"blocking_type"`
	BlockingAnswerTtl            types.Int64  `tfsdk:"blocking_answer_ttl"`
	CustomBlockingAddresses      types.List   `tfsdk:"custom_blocking_addresses"`
	BlockListUrls                types.List   `tfsdk:"block_list_urls"`
	BlockListUpdateIntervalHours types.Int64  `tfsdk:"block_list_update_interval_hours"`

	// Proxy
	ProxyType     types.String `tfsdk:"proxy_type"`
	ProxyAddress  types.String `tfsdk:"proxy_address"`
	ProxyPort     types.Int64  `tfsdk:"proxy_port"`
	ProxyUsername types.String `tfsdk:"proxy_username"`
	ProxyPassword types.String `tfsdk:"proxy_password"`
	ProxyBypass   types.String `tfsdk:"proxy_bypass"`

	// Forwarders
	Forwarders           types.List   `tfsdk:"forwarders"`
	ForwarderProtocol    types.String `tfsdk:"forwarder_protocol"`
	ConcurrentForwarding types.Bool   `tfsdk:"concurrent_forwarding"`
	ForwarderRetries     types.Int64  `tfsdk:"forwarder_retries"`
	ForwarderTimeout     types.Int64  `tfsdk:"forwarder_timeout"`
	ForwarderConcurrency types.Int64  `tfsdk:"forwarder_concurrency"`

	// Logging
	LoggingType         types.String `tfsdk:"logging_type"`
	IgnoreResolverLogs  types.Bool   `tfsdk:"ignore_resolver_logs"`
	LogQueries          types.Bool   `tfsdk:"log_queries"`
	UseLocalTime        types.Bool   `tfsdk:"use_local_time"`
	LogFolder           types.String `tfsdk:"log_folder"`
	MaxLogFileDays      types.Int64  `tfsdk:"max_log_file_days"`
	EnableInMemoryStats types.Bool   `tfsdk:"enable_in_memory_stats"`
	MaxStatFileDays     types.Int64  `tfsdk:"max_stat_file_days"`
}

// Helper to convert []string to types.List of StringType
func stringSliceToTFList(values []string) types.List {
	if values == nil {
		values = []string{}
	}
	elems := make([]types.String, len(values))
	for i, v := range values {
		elems[i] = types.StringValue(v)
	}
	list, _ := types.ListValueFrom(nil, types.StringType, elems)
	return list
}

// Helper to convert types.List to []string
func tfListToStringSlice(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elems := list.Elements()
	result := make([]string, len(elems))
	for i, e := range elems {
		if sv, ok := e.(types.String); ok {
			result[i] = sv.ValueString()
		}
	}
	return result
}

func modelSettings2tf(s model.DNSSettings) tfDNSSettings {
	return tfDNSSettings{
		ID: types.StringValue("settings"),

		// General
		DnsServerDomain:              types.StringValue(s.DnsServerDomain),
		DnsServerLocalEndPoints:      stringSliceToTFList(s.DnsServerLocalEndPoints),
		DnsServerIPv4SourceAddresses: stringSliceToTFList(s.DnsServerIPv4SourceAddresses),
		DnsServerIPv6SourceAddresses: stringSliceToTFList(s.DnsServerIPv6SourceAddresses),
		DefaultRecordTtl:             types.Int64Value(s.DefaultRecordTtl),
		DefaultNsRecordTtl:           types.Int64Value(s.DefaultNsRecordTtl),
		DefaultSoaRecordTtl:          types.Int64Value(s.DefaultSoaRecordTtl),
		DefaultResponsiblePerson:     types.StringValue(s.DefaultResponsiblePerson),
		UseSoaSerialDateScheme:       types.BoolValue(s.UseSoaSerialDateScheme),
		MinSoaRefresh:                types.Int64Value(s.MinSoaRefresh),
		MinSoaRetry:                  types.Int64Value(s.MinSoaRetry),
		ZoneTransferAllowedNetworks:  stringSliceToTFList(s.ZoneTransferAllowedNetworks),
		NotifyAllowedNetworks:        stringSliceToTFList(s.NotifyAllowedNetworks),
		DnsAppsEnableAutomaticUpdate: types.BoolValue(s.DnsAppsEnableAutomaticUpdate),

		// Network
		PreferIPv6:          types.BoolValue(s.PreferIPv6),
		EnableUdpSocketPool: types.BoolValue(s.EnableUdpSocketPool),
		UdpPayloadSize:      types.Int64Value(s.UdpPayloadSize),

		// DNSSEC
		DnssecValidation:                 types.BoolValue(s.DnssecValidation),
		EDnsClientSubnet:                 types.BoolValue(s.EDnsClientSubnet),
		EDnsClientSubnetIPv4PrefixLength: types.Int64Value(s.EDnsClientSubnetIPv4PrefixLength),
		EDnsClientSubnetIPv6PrefixLength: types.Int64Value(s.EDnsClientSubnetIPv6PrefixLength),
		EDnsClientSubnetIpv4Override:     types.StringValue(s.EDnsClientSubnetIpv4Override),
		EDnsClientSubnetIpv6Override:     types.StringValue(s.EDnsClientSubnetIpv6Override),

		// QPM Limits
		QpmLimitSampleMinutes:           types.Int64Value(s.QpmLimitSampleMinutes),
		QpmLimitUdpTruncationPercentage: types.Int64Value(s.QpmLimitUdpTruncationPercentage),
		QpmLimitBypassList:              stringSliceToTFList(s.QpmLimitBypassList),

		// Timeouts
		ClientTimeout:                   types.Int64Value(s.ClientTimeout),
		TcpSendTimeout:                  types.Int64Value(s.TcpSendTimeout),
		TcpReceiveTimeout:               types.Int64Value(s.TcpReceiveTimeout),
		QuicIdleTimeout:                 types.Int64Value(s.QuicIdleTimeout),
		QuicMaxInboundStreams:           types.Int64Value(s.QuicMaxInboundStreams),
		ListenBacklog:                   types.Int64Value(s.ListenBacklog),
		MaxConcurrentResolutionsPerCore: types.Int64Value(s.MaxConcurrentResolutionsPerCore),

		// Web Service
		WebServiceLocalAddresses:              stringSliceToTFList(s.WebServiceLocalAddresses),
		WebServiceHttpPort:                    types.Int64Value(s.WebServiceHttpPort),
		WebServiceEnableTls:                   types.BoolValue(s.WebServiceEnableTls),
		WebServiceEnableHttp3:                 types.BoolValue(s.WebServiceEnableHttp3),
		WebServiceHttpToTlsRedirect:           types.BoolValue(s.WebServiceHttpToTlsRedirect),
		WebServiceUseSelfSignedTlsCertificate: types.BoolValue(s.WebServiceUseSelfSignedTlsCertificate),
		WebServiceTlsPort:                     types.Int64Value(s.WebServiceTlsPort),
		WebServiceTlsCertificatePath:          types.StringValue(s.WebServiceTlsCertificatePath),
		WebServiceTlsCertificatePassword:      types.StringValue(s.WebServiceTlsCertificatePassword),
		WebServiceRealIpHeader:                types.StringValue(s.WebServiceRealIpHeader),

		// DNS-over-X protocols
		EnableDnsOverUdpProxy: types.BoolValue(s.EnableDnsOverUdpProxy),
		EnableDnsOverTcpProxy: types.BoolValue(s.EnableDnsOverTcpProxy),
		EnableDnsOverHttp:     types.BoolValue(s.EnableDnsOverHttp),
		EnableDnsOverTls:      types.BoolValue(s.EnableDnsOverTls),
		EnableDnsOverHttps:    types.BoolValue(s.EnableDnsOverHttps),
		EnableDnsOverHttp3:    types.BoolValue(s.EnableDnsOverHttp3),
		EnableDnsOverQuic:     types.BoolValue(s.EnableDnsOverQuic),
		DnsOverUdpProxyPort:   types.Int64Value(s.DnsOverUdpProxyPort),
		DnsOverTcpProxyPort:   types.Int64Value(s.DnsOverTcpProxyPort),
		DnsOverHttpPort:       types.Int64Value(s.DnsOverHttpPort),
		DnsOverTlsPort:        types.Int64Value(s.DnsOverTlsPort),
		DnsOverHttpsPort:      types.Int64Value(s.DnsOverHttpsPort),
		DnsOverQuicPort:       types.Int64Value(s.DnsOverQuicPort),

		// Reverse Proxy & TLS for DNS
		ReverseProxyNetworkACL:    stringSliceToTFList(s.ReverseProxyNetworkACL),
		DnsTlsCertificatePath:     types.StringValue(s.DnsTlsCertificatePath),
		DnsTlsCertificatePassword: types.StringValue(s.DnsTlsCertificatePassword),
		DnsOverHttpRealIpHeader:   types.StringValue(s.DnsOverHttpRealIpHeader),

		// Recursion
		Recursion:             types.StringValue(s.Recursion),
		RecursionNetworkACL:   stringSliceToTFList(s.RecursionNetworkACL),
		RandomizeName:         types.BoolValue(s.RandomizeName),
		QnameMinimization:     types.BoolValue(s.QnameMinimization),
		ResolverRetries:       types.Int64Value(s.ResolverRetries),
		ResolverTimeout:       types.Int64Value(s.ResolverTimeout),
		ResolverConcurrency:   types.Int64Value(s.ResolverConcurrency),
		ResolverMaxStackCount: types.Int64Value(s.ResolverMaxStackCount),

		// Cache
		SaveCache:                                 types.BoolValue(s.SaveCache),
		ServeStale:                                types.BoolValue(s.ServeStale),
		ServeStaleTtl:                             types.Int64Value(s.ServeStaleTtl),
		ServeStaleAnswerTtl:                       types.Int64Value(s.ServeStaleAnswerTtl),
		ServeStaleResetTtl:                        types.Int64Value(s.ServeStaleResetTtl),
		ServeStaleMaxWaitTime:                     types.Int64Value(s.ServeStaleMaxWaitTime),
		CacheMaximumEntries:                       types.Int64Value(s.CacheMaximumEntries),
		CacheMinimumRecordTtl:                     types.Int64Value(s.CacheMinimumRecordTtl),
		CacheMaximumRecordTtl:                     types.Int64Value(s.CacheMaximumRecordTtl),
		CacheNegativeRecordTtl:                    types.Int64Value(s.CacheNegativeRecordTtl),
		CacheFailureRecordTtl:                     types.Int64Value(s.CacheFailureRecordTtl),
		CachePrefetchEligibility:                  types.Int64Value(s.CachePrefetchEligibility),
		CachePrefetchTrigger:                      types.Int64Value(s.CachePrefetchTrigger),
		CachePrefetchSampleIntervalInMinutes:      types.Int64Value(s.CachePrefetchSampleIntervalInMinutes),
		CachePrefetchSampleEligibilityHitsPerHour: types.Int64Value(s.CachePrefetchSampleEligibilityHitsPerHour),

		// Blocking
		EnableBlocking:               types.BoolValue(s.EnableBlocking),
		AllowTxtBlockingReport:       types.BoolValue(s.AllowTxtBlockingReport),
		BlockingBypassList:           stringSliceToTFList(s.BlockingBypassList),
		BlockingType:                 types.StringValue(s.BlockingType),
		BlockingAnswerTtl:            types.Int64Value(s.BlockingAnswerTtl),
		CustomBlockingAddresses:      stringSliceToTFList(s.CustomBlockingAddresses),
		BlockListUrls:                stringSliceToTFList(s.BlockListUrls),
		BlockListUpdateIntervalHours: types.Int64Value(s.BlockListUpdateIntervalHours),

		// Proxy
		ProxyType:     types.StringValue(s.ProxyType),
		ProxyAddress:  types.StringValue(s.ProxyAddress),
		ProxyPort:     types.Int64Value(s.ProxyPort),
		ProxyUsername: types.StringValue(s.ProxyUsername),
		ProxyPassword: types.StringValue(s.ProxyPassword),
		ProxyBypass:   types.StringValue(s.ProxyBypass),

		// Forwarders
		Forwarders:           stringSliceToTFList(s.Forwarders),
		ForwarderProtocol:    types.StringValue(s.ForwarderProtocol),
		ConcurrentForwarding: types.BoolValue(s.ConcurrentForwarding),
		ForwarderRetries:     types.Int64Value(s.ForwarderRetries),
		ForwarderTimeout:     types.Int64Value(s.ForwarderTimeout),
		ForwarderConcurrency: types.Int64Value(s.ForwarderConcurrency),

		// Logging
		LoggingType:         types.StringValue(s.LoggingType),
		IgnoreResolverLogs:  types.BoolValue(s.IgnoreResolverLogs),
		LogQueries:          types.BoolValue(s.LogQueries),
		UseLocalTime:        types.BoolValue(s.UseLocalTime),
		LogFolder:           types.StringValue(s.LogFolder),
		MaxLogFileDays:      types.Int64Value(s.MaxLogFileDays),
		EnableInMemoryStats: types.BoolValue(s.EnableInMemoryStats),
		MaxStatFileDays:     types.Int64Value(s.MaxStatFileDays),
	}
}

func tfStringValue(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}

	return v.ValueString()
}

func tfInt64Value(v types.Int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return 0
	}

	return v.ValueInt64()
}

func tfBoolValue(v types.Bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return false
	}

	return v.ValueBool()
}

func tfSettings2model(tf tfDNSSettings) model.DNSSettings {
	return model.DNSSettings{
		// General
		DnsServerDomain:              tfStringValue(tf.DnsServerDomain),
		DnsServerLocalEndPoints:      tfListToStringSlice(tf.DnsServerLocalEndPoints),
		DnsServerIPv4SourceAddresses: tfListToStringSlice(tf.DnsServerIPv4SourceAddresses),
		DnsServerIPv6SourceAddresses: tfListToStringSlice(tf.DnsServerIPv6SourceAddresses),
		DefaultRecordTtl:             tfInt64Value(tf.DefaultRecordTtl),
		DefaultNsRecordTtl:           tfInt64Value(tf.DefaultNsRecordTtl),
		DefaultSoaRecordTtl:          tfInt64Value(tf.DefaultSoaRecordTtl),
		DefaultResponsiblePerson:     tfStringValue(tf.DefaultResponsiblePerson),
		UseSoaSerialDateScheme:       tfBoolValue(tf.UseSoaSerialDateScheme),
		MinSoaRefresh:                tfInt64Value(tf.MinSoaRefresh),
		MinSoaRetry:                  tfInt64Value(tf.MinSoaRetry),
		ZoneTransferAllowedNetworks:  tfListToStringSlice(tf.ZoneTransferAllowedNetworks),
		NotifyAllowedNetworks:        tfListToStringSlice(tf.NotifyAllowedNetworks),
		DnsAppsEnableAutomaticUpdate: tfBoolValue(tf.DnsAppsEnableAutomaticUpdate),
		// Network
		PreferIPv6:          tfBoolValue(tf.PreferIPv6),
		EnableUdpSocketPool: tfBoolValue(tf.EnableUdpSocketPool),
		UdpPayloadSize:      tfInt64Value(tf.UdpPayloadSize),

		// DNSSEC
		DnssecValidation:                 tfBoolValue(tf.DnssecValidation),
		EDnsClientSubnet:                 tfBoolValue(tf.EDnsClientSubnet),
		EDnsClientSubnetIPv4PrefixLength: tfInt64Value(tf.EDnsClientSubnetIPv4PrefixLength),
		EDnsClientSubnetIPv6PrefixLength: tfInt64Value(tf.EDnsClientSubnetIPv6PrefixLength),
		EDnsClientSubnetIpv4Override:     tfStringValue(tf.EDnsClientSubnetIpv4Override),
		EDnsClientSubnetIpv6Override:     tfStringValue(tf.EDnsClientSubnetIpv6Override),

		// QPM Limits
		QpmLimitSampleMinutes:           tfInt64Value(tf.QpmLimitSampleMinutes),
		QpmLimitUdpTruncationPercentage: tfInt64Value(tf.QpmLimitUdpTruncationPercentage),
		QpmLimitBypassList:              tfListToStringSlice(tf.QpmLimitBypassList),

		// Timeouts
		ClientTimeout:                   tfInt64Value(tf.ClientTimeout),
		TcpSendTimeout:                  tfInt64Value(tf.TcpSendTimeout),
		TcpReceiveTimeout:               tfInt64Value(tf.TcpReceiveTimeout),
		QuicIdleTimeout:                 tfInt64Value(tf.QuicIdleTimeout),
		QuicMaxInboundStreams:           tfInt64Value(tf.QuicMaxInboundStreams),
		ListenBacklog:                   tf.ListenBacklog.ValueInt64(),
		MaxConcurrentResolutionsPerCore: tf.MaxConcurrentResolutionsPerCore.ValueInt64(),

		// Web Service
		WebServiceLocalAddresses:              tfListToStringSlice(tf.WebServiceLocalAddresses),
		WebServiceHttpPort:                    tf.WebServiceHttpPort.ValueInt64(),
		WebServiceEnableTls:                   tf.WebServiceEnableTls.ValueBool(),
		WebServiceEnableHttp3:                 tf.WebServiceEnableHttp3.ValueBool(),
		WebServiceHttpToTlsRedirect:           tf.WebServiceHttpToTlsRedirect.ValueBool(),
		WebServiceUseSelfSignedTlsCertificate: tf.WebServiceUseSelfSignedTlsCertificate.ValueBool(),
		WebServiceTlsPort:                     tf.WebServiceTlsPort.ValueInt64(),
		WebServiceTlsCertificatePath:          tf.WebServiceTlsCertificatePath.ValueString(),
		WebServiceTlsCertificatePassword:      tf.WebServiceTlsCertificatePassword.ValueString(),
		WebServiceRealIpHeader:                tf.WebServiceRealIpHeader.ValueString(),

		// DNS-over-X protocols
		EnableDnsOverUdpProxy: tf.EnableDnsOverUdpProxy.ValueBool(),
		EnableDnsOverTcpProxy: tf.EnableDnsOverTcpProxy.ValueBool(),
		EnableDnsOverHttp:     tf.EnableDnsOverHttp.ValueBool(),
		EnableDnsOverTls:      tf.EnableDnsOverTls.ValueBool(),
		EnableDnsOverHttps:    tf.EnableDnsOverHttps.ValueBool(),
		EnableDnsOverHttp3:    tf.EnableDnsOverHttp3.ValueBool(),
		EnableDnsOverQuic:     tf.EnableDnsOverQuic.ValueBool(),
		DnsOverUdpProxyPort:   tf.DnsOverUdpProxyPort.ValueInt64(),
		DnsOverTcpProxyPort:   tf.DnsOverTcpProxyPort.ValueInt64(),
		DnsOverHttpPort:       tf.DnsOverHttpPort.ValueInt64(),
		DnsOverTlsPort:        tf.DnsOverTlsPort.ValueInt64(),
		DnsOverHttpsPort:      tf.DnsOverHttpsPort.ValueInt64(),
		DnsOverQuicPort:       tf.DnsOverQuicPort.ValueInt64(),

		// Reverse Proxy & TLS for DNS
		ReverseProxyNetworkACL:    tfListToStringSlice(tf.ReverseProxyNetworkACL),
		DnsTlsCertificatePath:     tf.DnsTlsCertificatePath.ValueString(),
		DnsTlsCertificatePassword: tf.DnsTlsCertificatePassword.ValueString(),
		DnsOverHttpRealIpHeader:   tf.DnsOverHttpRealIpHeader.ValueString(),

		// Recursion
		Recursion:             tf.Recursion.ValueString(),
		RecursionNetworkACL:   tfListToStringSlice(tf.RecursionNetworkACL),
		RandomizeName:         tf.RandomizeName.ValueBool(),
		QnameMinimization:     tf.QnameMinimization.ValueBool(),
		ResolverRetries:       tf.ResolverRetries.ValueInt64(),
		ResolverTimeout:       tf.ResolverTimeout.ValueInt64(),
		ResolverConcurrency:   tf.ResolverConcurrency.ValueInt64(),
		ResolverMaxStackCount: tf.ResolverMaxStackCount.ValueInt64(),

		// Cache
		SaveCache:                                 tf.SaveCache.ValueBool(),
		ServeStale:                                tf.ServeStale.ValueBool(),
		ServeStaleTtl:                             tf.ServeStaleTtl.ValueInt64(),
		ServeStaleAnswerTtl:                       tf.ServeStaleAnswerTtl.ValueInt64(),
		ServeStaleResetTtl:                        tf.ServeStaleResetTtl.ValueInt64(),
		ServeStaleMaxWaitTime:                     tf.ServeStaleMaxWaitTime.ValueInt64(),
		CacheMaximumEntries:                       tf.CacheMaximumEntries.ValueInt64(),
		CacheMinimumRecordTtl:                     tf.CacheMinimumRecordTtl.ValueInt64(),
		CacheMaximumRecordTtl:                     tf.CacheMaximumRecordTtl.ValueInt64(),
		CacheNegativeRecordTtl:                    tf.CacheNegativeRecordTtl.ValueInt64(),
		CacheFailureRecordTtl:                     tf.CacheFailureRecordTtl.ValueInt64(),
		CachePrefetchEligibility:                  tf.CachePrefetchEligibility.ValueInt64(),
		CachePrefetchTrigger:                      tf.CachePrefetchTrigger.ValueInt64(),
		CachePrefetchSampleIntervalInMinutes:      tf.CachePrefetchSampleIntervalInMinutes.ValueInt64(),
		CachePrefetchSampleEligibilityHitsPerHour: tf.CachePrefetchSampleEligibilityHitsPerHour.ValueInt64(),

		// Blocking
		EnableBlocking:               tf.EnableBlocking.ValueBool(),
		AllowTxtBlockingReport:       tf.AllowTxtBlockingReport.ValueBool(),
		BlockingBypassList:           tfListToStringSlice(tf.BlockingBypassList),
		BlockingType:                 tf.BlockingType.ValueString(),
		BlockingAnswerTtl:            tf.BlockingAnswerTtl.ValueInt64(),
		CustomBlockingAddresses:      tfListToStringSlice(tf.CustomBlockingAddresses),
		BlockListUrls:                tfListToStringSlice(tf.BlockListUrls),
		BlockListUpdateIntervalHours: tf.BlockListUpdateIntervalHours.ValueInt64(),

		// Proxy
		ProxyType:     tf.ProxyType.ValueString(),
		ProxyAddress:  tf.ProxyAddress.ValueString(),
		ProxyPort:     tf.ProxyPort.ValueInt64(),
		ProxyUsername: tf.ProxyUsername.ValueString(),
		ProxyPassword: tf.ProxyPassword.ValueString(),
		ProxyBypass:   tf.ProxyBypass.ValueString(),

		// Forwarders
		Forwarders:           tfListToStringSlice(tf.Forwarders),
		ForwarderProtocol:    tf.ForwarderProtocol.ValueString(),
		ConcurrentForwarding: tf.ConcurrentForwarding.ValueBool(),
		ForwarderRetries:     tf.ForwarderRetries.ValueInt64(),
		ForwarderTimeout:     tf.ForwarderTimeout.ValueInt64(),
		ForwarderConcurrency: tf.ForwarderConcurrency.ValueInt64(),

		// Logging
		LoggingType:         tf.LoggingType.ValueString(),
		IgnoreResolverLogs:  tf.IgnoreResolverLogs.ValueBool(),
		LogQueries:          tf.LogQueries.ValueBool(),
		UseLocalTime:        tf.UseLocalTime.ValueBool(),
		LogFolder:           tf.LogFolder.ValueString(),
		MaxLogFileDays:      tf.MaxLogFileDays.ValueInt64(),
		EnableInMemoryStats: tf.EnableInMemoryStats.ValueBool(),
		MaxStatFileDays:     tf.MaxStatFileDays.ValueInt64(),
	}
}
