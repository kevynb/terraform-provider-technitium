package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/kevynb/terraform-provider-technitium/internal/model"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// import separator
const IMPORT_SEP = ":"

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &RecordResource{}
	_ resource.ResourceWithConfigure   = &RecordResource{}
	_ resource.ResourceWithImportState = &RecordResource{}
)

type tfDNSRecord struct {
	Zone                           types.String `tfsdk:"zone"`
	Type                           types.String `tfsdk:"type"`
	Domain                         types.String `tfsdk:"domain"`
	TTL                            types.Int64  `tfsdk:"ttl"`
	IPAddress                      types.String `tfsdk:"ip_address"`
	Ptr                            types.Bool   `tfsdk:"ptr"`
	CreatePtrZone                  types.Bool   `tfsdk:"create_ptr_zone"`
	UpdateSvcbHints                types.Bool   `tfsdk:"update_svcb_hints"`
	NameServer                     types.String `tfsdk:"name_server"`
	Glue                           types.String `tfsdk:"glue"`
	CName                          types.String `tfsdk:"cname"`
	PtrName                        types.String `tfsdk:"ptr_name"`
	Exchange                       types.String `tfsdk:"exchange"`
	Preference                     types.Int64  `tfsdk:"preference"`
	Text                           types.String `tfsdk:"text"`
	SplitText                      types.Bool   `tfsdk:"split_text"`
	Mailbox                        types.String `tfsdk:"mailbox"`
	TxtDomain                      types.String `tfsdk:"txt_domain"`
	Priority                       types.Int64  `tfsdk:"priority"`
	Weight                         types.Int64  `tfsdk:"weight"`
	Port                           types.Int64  `tfsdk:"port"`
	Target                         types.String `tfsdk:"target"`
	NaptrOrder                     types.Int64  `tfsdk:"naptr_order"`
	NaptrPreference                types.Int64  `tfsdk:"naptr_preference"`
	NaptrFlags                     types.String `tfsdk:"naptr_flags"`
	NaptrServices                  types.String `tfsdk:"naptr_services"`
	NaptrRegexp                    types.String `tfsdk:"naptr_regexp"`
	NaptrReplacement               types.String `tfsdk:"naptr_replacement"`
	DName                          types.String `tfsdk:"dname"`
	KeyTag                         types.Int64  `tfsdk:"key_tag"`
	Algorithm                      types.String `tfsdk:"algorithm"`
	DigestType                     types.String `tfsdk:"digest_type"`
	Digest                         types.String `tfsdk:"digest"`
	SshfpAlgorithm                 types.String `tfsdk:"sshfp_algorithm"`
	SshfpFingerprintType           types.String `tfsdk:"sshfp_fingerprint_type"`
	SshfpFingerprint               types.String `tfsdk:"sshfp_fingerprint"`
	TlsaCertificateUsage           types.String `tfsdk:"tlsa_certificate_usage"`
	TlsaSelector                   types.String `tfsdk:"tlsa_selector"`
	TlsaMatchingType               types.String `tfsdk:"tlsa_matching_type"`
	TlsaCertificateAssociationData types.String `tfsdk:"tlsa_certificate_association_data"`
	SvcPriority                    types.Int64  `tfsdk:"svc_priority"`
	SvcTargetName                  types.String `tfsdk:"svc_target_name"`
	SvcParams                      types.String `tfsdk:"svc_params"`
	AutoIpv4Hint                   types.Bool   `tfsdk:"auto_ipv4_hint"`
	AutoIpv6Hint                   types.Bool   `tfsdk:"auto_ipv6_hint"`
	UriPriority                    types.Int64  `tfsdk:"uri_priority"`
	UriWeight                      types.Int64  `tfsdk:"uri_weight"`
	Uri                            types.String `tfsdk:"uri"`
	Flags                          types.String `tfsdk:"flags"`
	Tag                            types.String `tfsdk:"tag"`
	Value                          types.String `tfsdk:"value"`
	AName                          types.String `tfsdk:"aname"`
	Forwarder                      types.String `tfsdk:"forwarder"`
	ForwarderPriority              types.Int64  `tfsdk:"forwarder_priority"`
	DnssecValidation               types.Bool   `tfsdk:"dnssec_validation"`
	ProxyType                      types.String `tfsdk:"proxy_type"`
	ProxyAddress                   types.String `tfsdk:"proxy_address"`
	ProxyPort                      types.Int64  `tfsdk:"proxy_port"`
	ProxyUsername                  types.String `tfsdk:"proxy_username"`
	ProxyPassword                  types.String `tfsdk:"proxy_password"`
	AppName                        types.String `tfsdk:"app_name"`
	ClassPath                      types.String `tfsdk:"class_path"`
	RecordData                     types.String `tfsdk:"record_data"`
}

