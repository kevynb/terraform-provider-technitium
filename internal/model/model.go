//go:generate mockery --all

package model

import "context"

type DNSDomain string

type DNSRecordType string
type DNSRecordName string
type DNSRecordData string
type DNSRecordTTL uint32 // formally int32, but [0, 604800]
type DNSRecordPrio uint16
type DNSRecordSRVWeight uint16
type DNSRecordSRVProto string   // _tcp or _udp
type DNSRecordSRVService string // _ldap
type DNSRecordSRVPort uint16

const (
	REC_A     = DNSRecordType("A")
	REC_AAAA  = DNSRecordType("AAAA")
	REC_CNAME = DNSRecordType("CNAME")
	REC_MX    = DNSRecordType("MX")
	REC_NS    = DNSRecordType("NS")
	REC_SOA   = DNSRecordType("SOA")
	REC_SRV   = DNSRecordType("SRV")
	REC_TXT   = DNSRecordType("TXT")
	REC_PTR   = DNSRecordType("PTR")
	REC_NAPTR = DNSRecordType("NAPTR")
	REC_DNAME = DNSRecordType("DNAME")
	REC_DS    = DNSRecordType("DS")
	REC_SSHFP = DNSRecordType("SSHFP")
	REC_TLSA  = DNSRecordType("TLSA")
	REC_SVCB  = DNSRecordType("SVCB")
	REC_HTTPS = DNSRecordType("HTTPS")
	REC_URI   = DNSRecordType("URI")
	REC_CAA   = DNSRecordType("CAA")
	REC_ANAME = DNSRecordType("ANAME")
	REC_FWD   = DNSRecordType("FWD")
	REC_APP   = DNSRecordType("APP")
)

type DNSZoneType string

const (
	ZONE_PRIMARY            = DNSZoneType("Primary")
	ZONE_SECONDARY          = DNSZoneType("Secondary")
	ZONE_STUB               = DNSZoneType("Stub")
	ZONE_FORWARDER          = DNSZoneType("Forwarder")
	ZONE_SECONDARYFORWARDER = DNSZoneType("SecondaryForwarder")
	ZONE_CATALOG            = DNSZoneType("Catalog")
	ZONE_SECONDARYCATALOG   = DNSZoneType("SecondaryCatalog")
)

type DNSZone struct {
	Name         string      `json:"name"`
	Type         DNSZoneType `json:"type"`
	Internal     bool        `json:"internal"`
	DNSSecStatus string      `json:"dnssecStatus"`
	SOASerial    uint32      `json:"soaSerial"`
	Expiry       string      `json:"expiry"`
	IsExpired    bool        `json:"isExpired"`
	SyncFailed   bool        `json:"syncFailed"`
	LastModified string      `json:"lastModified"`
	Disabled     bool        `json:"disabled"`

	// Zone creation parameters
	Catalog                    string `json:"catalog,omitempty"`
	UseSoaSerialDateScheme     *bool  `json:"useSoaSerialDateScheme,omitempty"`
	PrimaryNameServerAddresses string `json:"primaryNameServerAddresses,omitempty"`
	ZoneTransferProtocol       string `json:"zoneTransferProtocol,omitempty"`
	TsigKeyName                string `json:"tsigKeyName,omitempty"`
	ValidateZone               *bool  `json:"validateZone,omitempty"`
	InitializeForwarder        *bool  `json:"initializeForwarder,omitempty"`
	Protocol                   string `json:"protocol,omitempty"`
	Forwarder                  string `json:"forwarder,omitempty"`
	DnssecValidation           *bool  `json:"dnssecValidation,omitempty"`
	ProxyType                  string `json:"proxyType,omitempty"`
	ProxyAddress               string `json:"proxyAddress,omitempty"`
	ProxyPort                  *int64 `json:"proxyPort,omitempty"`
	ProxyUsername              string `json:"proxyUsername,omitempty"`
	ProxyPassword              string `json:"proxyPassword,omitempty"`
}

