package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_database_instances"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSourceWithConfigure = &DatabaseInstancesDataSource{}

func NewDatabaseInstancesDataSource() datasource.DataSource {
	return &DatabaseInstancesDataSource{}
}

type DatabaseInstancesDataSource struct {
	*tursoProviderConfig
}

func (r *DatabaseInstancesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_instances"
}

func (r *DatabaseInstancesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_database_instances.DatabaseInstancesDataSourceSchema(ctx)
}

func (r *DatabaseInstancesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *DatabaseInstancesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_database_instances.DatabaseInstancesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.Client.ListDatabaseInstances(ctx, tursoclient.ListDatabaseInstancesParams{
		OrganizationName: r.Organization,
		DatabaseName:     data.Id.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database token", err.Error())
		return
	}
	instances := make([]attr.Value, len(res.Instances))
	for i, instance := range res.Instances {
		instanceVal, diags := datasource_database_instances.NewInstancesValue(datasource_database_instances.InstancesValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"hostname": basetypes.NewStringValue(instance.Hostname.Value),
			"name":     basetypes.NewStringValue(instance.Name.Value),
			"region":   basetypes.NewStringValue(instance.Region.Value),
			"type":     basetypes.NewStringValue(string(instance.Type.Value)),
			"uuid":     basetypes.NewStringValue(instance.UUID.Value),
		})
		resp.Diagnostics.Append(diags...)
		instances[i] = instanceVal
	}
	if resp.Diagnostics.HasError() {
		return
	}

	instancesTy := datasource_database_instances.InstancesValue{}.Type(ctx)
	data.Instances = basetypes.NewListValueMust(instancesTy, instances)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