// RecordResource defines the implementation of Technitium DNS records
type RecordResource struct {
	client   model.DNSApiClient
	reqMutex *sync.Mutex
}

func RecordResourceFactory(m *sync.Mutex) func() resource.Resource {
	return func() resource.Resource {
		return &RecordResource{reqMutex: m}
	}
}

func (r *RecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_record"
}

func (r *RecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DNS record in Technitium DNS Server.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "The DNS zone name. If not specified, it will be inferred from the domain.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS record type (e.g., A, AAAA, CNAME, etc.).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "MX", "NS", "SOA", "SRV", "TXT", "PTR", "NAPTR", "DNAME", "DS", "SSHFP", "TLSA", "SVCB", "HTTPS", "URI", "CAA", "ANAME", "FWD", "APP"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name for the DNS record (FQN)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "The time-to-live (TTL) of the DNS record, in seconds.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 604800),
				},
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The IP address for A or AAAA records.",
				Optional:            true,
			},
			"ptr": schema.BoolAttribute{
				MarkdownDescription: "Specifies if this record should create a PTR record for A/AAAA types.",
				Optional:            true,
			},
			"create_ptr_zone": schema.BoolAttribute{
				MarkdownDescription: "Specifies if the PTR zone should be automatically created for A/AAAA records.",
				Optional:            true,
			},
			"update_svcb_hints": schema.BoolAttribute{
				MarkdownDescription: "Whether to update SVCB hints for this record.",
				Optional:            true,
			},
			"name_server": schema.StringAttribute{
				MarkdownDescription: "The name server for NS records.",
				Optional:            true,
			},
			"glue": schema.StringAttribute{
				MarkdownDescription: "The glue record for NS records.",
				Optional:            true,
			},
			"cname": schema.StringAttribute{
				MarkdownDescription: "The canonical name for CNAME records.",
				Optional:            true,
			},
			"ptr_name": schema.StringAttribute{
				MarkdownDescription: "The PTR name for PTR records.",
				Optional:            true,
			},
			"exchange": schema.StringAttribute{
				MarkdownDescription: "The exchange server for MX records.",
				Optional:            true,
			},
			"preference": schema.Int64Attribute{
				MarkdownDescription: "The priority for MX records.",
				Optional:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The text value for TXT records.",
				Optional:            true,
			},
			"split_text": schema.BoolAttribute{
				MarkdownDescription: "Whether to split TXT record text into multiple character strings.",
				Optional:            true,
			},
			"mailbox": schema.StringAttribute{
				MarkdownDescription: "The mailbox for RP records.",
				Optional:            true,
			},
			"txt_domain": schema.StringAttribute{
				MarkdownDescription: "The TXT domain for RP records.",
				Optional:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority for SRV records.",
				Optional:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "The weight for SRV records.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port for SRV records.",
				Optional:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: "The target for SRV records.",
				Optional:            true,
			},
			"naptr_order": schema.Int64Attribute{
				MarkdownDescription: "The order for NAPTR records.",
				Optional:            true,
			},
			"naptr_preference": schema.Int64Attribute{
				MarkdownDescription: "The preference for NAPTR records.",
				Optional:            true,
			},
			"naptr_flags": schema.StringAttribute{
				MarkdownDescription: "The flags for NAPTR records.",
				Optional:            true,
			},
			"naptr_services": schema.StringAttribute{
				MarkdownDescription: "The services for NAPTR records.",
				Optional:            true,
			},
			"naptr_regexp": schema.StringAttribute{
				MarkdownDescription: "The regular expression for NAPTR records.",
				Optional:            true,
			},
			"naptr_replacement": schema.StringAttribute{
				MarkdownDescription: "The replacement field for NAPTR records.",
				Optional:            true,
			},
			"dname": schema.StringAttribute{
				MarkdownDescription: "The DNAME for DNAME records.",
				Optional:            true,
			},
			"key_tag": schema.Int64Attribute{
				MarkdownDescription: "The key tag for DS records.",
				Optional:            true,
			},
			"algorithm": schema.StringAttribute{
				MarkdownDescription: "The algorithm for DS records.",
				Optional:            true,
			},
			"digest_type": schema.StringAttribute{
				MarkdownDescription: "The digest type for DS records.",
				Optional:            true,
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "The digest for DS records.",
				Optional:            true,
			},
			"sshfp_algorithm": schema.StringAttribute{
				MarkdownDescription: "The SSHFP algorithm.",
				Optional:            true,
			},
			"sshfp_fingerprint_type": schema.StringAttribute{
				MarkdownDescription: "The SSHFP fingerprint type.",
				Optional:            true,
			},
			"sshfp_fingerprint": schema.StringAttribute{
				MarkdownDescription: "The SSHFP fingerprint.",
				Optional:            true,
			},
			"tlsa_certificate_usage": schema.StringAttribute{
				MarkdownDescription: "The TLSA certificate usage.",
				Optional:            true,
			},
			"tlsa_selector": schema.StringAttribute{
				MarkdownDescription: "The TLSA selector.",
				Optional:            true,
			},
			"tlsa_matching_type": schema.StringAttribute{
				MarkdownDescription: "The TLSA matching type.",
				Optional:            true,
			},
			"tlsa_certificate_association_data": schema.StringAttribute{
				MarkdownDescription: "The TLSA certificate association data.",
				Optional:            true,
			},
			"svc_priority": schema.Int64Attribute{
				MarkdownDescription: "The priority for SVCB/HTTPS records.",
				Optional:            true,
			},
			"svc_target_name": schema.StringAttribute{
				MarkdownDescription: "The target name for SVCB/HTTPS records.",
				Optional:            true,
			},
			"svc_params": schema.StringAttribute{
				MarkdownDescription: "The parameters for SVCB/HTTPS records.",
				Optional:            true,
			},
			"auto_ipv4_hint": schema.BoolAttribute{
				MarkdownDescription: "Whether to use automatic IPv4 hints for SVCB/HTTPS records.",
				Optional:            true,
			},
			"auto_ipv6_hint": schema.BoolAttribute{
				MarkdownDescription: "Whether to use automatic IPv6 hints for SVCB/HTTPS records.",
				Optional:            true,
			},
			"uri_priority": schema.Int64Attribute{
				MarkdownDescription: "The priority for URI records.",
				Optional:            true,
			},
			"uri_weight": schema.Int64Attribute{
				MarkdownDescription: "The weight for URI records.",
				Optional:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The URI for URI records.",
				Optional:            true,
			},
			"flags": schema.StringAttribute{
				MarkdownDescription: "The flags for CAA records.",
				Optional:            true,
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: "The tag for CAA records.",
				Optional:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The value for CAA records.",
				Optional:            true,
			},
			"aname": schema.StringAttribute{
				MarkdownDescription: "The ANAME value.",
				Optional:            true,
			},
			"forwarder": schema.StringAttribute{
				MarkdownDescription: "The forwarder address for FWD records.",
				Optional:            true,
			},
			"forwarder_priority": schema.Int64Attribute{
				MarkdownDescription: "The priority for FWD records.",
				Optional:            true,
			},
			"dnssec_validation": schema.BoolAttribute{
				MarkdownDescription: "Whether DNSSEC validation is enabled for FWD records.",
				Optional:            true,
			},
			"proxy_type": schema.StringAttribute{
				MarkdownDescription: "The proxy type for FWD records.",
				Optional:            true,
			},
			"proxy_address": schema.StringAttribute{
				MarkdownDescription: "The proxy address for FWD records.",
				Optional:            true,
			},
			"proxy_port": schema.Int64Attribute{
				MarkdownDescription: "The proxy port for FWD records.",
				Optional:            true,
			},
			"proxy_username": schema.StringAttribute{
				MarkdownDescription: "The proxy username for FWD records.",
				Optional:            true,
			},
			"proxy_password": schema.StringAttribute{
				MarkdownDescription: "The proxy password for FWD records.",
				Optional:            true,
				Sensitive:           true,
			},
			"app_name": schema.StringAttribute{
				MarkdownDescription: "The app name for APP records.",
				Optional:            true,
			},
			"class_path": schema.StringAttribute{
				MarkdownDescription: "The class path for APP records.",
				Optional:            true,
			},
			"record_data": schema.StringAttribute{
				MarkdownDescription: "The record data for APP records.",
				Optional:            true,
			},
		},
	}
}

