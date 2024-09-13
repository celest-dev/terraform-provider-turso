package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_database"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

// DatabaseDataSource defines the data source implementation.
type DatabaseDataSource struct {
	*tursoProviderConfig
}

func (d *DatabaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *DatabaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_database.DatabaseDataSourceSchema(ctx)
}

func (d *DatabaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*tursoProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.tursoProviderConfig = config
}

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_database.DatabaseModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.readDatabaseDataSource(ctx, data.Id.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseDataSource) readDatabaseDataSource(ctx context.Context, name string, data *datasource_database.DatabaseModel) diag.Diagnostics {
	db, diags := r.readDatabase(ctx, name)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(db.Name.Value)
	data.Database, diags = datasource_database.NewDatabaseValue(datasource_database.DatabaseValue{}.AttributeTypes(ctx), map[string]attr.Value{
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
	})
	return diags
}
