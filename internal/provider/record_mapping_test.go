package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
)

func TestTF2ModelMapping(t *testing.T) {
	ptrBool := func(v bool) *bool { return &v }
	ptrInt64 := func(v int64) *int64 { return &v }
	ptrString := func(v string) *string { return &v }

	cases := []struct {
		name               string
		input              tfDNSRecord
		wantType           model.DNSRecordType
		wantDomain         model.DNSRecordName
		wantTTL            *int64
		wantIPAddress      *string
		wantExchange       *string
		wantPreference     *int64
		wantPtr            *bool
		wantCreatePtrZone  *bool
		wantUpdateSvcbHint *bool
	}{
		{
			name: "A record",
			input: tfDNSRecord{
				Type:            types.StringValue("A"),
				Domain:          types.StringValue("host.example.com"),
				TTL:             types.Int64Value(3600),
				IPAddress:       types.StringValue("1.2.3.4"),
				Ptr:             types.BoolValue(true),
				CreatePtrZone:   types.BoolValue(true),
				UpdateSvcbHints: types.BoolValue(true),
			},
			wantType:           model.REC_A,
			wantDomain:         "host.example.com",
			wantTTL:            ptrInt64(3600),
			wantIPAddress:      ptrString("1.2.3.4"),
			wantPtr:            ptrBool(true),
			wantCreatePtrZone:  ptrBool(true),
			wantUpdateSvcbHint: ptrBool(true),
		},
		{
			name: "MX record",
			input: tfDNSRecord{
				Type:       types.StringValue("MX"),
				Domain:     types.StringValue("example.com"),
				TTL:        types.Int64Value(600),
				Exchange:   types.StringValue("mail.example.com"),
				Preference: types.Int64Value(10),
			},
			wantType:       model.REC_MX,
			wantDomain:     "example.com",
			wantTTL:        ptrInt64(600),
			wantExchange:   ptrString("mail.example.com"),
			wantPreference: ptrInt64(10),
		},
		{
			name: "CNAME record",
			input: tfDNSRecord{
				Type:   types.StringValue("CNAME"),
				Domain: types.StringValue("alias.example.com"),
				TTL:    types.Int64Value(300),
				CName:  types.StringValue("target.example.com"),
			},
			wantType:   model.REC_CNAME,
			wantDomain: "alias.example.com",
			wantTTL:    ptrInt64(300),
		},
		{
			name: "TXT record",
			input: tfDNSRecord{
				Type:      types.StringValue("TXT"),
				Domain:    types.StringValue("example.com"),
				TTL:       types.Int64Value(1200),
				Text:      types.StringValue("hello world"),
				SplitText: types.BoolValue(true),
			},
			wantType:   model.REC_TXT,
			wantDomain: "example.com",
			wantTTL:    ptrInt64(1200),
		},
		{
			name: "SRV record",
			input: tfDNSRecord{
				Type:     types.StringValue("SRV"),
				Domain:   types.StringValue("_service._tcp.example.com"),
				TTL:      types.Int64Value(600),
				Priority: types.Int64Value(10),
				Weight:   types.Int64Value(20),
				Port:     types.Int64Value(443),
				Target:   types.StringValue("srv.example.com"),
			},
			wantType:   model.REC_SRV,
			wantDomain: "_service._tcp.example.com",
			wantTTL:    ptrInt64(600),
		},
		{
			name: "CAA record",
			input: tfDNSRecord{
				Type:   types.StringValue("CAA"),
				Domain: types.StringValue("example.com"),
				TTL:    types.Int64Value(600),
				Flags:  types.StringValue("0"),
				Tag:    types.StringValue("issue"),
				Value:  types.StringValue("letsencrypt.org"),
			},
			wantType:   model.REC_CAA,
			wantDomain: "example.com",
			wantTTL:    ptrInt64(600),
		},
		{
			name: "Value field mapping",
			input: tfDNSRecord{
				Type:   types.StringValue("CAA"),
				Domain: types.StringValue("example.com"),
				TTL:    types.Int64Value(600),
				Value:  types.StringValue("value-field"),
			},
			wantType:   model.REC_CAA,
			wantDomain: "example.com",
			wantTTL:    ptrInt64(600),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tf2model(tc.input)

			if got.Type != tc.wantType {
				t.Fatalf("Type mismatch: got %q", got.Type)
			}
			if got.Domain != tc.wantDomain {
				t.Fatalf("Domain mismatch: got %q", got.Domain)
			}
			if tc.wantTTL != nil && int64(got.TTL) != *tc.wantTTL {
				t.Fatalf("TTL mismatch: got %d", got.TTL)
			}
			if tc.wantIPAddress != nil && got.IPAddress != *tc.wantIPAddress {
				t.Fatalf("IPAddress mismatch: got %q", got.IPAddress)
			}
			if tc.wantExchange != nil && got.Exchange != *tc.wantExchange {
				t.Fatalf("Exchange mismatch: got %q", got.Exchange)
			}
			if tc.wantPreference != nil && int64(got.Preference) != *tc.wantPreference {
				t.Fatalf("Preference mismatch: got %d", got.Preference)
			}
			if tc.wantPtr != nil && got.Ptr != *tc.wantPtr {
				t.Fatalf("Ptr mismatch: got %v", got.Ptr)
			}
			if tc.wantCreatePtrZone != nil && got.CreatePtrZone != *tc.wantCreatePtrZone {
				t.Fatalf("CreatePtrZone mismatch: got %v", got.CreatePtrZone)
			}
			if tc.wantUpdateSvcbHint != nil && got.UpdateSvcbHints != *tc.wantUpdateSvcbHint {
				t.Fatalf("UpdateSvcbHints mismatch: got %v", got.UpdateSvcbHints)
			}
		})
	}
}