func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// create will complain (and fail with client error) if same record is already present
// (mb as a result of calling "apply" with updated config with old record already gone)
// so state must be manually imported to continue (could step around this, but this will
// contradict terraform ideology -- see below)
func (r *RecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData tfDNSRecord
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setLogCtx(ctx, planData, "create")
	tflog.Info(ctx, "create: start")
	defer tflog.Info(ctx, "create: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	apiRecPlan := tf2model(planData)
	// "put"/"add" does not check prior state (terraform does not provide one for Create)
	// and so will fail on uniqueness violation (e.g. if record already exists
	// after external modification, or if it is the second CNAME etc)
	// - lets think it is ok for now -- let API do checking + run "import" if required
	// - alt/TODO: read records and do noop if target record is already there
	//   like `apiAllRecs, err := r.client.GetRecords(ctx, apiDomain, apiRecPlan.Type, apiRecPlan.Name)`
	//   but lets not be silent about that
	err := r.client.AddRecord(ctx, apiRecPlan)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create record: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

// TODO: The read function might need some caching mechanism because it is currently refetching the full record list every time.
func (r *RecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var stateData tfDNSRecord
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setLogCtx(ctx, stateData, "read")
	tflog.Info(ctx, "read: start")
	defer tflog.Info(ctx, "read: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	dnsRecordFromState := tf2model(stateData)

	allRecordsFromApi, err := r.client.GetRecords(ctx, dnsRecordFromState.Domain)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Reading DNS records: query failed: %s", err))
		return
	}
	numFound := 0
	if numberOfApiRecords := len(allRecordsFromApi); numberOfApiRecords == 0 {
		tflog.Debug(ctx, "Reading DNS record: currently absent")
	} else {
		tflog.Info(ctx, fmt.Sprintf(
			"Reading DNS record: got %d answers", numberOfApiRecords))
		// Look for a matching record to define if the resource was changed.
		for _, dnsRecordFromApi := range allRecordsFromApi {
			tflog.Debug(ctx, fmt.Sprintf("Got DNS record: %v", dnsRecordFromApi))
			if dnsRecordFromApi.SameKey(dnsRecordFromState) {
				tflog.Info(ctx, "matching DNS record found")
				model2tf(dnsRecordFromApi, &stateData)
				tflog.Info(ctx, " AutoIpv6Hint value "+stateData.AutoIpv6Hint.String())
				numFound += 1
			}
		}
	}

	if numFound == 0 {
		// mb quite ok, e.g. on creation
		tflog.Info(ctx, "Resource is currently absent")
		resp.State.RemoveResource(ctx)
	} else {
		if numFound > 1 {
			// unlikely to happen (mb several MXes with the same target?)
			tflog.Warn(ctx, "More than one instance of a resource present")
			resp.Diagnostics.AddWarning(
				"Duplicate resource instances present",
				"Will use the last one")
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &stateData)...)
	}
}

func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var planData tfDNSRecord
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setLogCtx(ctx, planData, "update")
	tflog.Info(ctx, "update: start")
	defer tflog.Info(ctx, "update: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	dnsRecordFromPlan := tf2model(planData)

	var stateData tfDNSRecord
	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dnsRecordFromState := tf2model(stateData)

	err := r.client.UpdateRecord(ctx, dnsRecordFromState, dnsRecordFromPlan)

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Updating DNS failed: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

func (r *RecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var stateData tfDNSRecord

	resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = setLogCtx(ctx, stateData, "delete")
	tflog.Info(ctx, "delete: start")
	defer tflog.Info(ctx, "delete: end")
	r.reqMutex.Lock()
	defer r.reqMutex.Unlock()

	dnsRecordFromState := tf2model(stateData)

	err := r.client.DeleteRecord(ctx, dnsRecordFromState)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Deleting DNS record failed: %s", err))
		return
	}
}

// terraform import technitium_record.new-cname zone:name:TYPE:value
func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID

	// Parse the import ID: zone:name:TYPE:value
	parts := strings.Split(id, IMPORT_SEP)
	if len(parts) != 4 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Import ID must be in format 'zone:name:TYPE:value', got: %s", id),
		)
		return
	}

	zone := parts[0]
	name := parts[1]
	recordType := parts[2]
	value := parts[3]

	// Construct full domain name
	var domain string
	if name == "@" {
		domain = zone
	} else {
		domain = name + "." + zone
	}

	// Set the domain and type
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), domain)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), recordType)...)

	// Set the value based on record type
	switch recordType {
	case "A", "AAAA":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip_address"), value)...)
	case "CNAME":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cname"), value)...)
	case "MX":
		// MX format: preference exchange
		mxParts := strings.SplitN(value, " ", 2)
		if len(mxParts) == 2 {
			if pref, err := strconv.ParseInt(mxParts[0], 10, 64); err == nil {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("preference"), pref)...)
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("exchange"), mxParts[1])...)
		}
	case "NS":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name_server"), value)...)
	case "PTR":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ptr_name"), value)...)
	case "SRV":
		// SRV format: priority weight port target
		srvParts := strings.Split(value, " ")
		if len(srvParts) >= 4 {
			if prio, err := strconv.ParseInt(srvParts[0], 10, 64); err == nil {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("priority"), prio)...)
			}
			if weight, err := strconv.ParseInt(srvParts[1], 10, 64); err == nil {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("weight"), weight)...)
			}
			if port, err := strconv.ParseInt(srvParts[2], 10, 64); err == nil {
				resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("port"), port)...)
			}
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("target"), srvParts[3])...)
		}
	case "TXT":
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("text"), value)...)
	case "CAA":
		// CAA format: flags tag value
		caaParts := strings.SplitN(value, " ", 3)
		if len(caaParts) >= 3 {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("flags"), caaParts[0])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tag"), caaParts[1])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("value"), caaParts[2])...)
		}
	default:
		// For other record types, try to set a generic value field if it exists
		switch recordType {
		case "ANAME":
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aname"), value)...)
		case "DNAME":
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("dname"), value)...)
		case "FWD":
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("forwarder"), value)...)
		case "URI":
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uri"), value)...)
		default:
			// For complex records or unknown types, set record_data
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("record_data"), value)...)
		}
	}

	// Set a default TTL since it's required
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ttl"), int64(3600))...)
}

