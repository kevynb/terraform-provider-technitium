package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"
	"github.com/veksh/terraform-provider-godaddy-dns/internal/model"
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
	Domain   types.String `tfsdk:"domain"`
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Data     types.String `tfsdk:"data"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Priority types.Int64  `tfsdk:"priority"`
}

// add record fields to context; export TF_LOG=debug to view
func setLogCtx(ctx context.Context, tfRec tfDNSRecord, op string) context.Context {
	ctx = tflog.SetField(ctx, "domain", tfRec.Domain.ValueString())
	ctx = tflog.SetField(ctx, "type", tfRec.Type.ValueString())
	ctx = tflog.SetField(ctx, "name", tfRec.Name.ValueString())
	ctx = tflog.SetField(ctx, "data", tfRec.Data.ValueString())
	ctx = tflog.SetField(ctx, "operation", op)
	return ctx
}

// convert from terraform data model into api data model
func tf2model(tfData tfDNSRecord) (model.DNSDomain, model.DNSRecord) {
	return model.DNSDomain(tfData.Domain.ValueString()),
		model.DNSRecord{
			Name:     model.DNSRecordName(tfData.Name.ValueString()),
			Type:     model.DNSRecordType(tfData.Type.ValueString()),
			Data:     model.DNSRecordData(tfData.Data.ValueString()),
			TTL:      model.DNSRecordTTL(tfData.TTL.ValueInt64()),
			Priority: model.DNSRecordPrio(tfData.Priority.ValueInt64()),
		}
}

// RecordResource defines the implementation of GoDaddy DNS RR
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
		MarkdownDescription: "DNS resource record represens a single RR in managed domain",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Name of main managed domain (top-level) for this RR",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Resource record type: A, CNAME etc",
				Required:            true,
				Validators: []validator.String{
					// TODO: SRV management
					// TODO: custom validator to require "priority" for type == MX
					stringvalidator.Any(
						// attempt to require priority only for MX: error message is not quite clear :)
						stringvalidator.OneOf([]string{"A", "AAAA", "CNAME", "NS", "TXT"}...),
						stringvalidator.All(
							// mx requires priority
							stringvalidator.OneOf([]string{"MX"}...),
							stringvalidator.AlsoRequires(path.Expressions{
								path.MatchRoot("priority"),
							}...),
						),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Record name name (part of FQN), may include `.` for records in sub-domains or be `@` for top-level records",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Record value returned for DNS query: target for CNAME, ip address for A etc",
				Required:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Record time-to-live, >= 600s <= 604800s (1 week), default 3600 seconds (1 hour)",
				Optional:            true,
				Computed:            true, // must be computed to use a default
				Default:             int64default.StaticInt64(3600),
				Validators: []validator.Int64{
					int64validator.AtLeast(600),
					int64validator.AtMost(604800),
				},
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Record priority, required for MX (lower is higher)",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
					int64validator.AtMost(1023),
				},
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

	apiDomain, apiRecPlan := tf2model(planData)
	// "put"/"add" does not check prior state (terraform does not provide one for Create)
	// and so will fail on uniqueness violation (e.g. if record already exists
	// after external modification, or if it is the second CNAME RR etc)
	// - lets think it is ok for now -- let API do checking + run "import" if required
	// - alt/TODO: read records and do noop if target record is already there
	//   like `apiAllRecs, err := r.client.GetRecords(ctx, apiDomain, apiRecPlan.Type, apiRecPlan.Name)`
	//   but lets not be silent about that
	err := r.client.AddRecords(ctx, apiDomain, []model.DNSRecord{apiRecPlan})

	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create record: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &planData)...)
}

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

	apiDomain, apiRecState := tf2model(stateData)

	apiAllRecs, err := r.client.GetRecords(ctx, apiDomain, apiRecState.Type, apiRecState.Name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Reading DNS records: query failed: %s", err))
		return
	}
	numFound := 0
	if numRecs := len(apiAllRecs); numRecs == 0 {
		tflog.Debug(ctx, "Reading DNS record: currently absent")
	} else {
		tflog.Info(ctx, fmt.Sprintf(
			"Reading DNS record: got %d answers", numRecs))
		// meaning of "match" is different between types
		//  - for CNAME (and SOA), there could be only 1 records with a given name
		//    in a (sub-)domain
		//  - for A, TXT, MX and NS there could be several, have to match by data
		//    - MXes could have different priorities; in theory, MX 0 and MX 10
		//      could point to the same "data", but lets think that it is a
		//      preversion and replace it with one :)
		//    - TXT and NS for same name could differ only in TTL
		//  - for SRV PK is proto+service+port+data, value is weight+prio+ttl
		for _, rec := range apiAllRecs {
			tflog.Debug(ctx, fmt.Sprintf("Got DNS record: %v", rec))
			if rec.SameKey(apiRecState) {
				tflog.Info(ctx, "matching DNS record found")
				stateData.Data = types.StringValue(string(rec.Data))
				stateData.TTL = types.Int64Value(int64(rec.TTL))
				switch rec.Type {
				case model.REC_MX:
					stateData.Priority = types.Int64Value(int64(rec.Priority))
				case model.REC_SRV:
					stateData.Priority = types.Int64Value(int64(rec.Priority))
					// TODO: weight
				}
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

// updating will fail if resource is already changed externally: old record will be "gone"
// after refresh, so actually "create" will be called for new one (and see above): i.e.
// changing "A -> 1.1.1.1" to "A -> 2.2.2.2" first in domain and then in main.tf will
// result in an error (refresh will mark it as gone and will try to create new)
// so, do not do that :)
// the way to settle things down in this case is "refresh" (will mark old as gone)
// + "import" to new (so state will be ok)
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

	apiDomain, apiRecPlan := tf2model(planData)

	var err error
	if apiRecPlan.Type.IsSingleValue() {
		// for CNAME: just one record replacing another
		err = r.client.SetRecords(ctx,
			apiDomain, apiRecPlan.Type, apiRecPlan.Name,
			[]model.DNSUpdateRecord{{
				Data: apiRecPlan.Data,
				TTL:  apiRecPlan.TTL,
			}})
	} else {
		// for multi-valued records: copy all the rest except previous state
		var stateData tfDNSRecord
		var apiUpdateRecs []model.DNSUpdateRecord
		resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiUpdateRecs, err = r.apiRecsToKeep(ctx, stateData)
		if err != nil && err != errRecordGone {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Getting DNS records to keep failed: %s", err))
			return
		}
		// lets try to detect the situation when old record is gone and new is present
		// actually this should not happen (implicit "refresh" before "apply" will remove
		// old record from the state), but who knows :)
		oldGone := false
		if err == errRecordGone {
			tflog.Info(ctx, "Current record is already gone")
			oldGone = true
		}
		tflog.Info(ctx, fmt.Sprintf("Got %d records to keep", len(apiUpdateRecs)))
		// and finally, add our record (TODO: SRV has more fields)
		ourRec := model.DNSUpdateRecord{
			Data:     apiRecPlan.Data,
			TTL:      apiRecPlan.TTL,
			Priority: apiRecPlan.Priority,
		}
		newPresent := false
		if slices.Index(apiUpdateRecs, ourRec) >= 0 {
			// still need to delete old value if not gone
			tflog.Info(ctx, "Updated record is already present")
			newPresent = true
		} else {
			apiUpdateRecs = append(apiUpdateRecs, ourRec)
		}
		if oldGone && newPresent {
			tflog.Info(ctx, "Nothing left to do")
			err = nil
		} else {
			err = r.client.SetRecords(ctx, apiDomain, apiRecPlan.Type, apiRecPlan.Name, apiUpdateRecs)
		}
	}

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

	apiDomain, apiRecState := tf2model(stateData)

	if apiRecState.Type.IsSingleValue() {
		// for single-value types, delete is ok; multi-valued have to be replaced
		err := r.client.DelRecords(ctx, apiDomain, apiRecState.Type, apiRecState.Name)
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Deleting DNS record failed: %s", err))
			return
		}
	} else {
		// for multi-valued records: copy all the rest except previous state
		var stateData tfDNSRecord
		resp.Diagnostics.Append(req.State.Get(ctx, &stateData)...)
		if resp.Diagnostics.HasError() {
			return
		}
		apiRecsToKeep, err := r.apiRecsToKeep(ctx, stateData)
		if err != nil {
			if err == errRecordGone {
				tflog.Info(ctx, "DNS record already gone")
				return
			} else {
				resp.Diagnostics.AddError("Client Error",
					fmt.Sprintf("Getting DNS records to keep failed: %s", err))
				return
			}
		}
		tflog.Info(ctx, fmt.Sprintf("Got %d records to keep", len(apiRecsToKeep)))
		if len(apiRecsToKeep) == 0 {
			err = r.client.DelRecords(ctx, apiDomain, apiRecState.Type, apiRecState.Name)
		} else {
			err = r.client.SetRecords(ctx, apiDomain, apiRecState.Type, apiRecState.Name, apiRecsToKeep)
		}
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Replacing DNS records failed: %s", err))
			return
		}
	}
}

// terraform import godaddy-dns_record.new-cname domain:CNAME:_test:testing.com
func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// resource.ImportStatePassthroughID(ctx, path.Root("data"), req, resp)

	// for some reason Terraform does not pass schema data to Read on import
	// either as a separate structure in ReadRequest or as defaults: if only
	// they were accessible, it would eliminate the need to pass anything here

	idParts := strings.SplitN(req.ID, IMPORT_SEP, 4)

	// mb check format and emptiness
	if len(idParts) != 4 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier format: domain:TYPE:name:data"+
				"like mydom.com:CNAME:www.subdom:www.other.com. Got: %q", req.ID),
		)
		return
	}

	for i, f := range []string{"domain", "type", "name", "data"} {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(f), idParts[i])...)
	}
}

var errRecordGone = errors.New("record already gone")

// get all records for type + name, return all of them except the record
// matching stateData (it will be deleted or updated), converted to update
// format (without type and name); these are intended to be kept unchanged
// during update/delete ops on target record
func (r *RecordResource) apiRecsToKeep(ctx context.Context, stateData tfDNSRecord) ([]model.DNSUpdateRecord, error) {
	// records may differ in data or value; should be present in current API reply

	ctx = tflog.SetField(ctx, "operation", "read-keep")
	tflog.Info(ctx, "recs-to-keep: start")
	defer tflog.Info(ctx, "recs-to-keep: end")

	res := []model.DNSUpdateRecord{}
	matchesWithState := 0
	apiDomain, apiRecState := tf2model(stateData)
	apiAllRecs, err := r.client.GetRecords(ctx, apiDomain, apiRecState.Type, apiRecState.Name)
	if err != nil {
		return res, errors.Wrap(err, "Client error: query failed")
	}
	if numRecs := len(apiAllRecs); numRecs == 0 {
		// strange but quite ok for both delete (NOOP) and update (keep nothing)
		tflog.Warn(ctx, "API returned no records, will continue")
	} else {
		tflog.Debug(ctx, fmt.Sprintf("Got %d answers from API", numRecs))
		for _, rec := range apiAllRecs {
			tflog.Debug(ctx,
				fmt.Sprintf("Got DNS RR: data %s, prio %d, ttl %d", rec.Data, rec.Priority, rec.TTL))
			if rec.SameKey(apiRecState) {
				tflog.Debug(ctx, "Matching DNS record found")
				matchesWithState += 1
			} else {
				// convert to update format
				res = append(res, rec.ToUpdate())
			}
		}
		tflog.Debug(ctx, fmt.Sprintf("Found %d records to keep", len(res)))
	}
	if matchesWithState != 1 {
		tflog.Warn(ctx, fmt.Sprintf("Reading DNS records: want == 1 record, got %d", matchesWithState))
		if matchesWithState == 0 {
			return res, errRecordGone
		}
	}
	return res, nil
}
