package provider

import (
	"strings"
	"testing"
)

func TestParseRecordImportID(t *testing.T) {
	cases := []struct {
		name           string
		id             string
		wantParts      recordImportParts
		wantErrContain string
	}{
		{
			name: "valid",
			id:   "example.com:@:A:1.2.3.4",
			wantParts: recordImportParts{
				zone:       "example.com",
				name:       "@",
				recordType: "A",
				value:      "1.2.3.4",
			},
		},
		{
			name: "valid value with colons",
			id:   "example.com:@:TXT:v=spf1 include:example.com ~all",
			wantParts: recordImportParts{
				zone:       "example.com",
				name:       "@",
				recordType: "TXT",
				value:      "v=spf1 include:example.com ~all",
			},
		},
		{
			name:           "invalid",
			id:             "bad",
			wantErrContain: "Import ID must be in format",
		},
		{
			name:           "invalid missing value",
			id:             "example.com:@:A:",
			wantErrContain: "Import ID must be in format",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRecordImportID(tc.id)
			if tc.wantErrContain != "" {
				if err == nil {
					t.Fatalf("expected error")
				}
				if !strings.Contains(err.Error(), tc.wantErrContain) {
					t.Fatalf("unexpected error: %s", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got != tc.wantParts {
				t.Fatalf("unexpected parts: %+v", got)
			}
		})
	}
}

func TestParseMXImportValue(t *testing.T) {
	cases := []struct {
		name           string
		value          string
		wantPreference int64
		wantExchange   string
		wantErrSummary string
	}{
		{
			name:           "valid",
			value:          "10:mail.example.com",
			wantPreference: 10,
			wantExchange:   "mail.example.com",
		},
		{
			name:           "invalid format",
			value:          "badvalue",
			wantErrSummary: "Invalid MX record format",
		},
		{
			name:           "invalid missing exchange",
			value:          "10:",
			wantErrSummary: "Invalid MX record format",
		},
		{
			name:           "invalid preference",
			value:          "nope:mail.example.com",
			wantErrSummary: "Invalid MX preference",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseMXImportValue(tc.value)
			if tc.wantErrSummary != "" {
				if err == nil {
					t.Fatalf("expected error")
				}
				diagErr, ok := err.(importValueError)
				if !ok || diagErr.summary != tc.wantErrSummary {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got.preference != tc.wantPreference || got.exchange != tc.wantExchange {
				t.Fatalf("unexpected mx data: %+v", got)
			}
		})
	}
}

func TestParseSRVImportValue(t *testing.T) {
	cases := []struct {
		name           string
		value          string
		wantPriority   int64
		wantWeight     int64
		wantPort       int64
		wantTarget     string
		wantErrSummary string
	}{
		{
			name:         "valid",
			value:        "10:20:443:service.example.com",
			wantPriority: 10,
			wantWeight:   20,
			wantPort:     443,
			wantTarget:   "service.example.com",
		},
		{
			name:           "invalid format",
			value:          "10:20:443",
			wantErrSummary: "Invalid SRV record format",
		},
		{
			name:           "invalid missing target",
			value:          "10:20:443:",
			wantErrSummary: "Invalid SRV record format",
		},
		{
			name:           "invalid priority",
			value:          "nope:20:443:svc",
			wantErrSummary: "Invalid SRV priority",
		},
		{
			name:           "invalid weight",
			value:          "10:nope:443:svc",
			wantErrSummary: "Invalid SRV weight",
		},
		{
			name:           "invalid port",
			value:          "10:20:nope:svc",
			wantErrSummary: "Invalid SRV port",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSRVImportValue(tc.value)
			if tc.wantErrSummary != "" {
				if err == nil {
					t.Fatalf("expected error")
				}
				diagErr, ok := err.(importValueError)
				if !ok || diagErr.summary != tc.wantErrSummary {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got.priority != tc.wantPriority || got.weight != tc.wantWeight || got.port != tc.wantPort || got.target != tc.wantTarget {
				t.Fatalf("unexpected srv data: %+v", got)
			}
		})
	}
}

func TestParseCAAImportValue(t *testing.T) {
	cases := []struct {
		name           string
		value          string
		wantFlags      string
		wantTag        string
		wantValue      string
		wantErrSummary string
	}{
		{
			name:      "valid",
			value:     "0:issue:letsencrypt.org",
			wantFlags: "0",
			wantTag:   "issue",
			wantValue: "letsencrypt.org",
		},
		{
			name:           "invalid format",
			value:          "bad",
			wantErrSummary: "Invalid CAA record format",
		},
		{
			name:           "invalid missing value",
			value:          "0:issue:",
			wantErrSummary: "Invalid CAA record format",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCAAImportValue(tc.value)
			if tc.wantErrSummary != "" {
				if err == nil {
					t.Fatalf("expected error")
				}
				diagErr, ok := err.(importValueError)
				if !ok || diagErr.summary != tc.wantErrSummary {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if got.flags != tc.wantFlags || got.tag != tc.wantTag || got.value != tc.wantValue {
				t.Fatalf("unexpected caa data: %+v", got)
			}
		})
	}
}
