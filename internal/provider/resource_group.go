package provider

import (
	"context"
	"fmt"
	"slices"

	"github.com/celest-dev/terraform-provider-turso/internal/resource_group"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	*tursoProviderConfig
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_group.GroupResourceSchema(ctx)
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

	client, ok := req.ProviderData.(*tursoProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tursoProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.tursoProviderConfig = client
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_group.GroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var rawLocations basetypes.SetValue
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("locations"), &rawLocations)...)
	if resp.Diagnostics.HasError() {
		return
	}
	locations := decodeStringSet(rawLocations)

	primaryLocation := data.Primary.ValueString()
	var extensions tursoclient.OptExtensions
	if !data.Extensions.IsNull() && !data.Extensions.IsUnknown() {
		extensions = tursoclient.NewOptExtensions(tursoclient.Extensions(data.Extensions.ValueString()))
	}
	input := &tursoclient.NewGroup{
		Name:       data.Name.ValueString(),
		Location:   primaryLocation,
		Extensions: extensions,
	}
	fmt.Printf("creating group: %+v\n", input)
	res, err := r.Client.CreateGroup(ctx, input, tursoclient.CreateGroupParams{
		OrganizationName: r.Organization,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err.Error()))
		return
	}
	groupData, ok := res.(*tursoclient.CreateGroupOK)
	if !ok {
		resp.Diagnostics.AddError("Client Error", "Unable to create group, got unexpected response")
		return
	}
	group := groupData.Group.Value
	fmt.Printf("created group: %+v\n", group)

	fmt.Printf("adding locations: %s\n", locations)
	for _, location := range locations {
		if location == primaryLocation {
			continue
		}
		res, err := r.Client.AddLocationToGroup(ctx, tursoclient.AddLocationToGroupParams{
			OrganizationName: r.Organization,
			GroupName:        group.Name.Value,
			Location:         location,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add location to group, got error: %s", err.Error()))
			return
		}
		if _, ok := res.(*tursoclient.AddLocationToGroupOK); !ok {
			resp.Diagnostics.AddError("Client Error", "Unable to add location to group, got unexpected response")
			return
		}
	}

	tflog.Trace(ctx, "created group resource")
	resp.Diagnostics.Append(r.readGroupResource(ctx, group.Name.Value, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_group.GroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readGroupResource(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_group.GroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var curr resource_group.GroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &curr)...)
	if resp.Diagnostics.HasError() {
		return
	}

	currentLocations := decodeStringSet(curr.Locations)
	requestedLocations := decodeStringSet(data.Locations)

	addLocations := make([]string, 0, len(requestedLocations))
	for _, location := range requestedLocations {
		if !slices.Contains(currentLocations, location) {
			addLocations = append(addLocations, location)
		}
	}
	fmt.Printf("adding locations: %+v\n", addLocations)
	for _, location := range addLocations {
		res, err := r.Client.AddLocationToGroup(ctx, tursoclient.AddLocationToGroupParams{
			OrganizationName: r.Organization,
			GroupName:        data.Name.ValueString(),
			Location:         location,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add location to group, got error: %s", err.Error()))
			return
		}
		if _, ok := res.(*tursoclient.AddLocationToGroupOK); !ok {
			resp.Diagnostics.AddError("Client Error", "Unable to add location to group, got unexpected response")
			return
		}
	}

	removeLocations := make([]string, 0, len(currentLocations))
	for _, location := range currentLocations {
		if !slices.Contains(requestedLocations, location) {
			removeLocations = append(removeLocations, location)
		}
	}
	fmt.Printf("removing locations: %+v\n", removeLocations)
	for _, location := range removeLocations {
		res, err := r.Client.RemoveLocationFromGroup(ctx, tursoclient.RemoveLocationFromGroupParams{
			OrganizationName: r.Organization,
			GroupName:        data.Name.ValueString(),
			Location:         location,
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove location from group, got error: %s", err.Error()))
			return
		}
		if _, ok := res.(*tursoclient.RemoveLocationFromGroupOK); !ok {
			resp.Diagnostics.AddError("Client Error", "Unable to remove location from group, got unexpected response")
			return
		}
	}

	resp.Diagnostics.Append(r.readGroupResource(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_group.GroupModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Printf("deleting group: %+v\n", data)
	_, err := r.Client.DeleteGroup(ctx, tursoclient.DeleteGroupParams{
		OrganizationName: r.Organization,
		GroupName:        data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err.Error()))
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fmt.Printf("importing group: %+v\n", req)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *tursoProviderConfig) readGroup(ctx context.Context, name string) (tursoclient.BaseGroup, diag.Diagnostics) {
	resp, err := r.Client.GetGroup(ctx, tursoclient.GetGroupParams{
		OrganizationName: r.Organization,
		GroupName:        name,
	})
	if err != nil {
		return tursoclient.BaseGroup{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", err.Error()),
		}
	}
	groupData, ok := resp.(*tursoclient.GetGroupOK)
	if !ok {
		return tursoclient.BaseGroup{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", "group not returned from server"),
		}
	}
	group := groupData.Group.Value
	fmt.Printf("read group: %+v\n", group)
	return group, nil
}

func (r *GroupResource) readGroupResource(ctx context.Context, name string, data *resource_group.GroupModel) diag.Diagnostics {
	group, diags := r.readGroup(ctx, name)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(group.Name.Value)
	data.Name = types.StringValue(group.Name.Value)
	data.Primary = types.StringValue(group.Primary.Value)
	if data.Extensions.IsUnknown() {
		data.Extensions = types.StringNull()
	}

	locations := encodeStringSet(mergeLists(group.Locations, []string{group.Primary.Value}))
	data.Locations = locations
	data.Group, diags = resource_group.NewGroupValue(resource_group.GroupValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"archived":  types.BoolValue(group.Archived.Value),
		"name":      types.StringValue(group.Name.Value),
		"primary":   types.StringValue(group.Primary.Value),
		"uuid":      types.StringValue(group.UUID.Value),
		"version":   types.StringValue(group.Version.Value),
		"locations": locations,
	})
	return diags
}
