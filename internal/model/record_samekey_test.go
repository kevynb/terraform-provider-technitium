package model

import "testing"

func TestDNSRecordSameKey(t *testing.T) {
	cases := []struct {
		name string
		r1   DNSRecord
		r2   DNSRecord
		want bool
	}{
		{
			name: "A uses IP address or value",
			r1: DNSRecord{
				Type:      REC_A,
				Domain:    "example.com",
				IPAddress: "1.2.3.4",
			},
			r2: DNSRecord{
				Type:   REC_A,
				Domain: "example.com",
				Value:  "1.2.3.4",
			},
			want: true,
		},
		{
			name: "CNAME ignores target",
			r1: DNSRecord{
				Type:   REC_CNAME,
				Domain: "alias.example.com",
				CName:  "target-one.example.com",
			},
			r2: DNSRecord{
				Type:   REC_CNAME,
				Domain: "alias.example.com",
				CName:  "target-two.example.com",
			},
			want: true,
		},
		{
			name: "SRV matches port and target",
			r1: DNSRecord{
				Type:   REC_SRV,
				Domain: "_svc.example.com",
				Port:   443,
				Target: "_tcp",
			},
			r2: DNSRecord{
				Type:   REC_SRV,
				Domain: "_svc.example.com",
				Port:   443,
				Target: "_tcp",
			},
			want: true,
		},
		{
			name: "SRV different port does not match",
			r1: DNSRecord{
				Type:   REC_SRV,
				Domain: "_svc.example.com",
				Port:   443,
				Target: "_tcp",
			},
			r2: DNSRecord{
				Type:   REC_SRV,
				Domain: "_svc.example.com",
				Port:   8443,
				Target: "_tcp",
			},
			want: false,
		},
		{
			name: "TXT matches text",
			r1: DNSRecord{
				Type:   REC_TXT,
				Domain: "example.com",
				Text:   "hello",
			},
			r2: DNSRecord{
				Type:   REC_TXT,
				Domain: "example.com",
				Text:   "hello",
			},
			want: true,
		},
		{
			name: "TXT different text does not match",
			r1: DNSRecord{
				Type:   REC_TXT,
				Domain: "example.com",
				Text:   "hello",
			},
			r2: DNSRecord{
				Type:   REC_TXT,
				Domain: "example.com",
				Text:   "goodbye",
			},
			want: false,
		},
		{
			name: "unknown type does not match",
			r1: DNSRecord{
				Type:   DNSRecordType("BOGUS"),
				Domain: "example.com",
			},
			r2: DNSRecord{
				Type:   DNSRecordType("BOGUS"),
				Domain: "example.com",
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.r1.SameKey(tc.r2); got != tc.want {
				t.Fatalf("SameKey mismatch: got %v want %v", got, tc.want)
			}
		})
	}
}
