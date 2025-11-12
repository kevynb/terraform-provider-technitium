package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/kevynb/terraform-provider-technitium/internal/model"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                  = &ZoneResource{}
	_ resource.ResourceWithConfigure     = &ZoneResource{}
	_ resource.ResourceWithImportState   = &ZoneResource{}
	_ datasource.DataSource              = &ZoneDataSource{}
	_ datasource.DataSourceWithConfigure = &ZoneDataSource{}
)

type tfDNSZone struct {
	Name                       types.String `tfsdk:"name"`
	Type                       types.String `tfsdk:"type"`
	Catalog                    types.String `tfsdk:"catalog"`
	UseSoaSerialDateScheme     types.Bool   `tfsdk:"use_soa_serial_date_scheme"`
	PrimaryNameServerAddresses types.String `tfsdk:"primary_name_server_addresses"`
	ZoneTransferProtocol       types.String `tfsdk:"zone_transfer_protocol"`
	TsigKeyName                types.String `tfsdk:"tsig_key_name"`
	ValidateZone               types.Bool   `tfsdk:"validate_zone"`
	InitializeForwarder        types.Bool   `tfsdk:"initialize_forwarder"`
	Protocol                   types.String `tfsdk:"protocol"`
	Forwarder                  types.String `tfsdk:"forwarder"`
	DnssecValidation           types.Bool   `tfsdk:"dnssec_validation"`
	ProxyType                  types.String `tfsdk:"proxy_type"`
	ProxyAddress               types.String `tfsdk:"proxy_address"`
	ProxyPort                  types.Int64  `tfsdk:"proxy_port"`
	ProxyUsername              types.String `tfsdk:"proxy_username"`
	ProxyPassword              types.String `tfsdk:"proxy_password"`
}

// ZoneResource defines the implementation of Technitium DNS zones
type ZoneResource struct {
	client   model.DNSApiClient
	reqMutex *sync.Mutex
}

func ZoneResourceFactory(m *sync.Mutex) func() resource.Resource {
	return func() resource.Resource {
		return &ZoneResource{reqMutex: m}
	}
}

