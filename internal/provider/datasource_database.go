package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_database"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
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

// DatabaseDataSourceModel describes the data source data model.
type DatabaseDataSourceModel struct {
	Name types.String `tfsdk:"name"`

	// Computed
	DbId     types.String `tfsdk:"db_id"`
	Hostname types.String `tfsdk:"hostname"`
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
	var data DatabaseDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.Client.GetDatabase(ctx, tursoclient.GetDatabaseParams{
		OrganizationName: d.Organization,
		DatabaseName:     data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database, got error: %s", err.Error()))
		return
	}
	switch res := res.(type) {
	case *tursoclient.DatabaseNotFoundResponse:
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Database %s not found", data.Name.ValueString()))
	case *tursoclient.GetDatabaseOK:
		db, ok := res.Database.Get()
		if !ok {
			resp.Diagnostics.AddError("client error", "database not returned from server")
			return
		}
		data.DbId = types.StringValue(db.DbId.Value)
		data.Hostname = types.StringValue(db.Hostname.Value)

		// Save data into Terraform state
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}