type DNSRecord struct {
	Type   DNSRecordType // from the enum above
	Domain DNSRecordName // @ for top-level TXT/MX/A/NS...

	TTL DNSRecordTTL // min 600, def 3600

	Comments  string       // comment for the added resource
	ExpiryTTL DNSRecordTTL // automatically delete the record when the value in seconds elapses

	IPAddress       string // ip address, required for A or AAAA record
	Ptr             bool   // This option is used only for A and AAAA records.
	CreatePtrZone   bool   // This option is used for A and AAAA records.
	UpdateSvcbHints bool   // This option is used for A and AAAA records.

	NameServer string // This option is required for adding NS record.
	Glue       string // This optional parameter is used for adding NS record.

	CName string // This option is required for adding CNAME record.

	PtrName string // This option is required for adding PTR record.

	Exchange   string        // This option is required for adding MX record.
	Preference DNSRecordPrio // This option is required for adding MX record.

	Text      string //  This option is required for adding TXT record.
	SplitText bool   // Set to true for using new line char to split text into multiple character-strings for adding TXT record.

	Mailbox   string // for adding RP record.
	TxtDomain string // Set a TXT record's domain name for adding RP record.

	Priority DNSRecordPrio       // This parameter is required for adding the SRV record.
	Weight   DNSRecordSRVWeight  // This parameter is required for adding the SRV record.
	Port     DNSRecordSRVPort    // This parameter is required for adding the SRV record.
	Target   DNSRecordSRVService // This parameter is required for adding the SRV record.

	NaptrOrder       uint16 // This parameter is required for adding the NAPTR record.
	NaptrPreference  uint16 // This parameter is required for adding the NAPTR record.
	NaptrFlags       string // This parameter is required for adding the NAPTR record.
	NaptrServices    string // This parameter is required for adding the NAPTR record.
	NaptrRegexp      string // This parameter is required for adding the NAPTR record.
	NaptrReplacement string // This parameter is required for adding the NAPTR record.

	DName string // This parameter is required for adding DNAME record.

	KeyTag     uint16 // This parameter is required for adding DS record.
	Algorithm  string // This parameter is required for adding DS record.
	DigestType string // This parameter is required for adding DS record.
	Digest     string // This parameter is required for adding DS record.

	SshfpAlgorithm       string // This parameter is required for adding SSHFP record.
	SshfpFingerprintType string // This parameter is required for adding SSHFP record.
	SshfpFingerprint     string // This parameter is required for adding SSHFP record.

	TlsaCertificateUsage           string // This parameter is required for adding TLSA record.
	TlsaSelector                   string // This parameter is required for adding TLSA record.
	TlsaMatchingType               string // This parameter is required for adding TLSA record.
	TlsaCertificateAssociationData string // This parameter is required for adding TLSA record.

	SvcPriority   uint16 // This parameter is required for adding SCVB or HTTPS record.
	SvcTargetName string // This parameter is required for adding SCVB or HTTPS record.
	SvcParams     string // This parameter is required for adding SCVB or HTTPS record.

	AutoIpv4Hint bool // This parameter is optional for adding SCVB or HTTPS record.
	AutoIpv6Hint bool // This parameter is optional for adding SCVB or HTTPS record.

	UriPriority uint16 // This parameter is required for adding URI record.
	UriWeight   uint16 // This parameter is required for adding URI record.
	Uri         string // This parameter is required for adding URI record.

	Flags string // This parameter is required for adding the CAA record.
	Tag   string // This parameter is required for adding the CAA record.
	Value string // This parameter is required for adding the CAA record.

	AName string // This parameter is required for adding the ANAME record.

	Protocol          string // This parameter is optional for adding the FWD record (Udp, Tcp, Tls, Https, Quic).
	Forwarder         string // This parameter is required for adding the FWD record.
	ForwarderPriority uint16 // This parameter is required for adding the FWD record.
	DnssecValidation  bool   // This parameter is optional for adding the FWD record.
	ProxyType         string // This parameter is optional for adding the FWD record.
	ProxyAddress      string // This parameter is optional for adding the FWD record.
	ProxyPort         uint16 // This parameter is optional for adding the FWD record.
	ProxyUsername     string // This parameter is optional for adding the FWD record.
	ProxyPassword     string // This parameter is optional for adding the FWD record.

	AppName    string //  This parameter is required for adding the APP record.
	ClassPath  string //  This parameter is required for adding the APP record.
	RecordData string //  This parameter is required for adding the APP record.
}