func (r *ZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *ZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		MarkdownDescription: "Manages a DNS zone in Technitium DNS Server.",
		Attributes: map[string]rschema.Attribute{
			"name": rschema.StringAttribute{
				MarkdownDescription: "The domain name for the DNS zone.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": rschema.StringAttribute{
				MarkdownDescription: "The type of zone to create. Valid values are `Primary`, `Secondary`, `Stub`, `Forwarder`, `SecondaryForwarder`, `Catalog`, `SecondaryCatalog`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"catalog": rschema.StringAttribute{
				MarkdownDescription: "The name of the catalog zone to become its member zone. Valid only for `Primary`, `Stub`, and `Forwarder` zones.",
				Optional:            true,
			},
			"use_soa_serial_date_scheme": rschema.BoolAttribute{
				MarkdownDescription: "Set to `true` to enable using date scheme for SOA serial. Valid only with `Primary`, `Forwarder`, and `Catalog` zones.",
				Optional:            true,
			},
			"primary_name_server_addresses": rschema.StringAttribute{
				MarkdownDescription: "List of comma separated IP addresses or domain names of the primary name server. Required for `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones.",
				Optional:            true,
			},
			"zone_transfer_protocol": rschema.StringAttribute{
				MarkdownDescription: "The zone transfer protocol to be used by `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones. Valid values are `Tcp`, `Tls`, `Quic`.",
				Optional:            true,
			},
			"tsig_key_name": rschema.StringAttribute{
				MarkdownDescription: "The TSIG key name to be used by `Secondary`, `SecondaryForwarder`, and `SecondaryCatalog` zones.",
				Optional:            true,
			},
			"validate_zone": rschema.BoolAttribute{
				MarkdownDescription: "Set to `true` to enable ZONEMD validation. Valid only for `Secondary` zones.",
				Optional:            true,
			},
			"initialize_forwarder": rschema.BoolAttribute{
				MarkdownDescription: "Set to `true` to initialize the Conditional Forwarder zone with an FWD record. Valid for Conditional Forwarder zones.",
				Optional:            true,
			},
			"protocol": rschema.StringAttribute{
				MarkdownDescription: "The DNS transport protocol to be used by the Conditional Forwarder zone. Valid values are `Udp`, `Tcp`, `Tls`, `Https`, `Quic`.",
				Optional:            true,
			},
			"forwarder": rschema.StringAttribute{
				MarkdownDescription: "The address of the DNS server to be used as a forwarder. Required for Conditional Forwarder zones.",
				Optional:            true,
			},
			"dnssec_validation": rschema.BoolAttribute{
				MarkdownDescription: "Set to `true` to enable DNSSEC validation. Valid for Conditional Forwarder zones.",
				Optional:            true,
			},
			"proxy_type": rschema.StringAttribute{
				MarkdownDescription: "The type of proxy to be used for conditional forwarding. Valid values are `NoProxy`, `DefaultProxy`, `Http`, `Socks5`.",
				Optional:            true,
			},
			"proxy_address": rschema.StringAttribute{
				MarkdownDescription: "The proxy server address.",
				Optional:            true,
			},
			"proxy_port": rschema.Int64Attribute{
				MarkdownDescription: "The proxy server port.",
				Optional:            true,
			},
			"proxy_username": rschema.StringAttribute{
				MarkdownDescription: "The proxy server username.",
				Optional:            true,
			},
			"proxy_password": rschema.StringAttribute{
				MarkdownDescription: "The proxy server password.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *ZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// or it will panic on none
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(model.DNSApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Internal error: expected *model.DNSApiClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData tfDNSZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setZoneLogCtx(ctx, planData, "create")
	tflog.Info(ctx, "create: start")
	defer tflog.Info(ctx, "create: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	apiZone := tfZone2model(planData)

	err := r.client.CreateZone(ctx, apiZone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create zone: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *ZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfDNSZone
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setZoneLogCtx(ctx, stateData, "read")
	tflog.Info(ctx, "read: start")
	defer tflog.Info(ctx, "read: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	// Get all zones and find the matching one
	zones, err := r.client.ListZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Reading DNS zones: query failed: %s", err))
		return
	}

	zoneName := stateData.Name.ValueString()
	for _, zone := range zones {
		if zone.Name == zoneName {
			stateData = modelZone2tf(zone)
			resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
			return
		}
	}

	// Zone not found, remove from state
	resp.State.RemoveResource(ctx)
}

func (r *ZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData tfDNSZone
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setZoneLogCtx(ctx, planData, "update")
	tflog.Info(ctx, "update: start")
	defer tflog.Info(ctx, "update: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	// For now, zones are immutable - delete and recreate
	var stateData tfDNSZone
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete old zone
	err := r.client.DeleteZone(ctx, stateData.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete old zone: %s", err))
		return
	}

	// Create new zone
	apiZone := tfZone2model(planData)
	err = r.client.CreateZone(ctx, apiZone)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create new zone: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *ZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateData tfDNSZone
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setZoneLogCtx(ctx, stateData, "delete")
	tflog.Info(ctx, "delete: start")
	defer tflog.Info(ctx, "delete: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	err := r.client.DeleteZone(ctx, stateData.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Deleting DNS zone failed: %s", err))
		return
	}
}

// terraform import technitium_zone.example example.com
func (r *ZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zoneName := req.ID

	// Set the zone name in the state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), zoneName)...)

	// Set a default type since it's required - this will be updated by Read()
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), "Primary")...)
}

// ZoneDataSource defines the data source implementation
type ZoneDataSource struct {
	client   model.DNSApiClient
	reqMutex *sync.Mutex
}

func ZoneDataSourceFactory(m *sync.Mutex) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &ZoneDataSource{reqMutex: m}
	}
}

