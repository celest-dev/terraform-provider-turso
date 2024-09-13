package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/resource_database"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	*tursoProviderConfig
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_database.DatabaseResourceSchema(ctx)
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tursoProviderConfig)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tursoadmin.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.tursoProviderConfig = client
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_database.DatabaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "create database plan", map[string]interface{}{
		"plan": data,
	})
	fmt.Printf("create database plan: %+v\n", data)

	var dbSeed tursoclient.OptCreateDatabaseInputSeed
	if !data.Seed.IsNull() && !data.Seed.IsUnknown() {
		dbSeed = tursoclient.NewOptCreateDatabaseInputSeed(tursoclient.CreateDatabaseInputSeed{
			Type:      tursoclient.NewOptCreateDatabaseInputSeedType(tursoclient.CreateDatabaseInputSeedType(data.Seed.SeedType.ValueString())),
			Name:      optString(data.Seed.Name),
			URL:       optString(data.Seed.Url),
			Timestamp: optString(data.Seed.Timestamp),
		})
	}

	createReq := tursoclient.CreateDatabaseInput{
		Name:      data.Name.ValueString(),
		Group:     data.Group.ValueString(),
		Seed:      dbSeed,
		SizeLimit: optString(data.SizeLimit),
		IsSchema:  optBool(data.IsSchema),
		Schema:    optString(data.Schema),
	}
	tflog.Trace(ctx, "creating database", map[string]interface{}{
		"request": createReq,
	})
	fmt.Printf("creating database: %+v\n", createReq)
	res, err := r.Client.CreateDatabase(ctx, &createReq, tursoclient.CreateDatabaseParams{
		OrganizationName: r.Organization,
	})
	if err != nil {
		resp.Diagnostics.AddError("error creating database", err.Error())
		return
	}
	db, ok := res.(*tursoclient.CreateDatabaseOK)
	if !ok {
		resp.Diagnostics.AddError("error creating database", "unexpected response from server")
		return
	}

	dbName := string(db.Database.Value.Name.Value)
	resp.Diagnostics.Append(r.readDatabaseResource(ctx, dbName, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created database resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_database.DatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readDatabaseResource(ctx, data.Name.ValueString(), &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("not implemented", "database resource does not support updates")
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_database.DatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Printf("delete database: %+v\n", data)
	_, err := r.Client.DeleteDatabase(ctx, tursoclient.DeleteDatabaseParams{
		OrganizationName: r.Organization,
		DatabaseName:     data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fmt.Printf("importing database: %+v\n", req)
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *tursoProviderConfig) readDatabase(ctx context.Context, name string) (tursoclient.Database, diag.Diagnostics) {
	resp, err := r.Client.GetDatabase(ctx, tursoclient.GetDatabaseParams{
		OrganizationName: r.Organization,
		DatabaseName:     name,
	})
	if err != nil {
		return tursoclient.Database{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", err.Error()),
		}
	}
	dbData, ok := resp.(*tursoclient.GetDatabaseOK)
	if !ok {
		return tursoclient.Database{}, diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", "database not returned from server"),
		}
	}
	db := dbData.Database.Value
	fmt.Printf("read database: %+v\n", db)
	return db, nil
}

func (r *DatabaseResource) readDatabaseResource(ctx context.Context, name string, data *resource_database.DatabaseModel) diag.Diagnostics {
	db, diags := r.readDatabase(ctx, name)
	if diags.HasError() {
		return diags
	}

	data.Id = types.StringValue(db.Name.Value)
	data.Name = types.StringValue(db.Name.Value)
	data.Group = types.StringValue(db.Group.Value)
	if data.SizeLimit.IsUnknown() {
		data.SizeLimit = types.StringNull()
	}
	if data.Schema.IsUnknown() {
		data.Schema = types.StringNull()
	}
	if data.IsSchema.IsUnknown() {
		data.IsSchema = types.BoolNull()
	}
	if data.Seed.IsUnknown() {
		data.Seed = resource_database.NewSeedValueNull()
	}
	dbVal, diags := resource_database.NewDatabaseValue(resource_database.DatabaseValue{}.AttributeTypes(ctx), map[string]attr.Value{
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
	if diags.HasError() {
		return diags
	}
	data.Database = dbVal

	return nil
}