func TestModel2TFMapping(t *testing.T) {
	cases := []struct {
		name   string
		input  model.DNSRecord
		assert func(t *testing.T, tfData tfDNSRecord)
	}{
		{
			name: "TXT record",
			input: model.DNSRecord{
				Type:      model.REC_TXT,
				Domain:    "example.com",
				TTL:       1200,
				Text:      "hello world",
				SplitText: true,
			},
			assert: func(t *testing.T, tfData tfDNSRecord) {
				if tfData.Type.IsNull() || tfData.Type.ValueString() != "TXT" {
					t.Fatalf("Type mapping mismatch: %v", tfData.Type)
				}
				if tfData.Domain.IsNull() || tfData.Domain.ValueString() != "example.com" {
					t.Fatalf("Domain mapping mismatch: %v", tfData.Domain)
				}
				if tfData.TTL.IsNull() || tfData.TTL.ValueInt64() != 1200 {
					t.Fatalf("TTL mapping mismatch: %v", tfData.TTL)
				}
				if tfData.Text.IsNull() || tfData.Text.ValueString() != "hello world" {
					t.Fatalf("Text mapping mismatch: %v", tfData.Text)
				}
				if tfData.SplitText.IsNull() || !tfData.SplitText.ValueBool() {
					t.Fatalf("SplitText mapping mismatch: %v", tfData.SplitText)
				}
			},
		},
		{
			name: "CNAME record",
			input: model.DNSRecord{
				Type:   model.REC_CNAME,
				Domain: "alias.example.com",
				TTL:    300,
				CName:  "target.example.com",
			},
			assert: func(t *testing.T, tfData tfDNSRecord) {
				if tfData.Type.IsNull() || tfData.Type.ValueString() != "CNAME" {
					t.Fatalf("Type mapping mismatch: %v", tfData.Type)
				}
				if tfData.Domain.IsNull() || tfData.Domain.ValueString() != "alias.example.com" {
					t.Fatalf("Domain mapping mismatch: %v", tfData.Domain)
				}
				if tfData.TTL.IsNull() || tfData.TTL.ValueInt64() != 300 {
					t.Fatalf("TTL mapping mismatch: %v", tfData.TTL)
				}
				if tfData.CName.IsNull() || tfData.CName.ValueString() != "target.example.com" {
					t.Fatalf("CName mapping mismatch: %v", tfData.CName)
				}
			},
		},
		{
			name: "SRV record",
			input: model.DNSRecord{
				Type:     model.REC_SRV,
				Domain:   "_service._tcp.example.com",
				TTL:      600,
				Priority: 10,
				Weight:   20,
				Port:     443,
				Target:   "_tcp",
			},
			assert: func(t *testing.T, tfData tfDNSRecord) {
				if tfData.Type.IsNull() || tfData.Type.ValueString() != "SRV" {
					t.Fatalf("Type mapping mismatch: %v", tfData.Type)
				}
				if tfData.Domain.IsNull() || tfData.Domain.ValueString() != "_service._tcp.example.com" {
					t.Fatalf("Domain mapping mismatch: %v", tfData.Domain)
				}
				if tfData.Priority.IsNull() || tfData.Priority.ValueInt64() != 10 {
					t.Fatalf("Priority mapping mismatch: %v", tfData.Priority)
				}
				if tfData.Weight.IsNull() || tfData.Weight.ValueInt64() != 20 {
					t.Fatalf("Weight mapping mismatch: %v", tfData.Weight)
				}
				if tfData.Port.IsNull() || tfData.Port.ValueInt64() != 443 {
					t.Fatalf("Port mapping mismatch: %v", tfData.Port)
				}
				if tfData.Target.IsNull() || tfData.Target.ValueString() != "_tcp" {
					t.Fatalf("Target mapping mismatch: %v", tfData.Target)
				}
			},
		},
		{
			name: "CAA record",
			input: model.DNSRecord{
				Type:   model.REC_CAA,
				Domain: "example.com",
				TTL:    600,
				Flags:  "0",
				Tag:    "issue",
				Value:  "letsencrypt.org",
			},
			assert: func(t *testing.T, tfData tfDNSRecord) {
				if tfData.Type.IsNull() || tfData.Type.ValueString() != "CAA" {
					t.Fatalf("Type mapping mismatch: %v", tfData.Type)
				}
				if tfData.Domain.IsNull() || tfData.Domain.ValueString() != "example.com" {
					t.Fatalf("Domain mapping mismatch: %v", tfData.Domain)
				}
				if tfData.Flags.IsNull() || tfData.Flags.ValueString() != "0" {
					t.Fatalf("Flags mapping mismatch: %v", tfData.Flags)
				}
				if tfData.Tag.IsNull() || tfData.Tag.ValueString() != "issue" {
					t.Fatalf("Tag mapping mismatch: %v", tfData.Tag)
				}
				if tfData.Value.IsNull() || tfData.Value.ValueString() != "letsencrypt.org" {
					t.Fatalf("Value mapping mismatch: %v", tfData.Value)
				}
			},
		},
		{
			name: "Value field mapping",
			input: model.DNSRecord{
				Type:   model.REC_CAA,
				Domain: "example.com",
				TTL:    600,
				Value:  "value-field",
			},
			assert: func(t *testing.T, tfData tfDNSRecord) {
				if tfData.Value.IsNull() || tfData.Value.ValueString() != "value-field" {
					t.Fatalf("Value mapping mismatch: %v", tfData.Value)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var tfData tfDNSRecord
			model2tf(tc.input, &tfData)
			tc.assert(t, tfData)
		})
	}
}
