package provider

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/oauth2"
)

// Ensure TursoProvider satisfies various provider interfaces.
var _ provider.Provider = &TursoProvider{}
var _ provider.ProviderWithFunctions = &TursoProvider{}

// TursoProvider defines the provider implementation.
type TursoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TursoProviderModel describes the provider data model.
type TursoProviderModel struct {
	Organization types.String `tfsdk:"organization"`
	ApiToken     types.String `tfsdk:"api_token"`
}

// tursoProviderConfig holds common config for the provider.
type tursoProviderConfig struct {
	Organization string
	Client       *tursoclient.Client
}

func (p *TursoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "turso"
	resp.Version = p.version
}

func (p *TursoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				MarkdownDescription: "The name of the Turso organization",
				Required:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "The API token to authenticate with Turso API. If not provided, the TURSO_API_TOKEN environment variable will be used. Finally, `turso auth token` is used to get the token.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *TursoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TursoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var apiToken string
	if !config.ApiToken.IsNull() && !config.ApiToken.IsUnknown() {
		apiToken = config.ApiToken.ValueString()
	} else if token, ok := os.LookupEnv("TURSO_API_TOKEN"); ok {
		apiToken = token
	} else {
		out, err := exec.Command("turso", "auth", "token").Output()
		if err == nil {
			apiToken = strings.TrimSpace(string(out))
		}
	}

	if apiToken == "" {
		resp.Diagnostics.AddError("api_token is required", "Must be provided in the configuration, the TURSO_API_TOKEN environment variable, or by logging into the Turso CLI.")
		return
	}

	authClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiToken}))
	client, err := tursoclient.NewClient("https://api.turso.tech", tursoclient.WithClient(authClient))
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), err.Error())
		return
	}
	providerConfig := &tursoProviderConfig{
		Organization: config.Organization.ValueString(),
		Client:       client,
	}
	resp.DataSourceData = providerConfig
	resp.ResourceData = providerConfig
}

func (p *TursoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDatabaseResource,
		NewGroupResource,
	}
}

func (p *TursoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDatabaseDataSource,
		NewDatabaseTokenDataSource,
		NewDatabaseInstancesDataSource,
		NewGroupTokenDataSource,
		NewGroupDataSource,
	}
}

func (p *TursoProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TursoProvider{
			version: version,
		}
	}
}