func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a DNS zone in Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The domain name of the DNS zone.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the zone.",
				Computed:            true,
			},
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Whether the zone is internal.",
				Computed:            true,
			},
			"dnssec_status": schema.StringAttribute{
				MarkdownDescription: "The DNSSEC status of the zone.",
				Computed:            true,
			},
			"soa_serial": schema.Int64Attribute{
				MarkdownDescription: "The SOA serial number.",
				Computed:            true,
			},
			"expiry": schema.StringAttribute{
				MarkdownDescription: "The expiry time of the zone.",
				Computed:            true,
			},
			"is_expired": schema.BoolAttribute{
				MarkdownDescription: "Whether the zone is expired.",
				Computed:            true,
			},
			"sync_failed": schema.BoolAttribute{
				MarkdownDescription: "Whether the last sync failed.",
				Computed:            true,
			},
			"last_modified": schema.StringAttribute{
				MarkdownDescription: "The last modified time.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the zone is disabled.",
				Computed:            true,
			},
		},
	}
}

func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(model.DNSApiClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Internal error: expected *model.DNSApiClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tfDNSZoneDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d.reqMutex.Lock()
	defer d.reqMutex.Unlock()

	// Get all zones and find the matching one
	zones, err := d.client.ListZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Reading DNS zones: query failed: %s", err))
		return
	}

	zoneName := config.Name.ValueString()
	for _, zone := range zones {
		if zone.Name == zoneName {
			result := modelZone2tfDataSource(zone)
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
			return
		}
	}

	resp.Diagnostics.AddError("Zone not found",
		fmt.Sprintf("Zone with name '%s' not found", zoneName))
}

type tfDNSZoneDataSource struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Internal     types.Bool   `tfsdk:"internal"`
	DNSSecStatus types.String `tfsdk:"dnssec_status"`
	SOASerial    types.Int64  `tfsdk:"soa_serial"`
	Expiry       types.String `tfsdk:"expiry"`
	IsExpired    types.Bool   `tfsdk:"is_expired"`
	SyncFailed   types.Bool   `tfsdk:"sync_failed"`
	LastModified types.String `tfsdk:"last_modified"`
	Disabled     types.Bool   `tfsdk:"disabled"`
}

// Helper functions

func setZoneLogCtx(ctx context.Context, tfZone tfDNSZone, op string) context.Context {
	logAttributes := map[string]interface{}{
		"operation": op,
		"name":      tfZone.Name.ValueString(),
		"type":      tfZone.Type.ValueString(),
	}

	for k, v := range logAttributes {
		if v != nil && v != "" {
			ctx = tflog.SetField(ctx, k, v)
		}
	}

	return ctx
}

func tfZone2model(tfData tfDNSZone) model.DNSZone {
	return model.DNSZone{
		Name: tfData.Name.ValueString(),
		Type: model.DNSZoneType(tfData.Type.ValueString()),
	}
}

func modelZone2tf(apiData model.DNSZone) tfDNSZone {
	return tfDNSZone{
		Name: types.StringValue(apiData.Name),
		Type: types.StringValue(string(apiData.Type)),
	}
}

func modelZone2tfDataSource(apiData model.DNSZone) tfDNSZoneDataSource {
	return tfDNSZoneDataSource{
		Name:         types.StringValue(apiData.Name),
		Type:         types.StringValue(string(apiData.Type)),
		Internal:     types.BoolValue(apiData.Internal),
		DNSSecStatus: types.StringValue(apiData.DNSSecStatus),
		SOASerial:    types.Int64Value(int64(apiData.SOASerial)),
		Expiry:       types.StringValue(apiData.Expiry),
		IsExpired:    types.BoolValue(apiData.IsExpired),
		SyncFailed:   types.BoolValue(apiData.SyncFailed),
		LastModified: types.StringValue(apiData.LastModified),
		Disabled:     types.BoolValue(apiData.Disabled),
	}
}
