package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

func TestTFZone2ModelMapping(t *testing.T) {
	cases := []struct {
		name  string
		input tfDNSZone
		want  zoneModelExpect
	}{
		{
			name: "all optional fields",
			input: tfDNSZone{
				Name:                       types.StringValue("example.com"),
				Type:                       types.StringValue("Primary"),
				Catalog:                    types.StringValue("catalog.example.com"),
				UseSoaSerialDateScheme:     types.BoolValue(true),
				PrimaryNameServerAddresses: types.StringValue("1.1.1.1"),
				ZoneTransferProtocol:       types.StringValue("Tcp"),
				TsigKeyName:                types.StringValue("tsig-key"),
				ValidateZone:               types.BoolValue(false),
				InitializeForwarder:        types.BoolValue(true),
				Protocol:                   types.StringValue("Udp"),
				Forwarder:                  types.StringValue("8.8.8.8"),
				DnssecValidation:           types.BoolValue(true),
				ProxyType:                  types.StringValue("Http"),
				ProxyAddress:               types.StringValue("10.0.0.1"),
				ProxyPort:                  types.Int64Value(8080),
				ProxyUsername:              types.StringValue("user"),
				ProxyPassword:              types.StringValue("pass"),
			},
			want: zoneModelExpect{
				name:                       "example.com",
				zoneType:                   model.ZONE_PRIMARY,
				catalog:                    ptrString("catalog.example.com"),
				useSoaSerialDateScheme:     ptrBool(true),
				primaryNameServerAddresses: ptrString("1.1.1.1"),
				zoneTransferProtocol:       ptrString("Tcp"),
				tsigKeyName:                ptrString("tsig-key"),
				validateZone:               ptrBool(false),
				initializeForwarder:        ptrBool(true),
				protocol:                   ptrString("Udp"),
				forwarder:                  ptrString("8.8.8.8"),
				dnssecValidation:           ptrBool(true),
				proxyType:                  ptrString("Http"),
				proxyAddress:               ptrString("10.0.0.1"),
				proxyPort:                  ptrInt64(8080),
				proxyUsername:              ptrString("user"),
				proxyPassword:              ptrString("pass"),
			},
		},
		{
			name: "null optionals",
			input: tfDNSZone{
				Name: types.StringValue("example.org"),
				Type: types.StringValue("Secondary"),
			},
			want: zoneModelExpect{
				name:     "example.org",
				zoneType: model.ZONE_SECONDARY,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tfZone2model(tc.input)
			assertZoneModel(t, got, tc.want)
		})
	}
}