// compare key field to determine if two records refer to the same object
//   - for CNAME there could be only 1 RR with the same name, TTL is the only value
//   - for A, TXT and NS there could be several (so need to match by data),
//   - MX matches the same way, value is ttl + prio (in theory, MX 0 and MX 10
//     could point to the same host in "data", but lets think that it is a perversion
//     and replace it with one record
//   - and SRV same if Port and Target are matched
//
// ...
func (r DNSRecord) SameKey(r1 DNSRecord) bool {
	if r.Type != r1.Type || r.Domain != r1.Domain {

		println("RType", r.Type, "R1Type", r1.Type, "Domain", r.Domain, "R1Domain", r1.Domain)
		return false
	}

	switch r.Type {
	case REC_A, REC_AAAA:
		ip1 := r.IPAddress
		if ip1 == "" {
			ip1 = r.Value
		}
		ip2 := r1.IPAddress
		if ip2 == "" {
			ip2 = r1.Value
		}
		return ip1 == ip2 && ip1 != ""
	case REC_CNAME, REC_ANAME, REC_DNAME:
		return true
	case REC_SRV:
		return r.Port == r1.Port && r.Target == r1.Target
	case REC_MX:
		return r.Exchange == r1.Exchange
	case REC_TXT:
		return r.Text == r1.Text
	case REC_PTR:
		return r.PtrName == r1.PtrName
	case REC_NS:
		return r.NameServer == r1.NameServer
	case REC_NAPTR:
		return r.NaptrFlags == r1.NaptrFlags && r.NaptrServices == r1.NaptrServices && r.NaptrRegexp == r1.NaptrRegexp && r.NaptrReplacement == r1.NaptrReplacement
	case REC_DS:
		return r.KeyTag == r1.KeyTag && r.Algorithm == r1.Algorithm && r.DigestType == r1.DigestType && r.Digest == r1.Digest
	case REC_SSHFP:
		return r.SshfpAlgorithm == r1.SshfpAlgorithm && r.SshfpFingerprintType == r1.SshfpFingerprintType && r.SshfpFingerprint == r1.SshfpFingerprint
	case REC_TLSA:
		return r.TlsaCertificateUsage == r1.TlsaCertificateUsage && r.TlsaSelector == r1.TlsaSelector && r.TlsaMatchingType == r1.TlsaMatchingType && r.TlsaCertificateAssociationData == r1.TlsaCertificateAssociationData
	case REC_SVCB, REC_HTTPS:
		return r.SvcTargetName == r1.SvcTargetName && r.SvcParams == r1.SvcParams
	case REC_URI:
		return r.UriPriority == r1.UriPriority && r.UriWeight == r1.UriWeight && r.Uri == r1.Uri
	case REC_CAA:
		return r.Flags == r1.Flags && r.Tag == r1.Tag && r.Value == r1.Value
	case REC_FWD:
		return r.Forwarder == r1.Forwarder
	case REC_APP:
		return r.AppName == r1.AppName && r.ClassPath == r1.ClassPath
	default:
		return false
	}
}

