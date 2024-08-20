package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}
var _ resource.ResourceWithConfigValidators = &GroupResource{}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

type GroupResource struct {
	client *tursoadmin.Client
}

type GroupResourceModel struct {
	// The group name, unique across your organization.
	Name types.String `tfsdk:"name"`

	// The primary location key.
	Primary types.String `tfsdk:"primary"`

	// The locations where the databases in the group are running.
	Locations []types.String `tfsdk:"locations"`

	// Computed

	// The current libSQL server version the databases in that group are running.
	Version types.String `tfsdk:"version"`

	// The group universal unique identifier (UUID).
	UUID types.String `tfsdk:"uuid"`

	// Groups on the free tier go to sleep after some inactivity.
	Archived types.Bool `tfsdk:"archived"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Turso group resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"primary": schema.StringAttribute{
				MarkdownDescription: "The primary location of the group. Required if multiple `locations` are specified.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locations": schema.SetAttribute{
				MarkdownDescription: "The locations of the group.",
				Required:            true,
				ElementType:         basetypes.StringType{},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The current libSQL server version the databases in that group are running.",
				Computed:            true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "The group universal unique identifier (UUID).",
				Computed:            true,
			},
			"archived": schema.BoolAttribute{
				MarkdownDescription: "Groups on the free tier go to sleep after some inactivity.",
				Computed:            true,
			},
		},
	}
}

type groupConfigValidator struct{}

var _ resource.ConfigValidator = &groupConfigValidator{}

// Description implements resource.ConfigValidator.
func (p *groupConfigValidator) Description(context.Context) string {
	return "Validate the group configuration."
}

// MarkdownDescription implements resource.ConfigValidator.
func (p *groupConfigValidator) MarkdownDescription(context.Context) string {
	return "Validate the group configuration."
}

// ValidateResource implements resource.ConfigValidator.
func (p *groupConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var locations types.Set
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("locations"), &locations)...)

	var primary types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("primary"), &primary)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if len(locations.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(path.Root("locations"), "Invalid locations", "At least one location must be specified.")
		return
	}

	if len(locations.Elements()) == 1 {
		// OK, will use the only location as primary
		return
	}

	if primary.IsNull() {
		resp.Diagnostics.AddAttributeError(path.Root("primary"), "Invalid primary location", "Primary location must be specified if multiple locations are specified.")
		return
	}
}

func (r *GroupResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&groupConfigValidator{},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	locations := make([]string, len(data.Locations))
	for i, location := range data.Locations {
		locations[i] = location.ValueString()
	}

	var primaryLocation string
	if !data.Primary.IsNull() && !data.Primary.IsUnknown() {
		primaryLocation = data.Primary.ValueString()
	} else {

		primaryLocation = locations[0]
	}

	group, err := r.client.CreateGroup(ctx, tursoadmin.CreateGroupRequest{
		Name:       data.Name.ValueString(),
		Location:   primaryLocation,
		Extensions: tursoadmin.ExtensionsAll,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err.Error()))
		return
	}

	for _, location := range locations {
		if location == primaryLocation {
			continue
		}
		group, err = r.client.AddGroupLocation(ctx, data.Name.ValueString(), location)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add location to group, got error: %s", err.Error()))
			return
		}
	}

	data.Name = types.StringValue(group.Name)
	data.Primary = types.StringValue(group.PrimaryLocation)
	data.Locations = make([]basetypes.StringValue, len(group.Locations))
	for i, location := range group.Locations {
		data.Locations[i] = types.StringValue(location)
	}
	data.Version = types.StringValue(group.Version)
	data.UUID = types.StringValue(group.UUID)
	data.Archived = types.BoolValue(group.Archived)

	tflog.Trace(ctx, "created group resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err.Error()))
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Primary = types.StringValue(group.PrimaryLocation)
	data.Locations = make([]basetypes.StringValue, len(group.Locations))
	for i, location := range group.Locations {
		data.Locations[i] = types.StringValue(location)
	}
	data.Version = types.StringValue(group.Version)
	data.UUID = types.StringValue(group.UUID)
	data.Archived = types.BoolValue(group.Archived)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err.Error()))
		return
	}

	currentLocations := group.Locations
	requestedLocations := make([]string, len(data.Locations))
	for i, location := range data.Locations {
		requestedLocations[i] = location.ValueString()
	}

	missingLocations := make([]string, 0, len(requestedLocations))
	for _, location := range requestedLocations {
		if !slices.Contains(currentLocations, location) {
			missingLocations = append(missingLocations, location)
		}
	}

	for _, location := range missingLocations {
		group, err = r.client.AddGroupLocation(ctx, data.Name.ValueString(), location)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add location to group, got error: %s", err.Error()))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGroup(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err.Error()))
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
