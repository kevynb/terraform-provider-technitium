package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveProviderConfig_Validation(t *testing.T) {
	cases := []struct {
		name              string
		confData          TechnitiumDNSProviderModel
		version           string
		env               map[string]string
		wantErrSummaries  []string
		wantPathBySummary map[string]path.Path
		wantNoDiagnostics bool
	}{
		{
			name: "missing url and token",
			confData: TechnitiumDNSProviderModel{
				APIURL: types.StringNull(),
				Token:  types.StringNull(),
			},
			version:          "dev",
			env:              map[string]string{},
			wantErrSummaries: []string{"Missing server URL Configuration", "Missing Token Configuration"},
			wantPathBySummary: map[string]path.Path{
				"Missing server URL Configuration": path.Root("url"),
			},
		},
		{
			name: "missing token only",
			confData: TechnitiumDNSProviderModel{
				APIURL: types.StringNull(),
				Token:  types.StringNull(),
			},
			version:          "dev",
			env:              map[string]string{"TECHNITIUM_API_URL": "https://example.test"},
			wantErrSummaries: []string{"Missing Token Configuration"},
		},
		{
			name: "unittest allows missing token",
			confData: TechnitiumDNSProviderModel{
				APIURL: types.StringNull(),
				Token:  types.StringNull(),
			},
			version:           "unittest",
			env:               map[string]string{"TECHNITIUM_API_URL": "https://example.test"},
			wantNoDiagnostics: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := resolveProviderConfig(tc.confData, tc.version, func(key string) string {
				return tc.env[key]
			})

			if tc.wantNoDiagnostics {
				if diags.HasError() {
					t.Fatalf("unexpected diagnostics: %+v", diags)
				}
				return
			}

			if !diags.HasError() {
				t.Fatalf("expected diagnostics")
			}

			for _, summary := range tc.wantErrSummaries {
				diagItem, ok := findDiagBySummary(diags, summary)
				if !ok {
					t.Fatalf("expected diagnostic summary %q", summary)
				}
				if wantPath, ok := tc.wantPathBySummary[summary]; ok {
					if withPath, ok := diagItem.(diag.DiagnosticWithPath); ok {
						if !withPath.Path().Equal(wantPath) {
							t.Fatalf("expected path %s, got %s", wantPath.String(), withPath.Path().String())
						}
					} else {
						t.Fatalf("expected diagnostic with path for %q", summary)
					}
				}
			}
		})
	}
}

func TestResolveProviderConfig_ConfigOverridesEnv(t *testing.T) {
	confData := TechnitiumDNSProviderModel{
		APIURL:                      types.StringValue("https://config.test"),
		Token:                       types.StringValue("config-token"),
		SkipCertificateVerification: types.BoolValue(true),
	}
	config, diags := resolveProviderConfig(confData, "dev", func(key string) string {
		if key == "TECHNITIUM_API_URL" {
			return "https://env.test"
		}
		if key == "TECHNITIUM_API_TOKEN" {
			return "env-token"
		}
		return ""
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if config.apiURL != "https://config.test" || config.token != "config-token" {
		t.Fatalf("expected config values to override env, got url=%q token=%q", config.apiURL, config.token)
	}
	if !config.skipCertificateVerification {
		t.Fatalf("expected skipCertificateVerification to be true")
	}
}

func TestResolveProviderConfig_SkipTLSVerificationDefaultFalse(t *testing.T) {
	confData := TechnitiumDNSProviderModel{
		APIURL: types.StringValue("https://config.test"),
		Token:  types.StringValue("config-token"),
	}
	config, diags := resolveProviderConfig(confData, "dev", func(string) string { return "" })

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if config.skipCertificateVerification {
		t.Fatalf("expected skipCertificateVerification to default to false")
	}
}

func findDiagBySummary(diags diag.Diagnostics, summary string) (diag.Diagnostic, bool) {
	for _, d := range diags {
		if d.Summary() == summary {
			return d, true
		}
	}
	return nil, false
}
