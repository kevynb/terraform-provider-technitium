package provider

import (
	"context"
	"os"
	"sync"

	"github.com/kevynb/terraform-provider-technitium/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/provider
var _ provider.Provider = &TechnitiumDNSProvider{}

type APIClientFactory func(apiURL, token string, skipCertificateVerification bool) (model.DNSApiClient, error)

type TechnitiumDNSProvider struct {
	// "dev" for local testing, "test" for acceptance tests, "v1.2.3" for prod
	version       string
	clientFactory APIClientFactory
	reqMutex      sync.Mutex
}

func (p *TechnitiumDNSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	// common prefix for resources
	resp.TypeName = "technitium"
	// set in configure
	resp.Version = p.version
}

// have to match schema
type TechnitiumDNSProviderModel struct {
	APIURL                      types.String `tfsdk:"url"`
	Token                       types.String `tfsdk:"token"`
	SkipCertificateVerification types.Bool   `tfsdk:"skip_certificate_verification"`
}

func (p *TechnitiumDNSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		// full documentation: conf in templates
		// see https://github.com/hashicorp/terraform-provider-tls/blob/main/templates/index.md.tmpl
		MarkdownDescription: "Technitium DNS provider",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The Technitium server URL.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Technitium API token.",
				Optional:            true,
				Sensitive:           true,
			},
			"skip_certificate_verification": schema.BoolAttribute{
				MarkdownDescription: "Skip https certificate verification. Useful for servers using self-signed certificates.",
				Optional:            true,
			},
		},
	}
}

func (p *TechnitiumDNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var confData TechnitiumDNSProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &confData)...) // Extract config data

	apiURL := os.Getenv("TECHNITIUM_API_URL")
	if !confData.APIURL.IsUnknown() && !confData.APIURL.IsNull() {
		apiURL = confData.APIURL.ValueString()
	}
	if apiURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing server URL Configuration",
			"While configuring the provider, the technitium server url was not found in "+
				"the TECHNITIUM_API_URL environment variable or provider "+
				"configuration block url attribute.",
		)
		return
	}

	token := os.Getenv("TECHNITIUM_API_TOKEN")
	if !confData.Token.IsUnknown() && !confData.Token.IsNull() {
		token = confData.Token.ValueString()
	}
	if token == "" && p.version != "unittest" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the TECHNITIUM_API_TOKEN environment variable or provider "+
				"configuration block token attribute.",
		)
		return
	}

	skipCertificateVerification := false
	if !confData.SkipCertificateVerification.IsUnknown() && !confData.SkipCertificateVerification.IsNull() {
		skipCertificateVerification = confData.SkipCertificateVerification.ValueBool()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := p.clientFactory(apiURL, token, skipCertificateVerification)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API client", err.Error())
		return
	}

	resp.ResourceData = client
}

func (p *TechnitiumDNSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		RecordResourceFactory(&p.reqMutex),
		ZoneResourceFactory(&p.reqMutex),
	}
}

func (p *TechnitiumDNSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func New(version string, clientFactory APIClientFactory) func() provider.Provider {
	return func() provider.Provider {
		return &TechnitiumDNSProvider{
			version:       version,
			clientFactory: clientFactory,
			reqMutex:      sync.Mutex{},
		}
	}
}
