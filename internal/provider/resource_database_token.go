package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &DatabaseTokenResource{}
var _ resource.ResourceWithConfigure = &DatabaseTokenResource{}

func NewDatabaseTokenResource() resource.Resource {
	return &DatabaseTokenResource{}
}

type DatabaseTokenResource struct {
	client *tursoadmin.Client
}

type DatabaseTokenResourceModel struct {
	DatabaseName types.String         `tfsdk:"database"`
	Expiration   timetypes.GoDuration `tfsdk:"expiration"`

	// Computed
	Token     types.String      `tfsdk:"token"`
	ExpiresAt timetypes.RFC3339 `tfsdk:"expires_at"`
}

func (r *DatabaseTokenResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_token"
}

func (r *DatabaseTokenResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database token resource",
		Attributes: map[string]schema.Attribute{
			"database": schema.StringAttribute{
				MarkdownDescription: "The name of the database.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration": schema.StringAttribute{
				MarkdownDescription: `
Expiration time for the token. If not provided, defaults to "never".

A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, 
such as "300s", "-1.5h" or "2h45m". 

Valid time units are "s", "m", "h"."`,
				Optional:   true,
				CustomType: timetypes.GoDurationType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The database authorization token (JWT).",
				Computed:            true,
				Sensitive:           true,
			},
			"expires_at": schema.StringAttribute{
				MarkdownDescription: "The computed expiration time of the token.",
				Computed:            true,
				CustomType:          timetypes.RFC3339Type{},
			},
		},
	}
}

func (r *DatabaseTokenResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tursoadmin.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DatabaseTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseTokenResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	now := time.Now().UTC()
	var expiration time.Duration
	if !data.Expiration.IsUnknown() && !data.Expiration.IsNull() {
		exp, diags := data.Expiration.ValueGoDuration()
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}
		expiration = exp
	} else {
		data.Expiration = timetypes.NewGoDurationNull()
	}

	res, err := r.client.CreateDatabaseToken(ctx, data.DatabaseName.ValueString(), expiration)
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
	data.Token = types.StringValue(res)
	if expiration != time.Duration(0) {
		data.ExpiresAt = timetypes.NewRFC3339TimeValue(now.Add(expiration))
	} else {
		data.ExpiresAt = timetypes.NewRFC3339Null()
	}

	tflog.Trace(ctx, "created database token")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	panic("should have forced replacement")
}

func (r *DatabaseTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Technically we can only invalidate ALL database tokens, but because we expect only one token to exist per database at a time
	// with no expiration, this is acceptable.
	fmt.Printf("Invalidating all database tokens for %s\n", data.DatabaseName.ValueString())
	err := r.client.InvalidateDatabaseTokens(ctx, data.DatabaseName.ValueString())
	if err != nil {
		// TODO: Invalidating the db token is not allowed when the DB is in a group ??
		resp.Diagnostics.AddWarning("client error", err.Error())
		return
	}

	tflog.Trace(ctx, "deleted group token")
}