// add record fields to context; export TF_LOG=debug to view
func setLogCtx(ctx context.Context, tfRec tfDNSRecord, op string) context.Context {
	logAttributes := map[string]interface{}{
		"operation":                         op,
		"zone":                              tfRec.Zone.ValueString(),
		"type":                              tfRec.Type.ValueString(),
		"domain":                            tfRec.Domain.ValueString(),
		"ttl":                               tfRec.TTL.ValueInt64(),
		"ip_address":                        tfRec.IPAddress.ValueString(),
		"ptr":                               tfRec.Ptr.ValueBool(),
		"create_ptr_zone":                   tfRec.CreatePtrZone.ValueBool(),
		"update_svcb_hints":                 tfRec.UpdateSvcbHints.ValueBool(),
		"name_server":                       tfRec.NameServer.ValueString(),
		"glue":                              tfRec.Glue.ValueString(),
		"cname":                             tfRec.CName.ValueString(),
		"ptr_name":                          tfRec.PtrName.ValueString(),
		"exchange":                          tfRec.Exchange.ValueString(),
		"preference":                        tfRec.Preference.ValueInt64(),
		"text":                              tfRec.Text.ValueString(),
		"split_text":                        tfRec.SplitText.ValueBool(),
		"mailbox":                           tfRec.Mailbox.ValueString(),
		"txt_domain":                        tfRec.TxtDomain.ValueString(),
		"priority":                          tfRec.Priority.ValueInt64(),
		"weight":                            tfRec.Weight.ValueInt64(),
		"port":                              tfRec.Port.ValueInt64(),
		"target":                            tfRec.Target.ValueString(),
		"naptr_order":                       tfRec.NaptrOrder.ValueInt64(),
		"naptr_preference":                  tfRec.NaptrPreference.ValueInt64(),
		"naptr_flags":                       tfRec.NaptrFlags.ValueString(),
		"naptr_services":                    tfRec.NaptrServices.ValueString(),
		"naptr_regexp":                      tfRec.NaptrRegexp.ValueString(),
		"naptr_replacement":                 tfRec.NaptrReplacement.ValueString(),
		"dname":                             tfRec.DName.ValueString(),
		"key_tag":                           tfRec.KeyTag.ValueInt64(),
		"algorithm":                         tfRec.Algorithm.ValueString(),
		"digest_type":                       tfRec.DigestType.ValueString(),
		"digest":                            tfRec.Digest.ValueString(),
		"sshfp_algorithm":                   tfRec.SshfpAlgorithm.ValueString(),
		"sshfp_fingerprint_type":            tfRec.SshfpFingerprintType.ValueString(),
		"sshfp_fingerprint":                 tfRec.SshfpFingerprint.ValueString(),
		"tlsa_certificate_usage":            tfRec.TlsaCertificateUsage.ValueString(),
		"tlsa_selector":                     tfRec.TlsaSelector.ValueString(),
		"tlsa_matching_type":                tfRec.TlsaMatchingType.ValueString(),
		"tlsa_certificate_association_data": tfRec.TlsaCertificateAssociationData.ValueString(),
		"svc_priority":                      tfRec.SvcPriority.ValueInt64(),
		"svc_target_name":                   tfRec.SvcTargetName.ValueString(),
		"svc_params":                        tfRec.SvcParams.ValueString(),
		"auto_ipv4_hint":                    tfRec.AutoIpv4Hint.ValueBool(),
		"auto_ipv6_hint":                    tfRec.AutoIpv6Hint.ValueBool(),
		"uri_priority":                      tfRec.UriPriority.ValueInt64(),
		"uri_weight":                        tfRec.UriWeight.ValueInt64(),
		"uri":                               tfRec.Uri.ValueString(),
		"flags":                             tfRec.Flags.ValueString(),
		"tag":                               tfRec.Tag.ValueString(),
		"value":                             tfRec.Value.ValueString(),
		"aname":                             tfRec.AName.ValueString(),
		"forwarder":                         tfRec.Forwarder.ValueString(),
		"forwarder_priority":                tfRec.ForwarderPriority.ValueInt64(),
		"dnssec_validation":                 tfRec.DnssecValidation.ValueBool(),
		"proxy_type":                        tfRec.ProxyType.ValueString(),
		"proxy_address":                     tfRec.ProxyAddress.ValueString(),
		"proxy_port":                        tfRec.ProxyPort.ValueInt64(),
		"proxy_username":                    tfRec.ProxyUsername.ValueString(),
		"proxy_password":                    tfRec.ProxyPassword.ValueString(),
		"app_name":                          tfRec.AppName.ValueString(),
		"class_path":                        tfRec.ClassPath.ValueString(),
		"record_data":                       tfRec.RecordData.ValueString(),
	}

	for k, v := range logAttributes {
		if v != nil && v != "" {
			ctx = tflog.SetField(ctx, k, v)
		}
	}

	return ctx
}

