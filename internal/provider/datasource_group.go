package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_group"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSourceWithConfigure = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

type GroupDataSource struct {
	*tursoProviderConfig
}

func (r *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_group.GroupDataSourceSchema(ctx)
}

func (r *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_group.GroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.readGroupDataSource(ctx, data.Id.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupDataSource) readGroupDataSource(ctx context.Context, name string, data *datasource_group.GroupModel) diag.Diagnostics {
	group, diags := r.readGroup(ctx, name)
	if diags.HasError() {
		return diags
	}

	locations := encodeStringList(mergeLists(group.Locations, []string{group.Primary.Value}))
	data.Group, diags = datasource_group.NewGroupValue(datasource_group.GroupValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"archived":  types.BoolValue(group.Archived.Value),
		"name":      types.StringValue(group.Name.Value),
		"primary":   types.StringValue(group.Primary.Value),
		"uuid":      types.StringValue(group.UUID.Value),
		"version":   types.StringValue(group.Version.Value),
		"locations": locations,
	})

	return diags
}
