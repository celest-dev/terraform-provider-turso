package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/datasource_database_token"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ datasource.DataSourceWithConfigure = &DatabaseTokenDataSource{}

func NewDatabaseTokenDataSource() datasource.DataSource {
	return &DatabaseTokenDataSource{}
}

type DatabaseTokenDataSource struct {
	*tursoProviderConfig
}

func (r *DatabaseTokenDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_token"
}

func (r *DatabaseTokenDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_database_token.DatabaseTokenDataSourceSchema(ctx)
	jwtRawAttr := resp.Schema.Attributes["jwt"]
	jwtAttr, ok := jwtRawAttr.(schema.StringAttribute)
	if !ok {
		resp.Diagnostics.AddError("Failed to set jwt attribute as sensitive", "Failed to set jwt attribute as sensitive")
		return
	}
	jwtAttr.Sensitive = true
	resp.Schema.Attributes["jwt"] = jwtAttr
}

func (r *DatabaseTokenDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tursoProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.tursoProviderConfig = client
}

func (r *DatabaseTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_database_token.DatabaseTokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var expiration tursoclient.OptString
	if !data.Expiration.IsNull() && !data.Expiration.IsUnknown() {
		expiration = tursoclient.NewOptString(data.Expiration.ValueString())
	}
	var authorization tursoclient.OptCreateDatabaseTokenAuthorization
	if !data.Authorization.IsNull() && !data.Authorization.IsUnknown() {
		authorization = tursoclient.NewOptCreateDatabaseTokenAuthorization(tursoclient.CreateDatabaseTokenAuthorization(data.Authorization.ValueString()))
	}
	token, err := r.Client.CreateDatabaseToken(ctx, tursoclient.OptCreateTokenInput{}, tursoclient.CreateDatabaseTokenParams{
		OrganizationName: r.Organization,
		DatabaseName:     data.Id.ValueString(),
		Expiration:       expiration,
		Authorization:    authorization,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database token", err.Error())
		return
	}
	switch token := token.(type) {
	case *tursoclient.CreateDatabaseTokenOK:
		data.Jwt = basetypes.NewStringValue(token.Jwt.Value)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	default:
		resp.Diagnostics.AddError("Failed to create database token", "Failed to create database token")
	}
}