// convert from terraform data model into api data model
func tf2model(tfData tfDNSRecord) model.DNSRecord {
	return model.DNSRecord{
		Type:                           model.DNSRecordType(tfData.Type.ValueString()),
		Domain:                         model.DNSRecordName(tfData.Domain.ValueString()),
		TTL:                            model.DNSRecordTTL(tfData.TTL.ValueInt64()),
		IPAddress:                      tfData.IPAddress.ValueString(),
		Ptr:                            tfData.Ptr.ValueBool(),
		CreatePtrZone:                  tfData.CreatePtrZone.ValueBool(),
		UpdateSvcbHints:                tfData.UpdateSvcbHints.ValueBool(),
		NameServer:                     tfData.NameServer.ValueString(),
		Glue:                           tfData.Glue.ValueString(),
		CName:                          tfData.CName.ValueString(),
		PtrName:                        tfData.PtrName.ValueString(),
		Exchange:                       tfData.Exchange.ValueString(),
		Preference:                     model.DNSRecordPrio(tfData.Preference.ValueInt64()),
		Text:                           tfData.Text.ValueString(),
		SplitText:                      tfData.SplitText.ValueBool(),
		Mailbox:                        tfData.Mailbox.ValueString(),
		TxtDomain:                      tfData.TxtDomain.ValueString(),
		Priority:                       model.DNSRecordPrio(tfData.Priority.ValueInt64()),
		Weight:                         model.DNSRecordSRVWeight(tfData.Weight.ValueInt64()),
		Port:                           model.DNSRecordSRVPort(tfData.Port.ValueInt64()),
		Target:                         model.DNSRecordSRVService(tfData.Target.ValueString()),
		NaptrOrder:                     uint16(tfData.NaptrOrder.ValueInt64()),
		NaptrPreference:                uint16(tfData.NaptrPreference.ValueInt64()),
		NaptrFlags:                     tfData.NaptrFlags.ValueString(),
		NaptrServices:                  tfData.NaptrServices.ValueString(),
		NaptrRegexp:                    tfData.NaptrRegexp.ValueString(),
		NaptrReplacement:               tfData.NaptrReplacement.ValueString(),
		DName:                          tfData.DName.ValueString(),
		KeyTag:                         uint16(tfData.KeyTag.ValueInt64()),
		Algorithm:                      tfData.Algorithm.ValueString(),
		DigestType:                     tfData.DigestType.ValueString(),
		Digest:                         tfData.Digest.ValueString(),
		SshfpAlgorithm:                 tfData.SshfpAlgorithm.ValueString(),
		SshfpFingerprintType:           tfData.SshfpFingerprintType.ValueString(),
		SshfpFingerprint:               tfData.SshfpFingerprint.ValueString(),
		TlsaCertificateUsage:           tfData.TlsaCertificateUsage.ValueString(),
		TlsaSelector:                   tfData.TlsaSelector.ValueString(),
		TlsaMatchingType:               tfData.TlsaMatchingType.ValueString(),
		TlsaCertificateAssociationData: tfData.TlsaCertificateAssociationData.ValueString(),
		SvcPriority:                    uint16(tfData.SvcPriority.ValueInt64()),
		SvcTargetName:                  tfData.SvcTargetName.ValueString(),
		SvcParams:                      tfData.SvcParams.ValueString(),
		AutoIpv4Hint:                   tfData.AutoIpv4Hint.ValueBool(),
		AutoIpv6Hint:                   tfData.AutoIpv6Hint.ValueBool(),
		UriPriority:                    uint16(tfData.UriPriority.ValueInt64()),
		UriWeight:                      uint16(tfData.UriWeight.ValueInt64()),
		Uri:                            tfData.Uri.ValueString(),
		Flags:                          tfData.Flags.ValueString(),
		Tag:                            tfData.Tag.ValueString(),
		Value:                          tfData.Value.ValueString(),
		AName:                          tfData.AName.ValueString(),
		Forwarder:                      tfData.Forwarder.ValueString(),
		ForwarderPriority:              uint16(tfData.ForwarderPriority.ValueInt64()),
		DnssecValidation:               tfData.DnssecValidation.ValueBool(),
		ProxyType:                      tfData.ProxyType.ValueString(),
		ProxyAddress:                   tfData.ProxyAddress.ValueString(),
		ProxyPort:                      uint16(tfData.ProxyPort.ValueInt64()),
		ProxyUsername:                  tfData.ProxyUsername.ValueString(),
		ProxyPassword:                  tfData.ProxyPassword.ValueString(),
		AppName:                        tfData.AppName.ValueString(),
		ClassPath:                      tfData.ClassPath.ValueString(),
		RecordData:                     tfData.RecordData.ValueString(),
	}
}

