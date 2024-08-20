package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabaseDataSource{}

func NewDatabaseDataSource() datasource.DataSource {
	return &DatabaseDataSource{}
}

// DatabaseDataSource defines the data source implementation.
type DatabaseDataSource struct {
	client *tursoadmin.Client
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the new database. Must contain only lowercase letters, numbers, dashes. No longer than 32 characters.",
				Required:            true,
			},
			"db_id": schema.StringAttribute{
				MarkdownDescription: "The database universal unique identifier (UUID).",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The DNS hostname used for client libSQL and HTTP connections.",
				Computed:            true,
			},
		},
	}
}

func (d *DatabaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tursoadmin.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DatabaseDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetDatabase(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database, got error: %s", err.Error()))
		return
	}

	data.DbId = types.StringValue(res.ID)
	data.Hostname = types.StringValue(res.Hostname)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
