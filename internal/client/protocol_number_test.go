package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Reproduces https://github.com/kevynb/terraform-provider-technitium/pull/5:
// Technitium returns rData.protocol as a JSON number (3, per RFC 4034) for
// the DNSKEY records it auto-generates on a DNSSEC-signed zone, but the
// provider's apiDNSRecordResponseItemRdata.Protocol field only accepted a
// string - breaking decode of the *entire* zone listing (GetRecords uses
// listZone=true and unmarshals every record in one response), not just the
// DNSKEY record itself.
const httpReplyZoneWithDNSKEY = `{
	"status": "ok",
	"response": {
		"zone": {"name": "test.com", "type": "Primary", "internal": false, "disabled": false},
		"records": [
			{
				"name": "test.com",
				"type": "DNSKEY",
				"ttl": 3600,
				"rData": {"protocol": 3}
			},
			{
				"name": "cn.test.com",
				"type": "CNAME",
				"ttl": 3600,
				"rData": {"cname": "something.other.com"}
			}
		]
	}
}`

func TestGetRecords_DNSKEYNumericProtocol(t *testing.T) {
	t.Parallel()
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, httpReplyZoneWithDNSKEY)
		}))
	defer ts.Close()

	c, err := NewClient(ts.URL, "dummyAPIToken", false)
	if err != nil {
		t.Fatal(err)
	}

	got, err := c.GetRecords(context.Background(), "test.com")
	if err != nil {
		t.Fatalf("GetRecords failed on a zone containing a DNSSEC-generated DNSKEY record: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 records (DNSKEY + CNAME), got %d", len(got))
	}

	var dnskey *string
	for _, r := range got {
		if string(r.Type) == "DNSKEY" {
			p := r.Protocol
			dnskey = &p
		}
	}
	if dnskey == nil {
		t.Fatal("DNSKEY record missing from results")
	}
	if *dnskey != "3" {
		t.Errorf("want Protocol \"3\", got %q", *dnskey)
	}
}
