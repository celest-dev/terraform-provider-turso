package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_databases"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabasesDataSource{}

func NewDatabasesDataSource() datasource.DataSource {
	return &DatabasesDataSource{}
}

// DatabasesDataSource defines the data source implementation.
type DatabasesDataSource struct {
	*tursoProviderConfig
}

func (d *DatabasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *DatabasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_databases.DatabasesDataSourceSchema(ctx)
}

func (d *DatabasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*tursoProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *tursoProviderConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.tursoProviderConfig = config
}

func (d *DatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_databases.DatabasesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.readDatabasesDataSource(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabasesDataSource) readDatabasesDataSource(ctx context.Context, data *datasource_databases.DatabasesModel) diag.Diagnostics {
	diags := diag.Diagnostics{}
	res, err := r.Client.ListDatabases(ctx, tursoclient.ListDatabasesParams{
		OrganizationName: r.Organization,
	})
	if err != nil {
		diags.AddError("Failed to list databases", err.Error())
		return diags
	}

	databases := make([]attr.Value, len(res.GetDatabases()))
	for i, db := range res.GetDatabases() {
		databases[i], diags = datasource_databases.NewDatabasesValue(
			datasource_databases.DatabasesValue{}.AttributeTypes(ctx),
			map[string]attr.Value{
				"db_id":          types.StringValue(db.DbId.Value),
				"name":           types.StringValue(db.Name.Value),
				"group":          types.StringValue(db.Group.Value),
				"hostname":       types.StringValue(db.Hostname.Value),
				"regions":        encodeStringList(db.Regions),
				"primary_region": types.StringValue(db.PrimaryRegion.Value),
				"schema":         types.StringValue(db.Schema.Value),
				"is_schema":      types.BoolValue(db.IsSchema.Value),
				"type":           types.StringValue(db.Type.Value),
				"archived":       types.BoolValue(db.Archived.Value),
				"version":        types.StringValue(db.Version.Value),
				"allow_attach":   types.BoolValue(db.AllowAttach.Value),
				"block_reads":    types.BoolValue(db.BlockReads.Value),
				"block_writes":   types.BoolValue(db.BlockWrites.Value),
			},
		)
		if diags.HasError() {
			return diags
		}
	}

	data.Databases, diags = types.ListValue(datasource_databases.DatabasesType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_databases.DatabasesValue{}.AttributeTypes(ctx),
		},
	}, databases)
	return diags
}