// DNSSettings represents the Technitium DNS server settings
type DNSSettings struct {
	// General
	DnsServerDomain              string   `json:"dnsServerDomain"`
	DnsServerLocalEndPoints      []string `json:"dnsServerLocalEndPoints"`
	DnsServerIPv4SourceAddresses []string `json:"dnsServerIPv4SourceAddresses"`
	DnsServerIPv6SourceAddresses []string `json:"dnsServerIPv6SourceAddresses"`
	DefaultRecordTtl             int64    `json:"defaultRecordTtl"`
	DefaultNsRecordTtl           int64    `json:"defaultNsRecordTtl"`
	DefaultSoaRecordTtl          int64    `json:"defaultSoaRecordTtl"`
	DefaultResponsiblePerson     string   `json:"defaultResponsiblePerson"`
	UseSoaSerialDateScheme       bool     `json:"useSoaSerialDateScheme"`
	MinSoaRefresh                int64    `json:"minSoaRefresh"`
	MinSoaRetry                  int64    `json:"minSoaRetry"`
	ZoneTransferAllowedNetworks  []string `json:"zoneTransferAllowedNetworks"`
	NotifyAllowedNetworks        []string `json:"notifyAllowedNetworks"`
	DnsAppsEnableAutomaticUpdate bool     `json:"dnsAppsEnableAutomaticUpdate"`

	// Network
	PreferIPv6              bool    `json:"preferIPv6"`
	EnableUdpSocketPool     bool    `json:"enableUdpSocketPool"`
	SocketPoolExcludedPorts []int64 `json:"socketPoolExcludedPorts"`
	UdpPayloadSize          int64   `json:"udpPayloadSize"`

	// DNSSEC
	DnssecValidation                 bool   `json:"dnssecValidation"`
	EDnsClientSubnet                 bool   `json:"eDnsClientSubnet"`
	EDnsClientSubnetIPv4PrefixLength int64  `json:"eDnsClientSubnetIPv4PrefixLength"`
	EDnsClientSubnetIPv6PrefixLength int64  `json:"eDnsClientSubnetIPv6PrefixLength"`
	EDnsClientSubnetIpv4Override     string `json:"eDnsClientSubnetIpv4Override"`
	EDnsClientSubnetIpv6Override     string `json:"eDnsClientSubnetIpv6Override"`

	// QPM Limits
	QpmLimitSampleMinutes           int64    `json:"qpmLimitSampleMinutes"`
	QpmLimitUdpTruncationPercentage int64    `json:"qpmLimitUdpTruncationPercentage"`
	QpmLimitBypassList              []string `json:"qpmLimitBypassList"`

	// Timeouts
	ClientTimeout                   int64 `json:"clientTimeout"`
	TcpSendTimeout                  int64 `json:"tcpSendTimeout"`
	TcpReceiveTimeout               int64 `json:"tcpReceiveTimeout"`
	QuicIdleTimeout                 int64 `json:"quicIdleTimeout"`
	QuicMaxInboundStreams           int64 `json:"quicMaxInboundStreams"`
	ListenBacklog                   int64 `json:"listenBacklog"`
	MaxConcurrentResolutionsPerCore int64 `json:"maxConcurrentResolutionsPerCore"`

	// Web Service
	WebServiceLocalAddresses              []string `json:"webServiceLocalAddresses"`
	WebServiceHttpPort                    int64    `json:"webServiceHttpPort"`
	WebServiceEnableTls                   bool     `json:"webServiceEnableTls"`
	WebServiceEnableHttp3                 bool     `json:"webServiceEnableHttp3"`
	WebServiceHttpToTlsRedirect           bool     `json:"webServiceHttpToTlsRedirect"`
	WebServiceUseSelfSignedTlsCertificate bool     `json:"webServiceUseSelfSignedTlsCertificate"`
	WebServiceTlsPort                     int64    `json:"webServiceTlsPort"`
	WebServiceTlsCertificatePath          string   `json:"webServiceTlsCertificatePath"`
	WebServiceTlsCertificatePassword      string   `json:"webServiceTlsCertificatePassword"`
	WebServiceRealIpHeader                string   `json:"webServiceRealIpHeader"`

	// DNS-over-X protocols
	EnableDnsOverUdpProxy bool  `json:"enableDnsOverUdpProxy"`
	EnableDnsOverTcpProxy bool  `json:"enableDnsOverTcpProxy"`
	EnableDnsOverHttp     bool  `json:"enableDnsOverHttp"`
	EnableDnsOverTls      bool  `json:"enableDnsOverTls"`
	EnableDnsOverHttps    bool  `json:"enableDnsOverHttps"`
	EnableDnsOverHttp3    bool  `json:"enableDnsOverHttp3"`
	EnableDnsOverQuic     bool  `json:"enableDnsOverQuic"`
	DnsOverUdpProxyPort   int64 `json:"dnsOverUdpProxyPort"`
	DnsOverTcpProxyPort   int64 `json:"dnsOverTcpProxyPort"`
	DnsOverHttpPort       int64 `json:"dnsOverHttpPort"`
	DnsOverTlsPort        int64 `json:"dnsOverTlsPort"`
	DnsOverHttpsPort      int64 `json:"dnsOverHttpsPort"`
	DnsOverQuicPort       int64 `json:"dnsOverQuicPort"`

	// Reverse Proxy & TLS for DNS protocols
	ReverseProxyNetworkACL    []string `json:"reverseProxyNetworkACL"`
	DnsTlsCertificatePath     string   `json:"dnsTlsCertificatePath"`
	DnsTlsCertificatePassword string   `json:"dnsTlsCertificatePassword"`
	DnsOverHttpRealIpHeader   string   `json:"dnsOverHttpRealIpHeader"`

	// Recursion
	Recursion             string   `json:"recursion"`
	RecursionNetworkACL   []string `json:"recursionNetworkACL"`
	RandomizeName         bool     `json:"randomizeName"`
	QnameMinimization     bool     `json:"qnameMinimization"`
	ResolverRetries       int64    `json:"resolverRetries"`
	ResolverTimeout       int64    `json:"resolverTimeout"`
	ResolverConcurrency   int64    `json:"resolverConcurrency"`
	ResolverMaxStackCount int64    `json:"resolverMaxStackCount"`

	// Cache
	SaveCache                                 bool  `json:"saveCache"`
	ServeStale                                bool  `json:"serveStale"`
	ServeStaleTtl                             int64 `json:"serveStaleTtl"`
	ServeStaleAnswerTtl                       int64 `json:"serveStaleAnswerTtl"`
	ServeStaleResetTtl                        int64 `json:"serveStaleResetTtl"`
	ServeStaleMaxWaitTime                     int64 `json:"serveStaleMaxWaitTime"`
	CacheMaximumEntries                       int64 `json:"cacheMaximumEntries"`
	CacheMinimumRecordTtl                     int64 `json:"cacheMinimumRecordTtl"`
	CacheMaximumRecordTtl                     int64 `json:"cacheMaximumRecordTtl"`
	CacheNegativeRecordTtl                    int64 `json:"cacheNegativeRecordTtl"`
	CacheFailureRecordTtl                     int64 `json:"cacheFailureRecordTtl"`
	CachePrefetchEligibility                  int64 `json:"cachePrefetchEligibility"`
	CachePrefetchTrigger                      int64 `json:"cachePrefetchTrigger"`
	CachePrefetchSampleIntervalInMinutes      int64 `json:"cachePrefetchSampleIntervalInMinutes"`
	CachePrefetchSampleEligibilityHitsPerHour int64 `json:"cachePrefetchSampleEligibilityHitsPerHour"`

	// Blocking
	EnableBlocking               bool     `json:"enableBlocking"`
	AllowTxtBlockingReport       bool     `json:"allowTxtBlockingReport"`
	BlockingBypassList           []string `json:"blockingBypassList"`
	BlockingType                 string   `json:"blockingType"`
	BlockingAnswerTtl            int64    `json:"blockingAnswerTtl"`
	CustomBlockingAddresses      []string `json:"customBlockingAddresses"`
	BlockListUrls                []string `json:"blockListUrls"`
	BlockListUpdateIntervalHours int64    `json:"blockListUpdateIntervalHours"`

	// Proxy
	ProxyType     string `json:"proxyType"`
	ProxyAddress  string `json:"proxyAddress"`
	ProxyPort     int64  `json:"proxyPort"`
	ProxyUsername string `json:"proxyUsername"`
	ProxyPassword string `json:"proxyPassword"`
	ProxyBypass   string `json:"proxyBypass"`

	// Forwarders
	Forwarders           []string `json:"forwarders"`
	ForwarderProtocol    string   `json:"forwarderProtocol"`
	ConcurrentForwarding bool     `json:"concurrentForwarding"`
	ForwarderRetries     int64    `json:"forwarderRetries"`
	ForwarderTimeout     int64    `json:"forwarderTimeout"`
	ForwarderConcurrency int64    `json:"forwarderConcurrency"`

	// Logging
	LoggingType         string `json:"loggingType"`
	IgnoreResolverLogs  bool   `json:"ignoreResolverLogs"`
	LogQueries          bool   `json:"logQueries"`
	UseLocalTime        bool   `json:"useLocalTime"`
	LogFolder           string `json:"logFolder"`
	MaxLogFileDays      int64  `json:"maxLogFileDays"`
	EnableInMemoryStats bool   `json:"enableInMemoryStats"`
	MaxStatFileDays     int64  `json:"maxStatFileDays"`
}

// client API interface
type DNSApiClient interface {
	GetRecords(ctx context.Context, domain DNSRecordName) ([]DNSRecord, error)
	GetZoneRecords(ctx context.Context, zoneName string) ([]DNSRecord, error)
	AddRecord(ctx context.Context, record DNSRecord) error
	UpdateRecord(ctx context.Context, oldRecord DNSRecord, newRecord DNSRecord) error
	DeleteRecord(ctx context.Context, record DNSRecord) error
	ListZones(ctx context.Context) ([]DNSZone, error)
	CreateZone(ctx context.Context, zone DNSZone) error
	DeleteZone(ctx context.Context, zoneName string) error
	GetSettings(ctx context.Context) (*DNSSettings, error)
	SetSettings(ctx context.Context, settings *DNSSettings) (*DNSSettings, error)
}
