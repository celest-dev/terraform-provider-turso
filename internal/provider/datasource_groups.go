package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_groups"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &GroupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

type GroupsDataSource struct {
	*tursoProviderConfig
}

func (r *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (r *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_groups.GroupsDataSourceSchema(ctx)
}

func (r *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_groups.GroupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.readGroupsDataSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupsDataSource) readGroupsDataSource(ctx context.Context, data *datasource_groups.GroupsModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	res, err := r.Client.ListGroups(ctx, tursoclient.ListGroupsParams{
		OrganizationName: r.Organization,
	})
	if err != nil {
		diags.AddError("Failed to list groups", err.Error())
		return diags
	}

	groups := make([]attr.Value, len(res.Groups))
	for i, group := range res.Groups {
		locations := encodeStringList(mergeLists(group.Locations, []string{group.Primary.Value}))
		groups[i], diags = datasource_groups.NewGroupsValue(datasource_groups.GroupsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"archived":  types.BoolValue(group.Archived.Value),
			"name":      types.StringValue(group.Name.Value),
			"primary":   types.StringValue(group.Primary.Value),
			"uuid":      types.StringValue(group.UUID.Value),
			"version":   types.StringValue(group.Version.Value),
			"locations": locations,
		})
		if diags.HasError() {
			return diags
		}
	}

	data.Groups, diags = types.ListValue(datasource_groups.GroupsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_groups.GroupsValue{}.AttributeTypes(ctx),
		},
	}, groups)
	return diags
}