func TestModelZone2TFMapping(t *testing.T) {
	cases := []struct {
		name  string
		input model.DNSZone
		want  wantTFZone
	}{
		{
			name: "all optional fields",
			input: func() model.DNSZone {
				useSoa := true
				validate := false
				initFwd := true
				dnssec := true
				proxyPort := int64(8443)
				return model.DNSZone{
					Name:                       "example.net",
					Type:                       model.ZONE_FORWARDER,
					Catalog:                    "catalog.example.net",
					UseSoaSerialDateScheme:     &useSoa,
					PrimaryNameServerAddresses: "2.2.2.2",
					ZoneTransferProtocol:       "Tls",
					TsigKeyName:                "tsig-zone",
					ValidateZone:               &validate,
					InitializeForwarder:        &initFwd,
					Protocol:                   "Https",
					Forwarder:                  "9.9.9.9",
					DnssecValidation:           &dnssec,
					ProxyType:                  "Socks5",
					ProxyAddress:               "10.0.0.2",
					ProxyPort:                  &proxyPort,
					ProxyUsername:              "proxy-user",
					ProxyPassword:              "proxy-pass",
				}
			}(),
			want: wantTFZone{
				name:                       ptrString("example.net"),
				zoneType:                   ptrString("Forwarder"),
				catalog:                    ptrString("catalog.example.net"),
				useSoaSerialDateScheme:     ptrBool(true),
				primaryNameServerAddresses: ptrString("2.2.2.2"),
				zoneTransferProtocol:       ptrString("Tls"),
				tsigKeyName:                ptrString("tsig-zone"),
				validateZone:               ptrBool(false),
				initializeForwarder:        ptrBool(true),
				protocol:                   ptrString("Https"),
				forwarder:                  ptrString("9.9.9.9"),
				dnssecValidation:           ptrBool(true),
				proxyType:                  ptrString("Socks5"),
				proxyAddress:               ptrString("10.0.0.2"),
				proxyPort:                  ptrInt64(8443),
				proxyUsername:              ptrString("proxy-user"),
				proxyPassword:              ptrString("proxy-pass"),
			},
		},
		{
			name: "empty optionals",
			input: model.DNSZone{
				Name: "example.io",
				Type: model.ZONE_PRIMARY,
			},
			want: wantTFZone{
				name:     ptrString("example.io"),
				zoneType: ptrString("Primary"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := modelZone2tf(tc.input)
			assertTFZone(t, got, tc.want)
		})
	}
}

func TestModelZone2TFDataSourceMapping(t *testing.T) {
	cases := []struct {
		name  string
		input model.DNSZone
		want  wantTFZoneDataSource
	}{
		{
			name: "full mapping",
			input: model.DNSZone{
				Name:         "example.ds",
				Type:         model.ZONE_STUB,
				Internal:     true,
				DNSSecStatus: "Unsigned",
				SOASerial:    12345,
				Expiry:       "2025-01-01T00:00:00Z",
				IsExpired:    false,
				SyncFailed:   true,
				LastModified: "2025-01-02T03:04:05Z",
				Disabled:     true,
			},
			want: wantTFZoneDataSource{
				name:         ptrString("example.ds"),
				zoneType:     ptrString("Stub"),
				internal:     ptrBool(true),
				dnssecStatus: ptrString("Unsigned"),
				soaSerial:    ptrInt64(12345),
				expiry:       ptrString("2025-01-01T00:00:00Z"),
				isExpired:    ptrBool(false),
				syncFailed:   ptrBool(true),
				lastModified: ptrString("2025-01-02T03:04:05Z"),
				disabled:     ptrBool(true),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := modelZone2tfDataSource(tc.input)
			assertTFZoneDataSource(t, got, tc.want)
		})
	}
}

type zoneModelExpect struct {
	name                       string
	zoneType                   model.DNSZoneType
	catalog                    *string
	useSoaSerialDateScheme     *bool
	primaryNameServerAddresses *string
	zoneTransferProtocol       *string
	tsigKeyName                *string
	validateZone               *bool
	initializeForwarder        *bool
	protocol                   *string
	forwarder                  *string
	dnssecValidation           *bool
	proxyType                  *string
	proxyAddress               *string
	proxyPort                  *int64
	proxyUsername              *string
	proxyPassword              *string
}

type wantTFZone struct {
	name                       *string
	zoneType                   *string
	catalog                    *string
	useSoaSerialDateScheme     *bool
	primaryNameServerAddresses *string
	zoneTransferProtocol       *string
	tsigKeyName                *string
	validateZone               *bool
	initializeForwarder        *bool
	protocol                   *string
	forwarder                  *string
	dnssecValidation           *bool
	proxyType                  *string
	proxyAddress               *string
	proxyPort                  *int64
	proxyUsername              *string
	proxyPassword              *string
}

type wantTFZoneDataSource struct {
	name         *string
	zoneType     *string
	internal     *bool
	dnssecStatus *string
	soaSerial    *int64
	expiry       *string
	isExpired    *bool
	syncFailed   *bool
	lastModified *string
	disabled     *bool
}

func assertZoneModel(t *testing.T, got model.DNSZone, want zoneModelExpect) {
	if got.Name != want.name || got.Type != want.zoneType {
		t.Fatalf("basic mapping mismatch: got name=%q type=%q", got.Name, got.Type)
	}
	assertStringField(t, "Catalog", got.Catalog, want.catalog)
	assertBoolPtrField(t, "UseSoaSerialDateScheme", got.UseSoaSerialDateScheme, want.useSoaSerialDateScheme)
	assertStringField(t, "PrimaryNameServerAddresses", got.PrimaryNameServerAddresses, want.primaryNameServerAddresses)
	assertStringField(t, "ZoneTransferProtocol", got.ZoneTransferProtocol, want.zoneTransferProtocol)
	assertStringField(t, "TsigKeyName", got.TsigKeyName, want.tsigKeyName)
	assertBoolPtrField(t, "ValidateZone", got.ValidateZone, want.validateZone)
	assertBoolPtrField(t, "InitializeForwarder", got.InitializeForwarder, want.initializeForwarder)
	assertStringField(t, "Protocol", got.Protocol, want.protocol)
	assertStringField(t, "Forwarder", got.Forwarder, want.forwarder)
	assertBoolPtrField(t, "DnssecValidation", got.DnssecValidation, want.dnssecValidation)
	assertStringField(t, "ProxyType", got.ProxyType, want.proxyType)
	assertStringField(t, "ProxyAddress", got.ProxyAddress, want.proxyAddress)
	assertInt64PtrField(t, "ProxyPort", got.ProxyPort, want.proxyPort)
	assertStringField(t, "ProxyUsername", got.ProxyUsername, want.proxyUsername)
	assertStringField(t, "ProxyPassword", got.ProxyPassword, want.proxyPassword)
}

func assertTFZone(t *testing.T, got tfDNSZone, want wantTFZone) {
	assertTFStringValue(t, "Name", got.Name, want.name)
	assertTFStringValue(t, "Type", got.Type, want.zoneType)
	assertTFStringValue(t, "Catalog", got.Catalog, want.catalog)
	assertTFBoolValue(t, "UseSoaSerialDateScheme", got.UseSoaSerialDateScheme, want.useSoaSerialDateScheme)
	assertTFStringValue(t, "PrimaryNameServerAddresses", got.PrimaryNameServerAddresses, want.primaryNameServerAddresses)
	assertTFStringValue(t, "ZoneTransferProtocol", got.ZoneTransferProtocol, want.zoneTransferProtocol)
	assertTFStringValue(t, "TsigKeyName", got.TsigKeyName, want.tsigKeyName)
	assertTFBoolValue(t, "ValidateZone", got.ValidateZone, want.validateZone)
	assertTFBoolValue(t, "InitializeForwarder", got.InitializeForwarder, want.initializeForwarder)
	assertTFStringValue(t, "Protocol", got.Protocol, want.protocol)
	assertTFStringValue(t, "Forwarder", got.Forwarder, want.forwarder)
	assertTFBoolValue(t, "DnssecValidation", got.DnssecValidation, want.dnssecValidation)
	assertTFStringValue(t, "ProxyType", got.ProxyType, want.proxyType)
	assertTFStringValue(t, "ProxyAddress", got.ProxyAddress, want.proxyAddress)
	assertTFInt64Value(t, "ProxyPort", got.ProxyPort, want.proxyPort)
	assertTFStringValue(t, "ProxyUsername", got.ProxyUsername, want.proxyUsername)
	assertTFStringValue(t, "ProxyPassword", got.ProxyPassword, want.proxyPassword)
}

func assertTFZoneDataSource(t *testing.T, got tfDNSZoneDataSource, want wantTFZoneDataSource) {
	assertTFStringValue(t, "Name", got.Name, want.name)
	assertTFStringValue(t, "Type", got.Type, want.zoneType)
	assertTFBoolValue(t, "Internal", got.Internal, want.internal)
	assertTFStringValue(t, "DNSSecStatus", got.DNSSecStatus, want.dnssecStatus)
	assertTFInt64Value(t, "SOASerial", got.SOASerial, want.soaSerial)
	assertTFStringValue(t, "Expiry", got.Expiry, want.expiry)
	assertTFBoolValue(t, "IsExpired", got.IsExpired, want.isExpired)
	assertTFBoolValue(t, "SyncFailed", got.SyncFailed, want.syncFailed)
	assertTFStringValue(t, "LastModified", got.LastModified, want.lastModified)
	assertTFBoolValue(t, "Disabled", got.Disabled, want.disabled)
}

func assertStringField(t *testing.T, field string, got string, want *string) {
	if want != nil {
		if got != *want {
			t.Fatalf("%s mismatch: got %q", field, got)
		}
		return
	}
	if got != "" {
		t.Fatalf("%s expected empty, got %q", field, got)
	}
}

func assertBoolPtrField(t *testing.T, field string, got *bool, want *bool) {
	if want != nil {
		if got == nil || *got != *want {
			t.Fatalf("%s mismatch: got %v", field, got)
		}
		return
	}
	if got != nil {
		t.Fatalf("%s expected nil, got %v", field, got)
	}
}

func assertInt64PtrField(t *testing.T, field string, got *int64, want *int64) {
	if want != nil {
		if got == nil || *got != *want {
			t.Fatalf("%s mismatch: got %v", field, got)
		}
		return
	}
	if got != nil {
		t.Fatalf("%s expected nil, got %v", field, got)
	}
}

func assertTFStringValue(t *testing.T, field string, got types.String, want *string) {
	if want != nil {
		if got.IsNull() || got.ValueString() != *want {
			t.Fatalf("%s mismatch: got %v", field, got)
		}
		return
	}
	if !got.IsNull() {
		t.Fatalf("%s expected null, got %v", field, got)
	}
}

func assertTFBoolValue(t *testing.T, field string, got types.Bool, want *bool) {
	if want != nil {
		if got.IsNull() || got.ValueBool() != *want {
			t.Fatalf("%s mismatch: got %v", field, got)
		}
		return
	}
	if !got.IsNull() {
		t.Fatalf("%s expected null, got %v", field, got)
	}
}

func assertTFInt64Value(t *testing.T, field string, got types.Int64, want *int64) {
	if want != nil {
		if got.IsNull() || got.ValueInt64() != *want {
			t.Fatalf("%s mismatch: got %v", field, got)
		}
		return
	}
	if !got.IsNull() {
		t.Fatalf("%s expected null, got %v", field, got)
	}
}

func ptrBool(v bool) *bool { return &v }

func ptrInt64(v int64) *int64 { return &v }

func ptrString(v string) *string { return &v }