// convert from api data model into terraform data model
func model2tf(apiData model.DNSRecord, tfData *tfDNSRecord) {
	if apiData.Type != "" {
		tfData.Type = types.StringValue(string(apiData.Type))
	}
	if apiData.Domain != "" {
		tfData.Domain = types.StringValue(string(apiData.Domain))
	}
	if apiData.TTL != 0 {
		tfData.TTL = types.Int64Value(int64(apiData.TTL))
	}
	if apiData.IPAddress != "" {
		tfData.IPAddress = types.StringValue(apiData.IPAddress)
	}
	if apiData.Value != "" {
		tfData.Value = types.StringValue(apiData.Value)
	}
	if apiData.Ptr {
		tfData.Ptr = types.BoolValue(apiData.Ptr)
	}
	if apiData.CreatePtrZone {
		tfData.CreatePtrZone = types.BoolValue(apiData.CreatePtrZone)
	}
	if apiData.UpdateSvcbHints {
		tfData.UpdateSvcbHints = types.BoolValue(apiData.UpdateSvcbHints)
	}
	if apiData.NameServer != "" {
		tfData.NameServer = types.StringValue(apiData.NameServer)
	}
	if apiData.Glue != "" {
		tfData.Glue = types.StringValue(apiData.Glue)
	}
	if apiData.CName != "" {
		tfData.CName = types.StringValue(apiData.CName)
	}
	if apiData.PtrName != "" {
		tfData.PtrName = types.StringValue(apiData.PtrName)
	}
	if apiData.Exchange != "" {
		tfData.Exchange = types.StringValue(apiData.Exchange)
	}
	if apiData.Preference != 0 {
		tfData.Preference = types.Int64Value(int64(apiData.Preference))
	}
	if apiData.Text != "" {
		tfData.Text = types.StringValue(apiData.Text)
	}
	if apiData.SplitText {
		tfData.SplitText = types.BoolValue(apiData.SplitText)
	}
	if apiData.Mailbox != "" {
		tfData.Mailbox = types.StringValue(apiData.Mailbox)
	}
	if apiData.TxtDomain != "" {
		tfData.TxtDomain = types.StringValue(apiData.TxtDomain)
	}
	if apiData.Priority != 0 {
		tfData.Priority = types.Int64Value(int64(apiData.Priority))
	}
	if apiData.Weight != 0 {
		tfData.Weight = types.Int64Value(int64(apiData.Weight))
	}
	if apiData.Port != 0 {
		tfData.Port = types.Int64Value(int64(apiData.Port))
	}
	if apiData.Target != "" {
		tfData.Target = types.StringValue(string(apiData.Target))
	}
	if apiData.NaptrOrder != 0 {
		tfData.NaptrOrder = types.Int64Value(int64(apiData.NaptrOrder))
	}
	if apiData.NaptrPreference != 0 {
		tfData.NaptrPreference = types.Int64Value(int64(apiData.NaptrPreference))
	}
	if apiData.NaptrFlags != "" {
		tfData.NaptrFlags = types.StringValue(apiData.NaptrFlags)
	}
	if apiData.NaptrServices != "" {
		tfData.NaptrServices = types.StringValue(apiData.NaptrServices)
	}
	if apiData.NaptrRegexp != "" {
		tfData.NaptrRegexp = types.StringValue(apiData.NaptrRegexp)
	}
	if apiData.NaptrReplacement != "" {
		tfData.NaptrReplacement = types.StringValue(apiData.NaptrReplacement)
	}
	if apiData.DName != "" {
		tfData.DName = types.StringValue(apiData.DName)
	}
	if apiData.KeyTag != 0 {
		tfData.KeyTag = types.Int64Value(int64(apiData.KeyTag))
	}
	if apiData.Algorithm != "" {
		tfData.Algorithm = types.StringValue(apiData.Algorithm)
	}
	if apiData.DigestType != "" {
		tfData.DigestType = types.StringValue(apiData.DigestType)
	}
	if apiData.Digest != "" {
		tfData.Digest = types.StringValue(apiData.Digest)
	}
	if apiData.SshfpAlgorithm != "" {
		tfData.SshfpAlgorithm = types.StringValue(apiData.SshfpAlgorithm)
	}
	if apiData.SshfpFingerprintType != "" {
		tfData.SshfpFingerprintType = types.StringValue(apiData.SshfpFingerprintType)
	}
	if apiData.SshfpFingerprint != "" {
		tfData.SshfpFingerprint = types.StringValue(apiData.SshfpFingerprint)
	}
	if apiData.TlsaCertificateUsage != "" {
		tfData.TlsaCertificateUsage = types.StringValue(apiData.TlsaCertificateUsage)
	}
	if apiData.TlsaSelector != "" {
		tfData.TlsaSelector = types.StringValue(apiData.TlsaSelector)
	}
	if apiData.TlsaMatchingType != "" {
		tfData.TlsaMatchingType = types.StringValue(apiData.TlsaMatchingType)
	}
	if apiData.TlsaCertificateAssociationData != "" {
		tfData.TlsaCertificateAssociationData = types.StringValue(apiData.TlsaCertificateAssociationData)
	}
	if apiData.SvcPriority != 0 {
		tfData.SvcPriority = types.Int64Value(int64(apiData.SvcPriority))
	}
	if apiData.SvcTargetName != "" {
		tfData.SvcTargetName = types.StringValue(apiData.SvcTargetName)
	}
	if apiData.SvcParams != "" {
		tfData.SvcParams = types.StringValue(apiData.SvcParams)
	}
	if apiData.AutoIpv4Hint {
		tfData.AutoIpv4Hint = types.BoolValue(apiData.AutoIpv4Hint)
	}
	if apiData.AutoIpv6Hint {
		tfData.AutoIpv6Hint = types.BoolValue(apiData.AutoIpv6Hint)
	}
	if apiData.UriPriority != 0 {
		tfData.UriPriority = types.Int64Value(int64(apiData.UriPriority))
	}
	if apiData.UriWeight != 0 {
		tfData.UriWeight = types.Int64Value(int64(apiData.UriWeight))
	}
	if apiData.Uri != "" {
		tfData.Uri = types.StringValue(apiData.Uri)
	}
	if apiData.Flags != "" {
		tfData.Flags = types.StringValue(apiData.Flags)
	}
	if apiData.Tag != "" {
		tfData.Tag = types.StringValue(apiData.Tag)
	}
	if apiData.Value != "" {
		tfData.Value = types.StringValue(apiData.Value)
	}
	if apiData.AName != "" {
		tfData.AName = types.StringValue(apiData.AName)
	}
	if apiData.Forwarder != "" {
		tfData.Forwarder = types.StringValue(apiData.Forwarder)
	}
	if apiData.ForwarderPriority != 0 {
		tfData.ForwarderPriority = types.Int64Value(int64(apiData.ForwarderPriority))
	}
	if apiData.DnssecValidation {
		tfData.DnssecValidation = types.BoolValue(apiData.DnssecValidation)
	}
	if apiData.ProxyType != "" {
		tfData.ProxyType = types.StringValue(apiData.ProxyType)
	}
	if apiData.ProxyAddress != "" {
		tfData.ProxyAddress = types.StringValue(apiData.ProxyAddress)
	}
	if apiData.ProxyPort != 0 {
		tfData.ProxyPort = types.Int64Value(int64(apiData.ProxyPort))
	}
	if apiData.ProxyUsername != "" {
		tfData.ProxyUsername = types.StringValue(apiData.ProxyUsername)
	}
	if apiData.ProxyPassword != "" {
		tfData.ProxyPassword = types.StringValue(apiData.ProxyPassword)
	}
	if apiData.AppName != "" {
		tfData.AppName = types.StringValue(apiData.AppName)
	}
	if apiData.ClassPath != "" {
		tfData.ClassPath = types.StringValue(apiData.ClassPath)
	}
	if apiData.RecordData != "" {
		tfData.RecordData = types.StringValue(apiData.RecordData)
	}
}
