package provider

import (
	"context"
	"fmt"

	"github.com/celest-dev/terraform-provider-turso/internal/resource_database"
	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
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
	var plan resource_database.DatabaseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "create database plan", map[string]interface{}{
		"plan": plan,
	})
	fmt.Printf("create database plan: %+v\n", plan)

	var dbSeed tursoclient.OptCreateDatabaseInputSeed
	if !plan.Seed.IsNull() && !plan.Seed.IsUnknown() {
		dbSeed = tursoclient.NewOptCreateDatabaseInputSeed(tursoclient.CreateDatabaseInputSeed{
			Type:      tursoclient.NewOptCreateDatabaseInputSeedType(tursoclient.CreateDatabaseInputSeedType(plan.Seed.SeedType.ValueString())),
			Name:      optString(plan.Seed.Name),
			URL:       optString(plan.Seed.Url),
			Timestamp: optString(plan.Seed.Timestamp),
		})
	}

	createReq := tursoclient.CreateDatabaseInput{
		Name:      plan.Name.ValueString(),
		Group:     plan.Group.ValueString(),
		Seed:      dbSeed,
		SizeLimit: optString(plan.SizeLimit),
		IsSchema:  optBool(plan.IsSchema),
		Schema:    optString(plan.Schema),
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
	resp.Diagnostics.Append(r.readDatabase(ctx, dbName, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created database resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_database.DatabaseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readDatabase(ctx, data.Name.ValueString(), &data)...)
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

	// Read Terraform prior state data into the model
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

func (r *DatabaseResource) readDatabase(ctx context.Context, name string, data *resource_database.DatabaseModel) diag.Diagnostics {
	resp, err := r.Client.GetDatabase(ctx, tursoclient.GetDatabaseParams{
		OrganizationName: r.Organization,
		DatabaseName:     name,
	})
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", err.Error()),
		}
	}
	db, ok := resp.(*tursoclient.GetDatabaseOK)
	if !ok {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", "database not returned from server"),
		}
	}
	fmt.Printf("read database: %+v\n", db)

	data.Id = types.StringValue(db.Database.Value.Name.Value)
	data.Name = types.StringValue(db.Database.Value.Name.Value)
	data.Group = types.StringValue(db.Database.Value.Group.Value)
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
	data.Database = resource_database.DatabaseValue{
		DbId:          types.StringValue(db.Database.Value.DbId.Value),
		Name:          types.StringValue(db.Database.Value.Name.Value),
		Group:         types.StringValue(db.Database.Value.Group.Value),
		Hostname:      types.StringValue(db.Database.Value.Hostname.Value),
		Regions:       encodeStringList(db.Database.Value.Regions),
		PrimaryRegion: types.StringValue(db.Database.Value.PrimaryRegion.Value),
		Schema:        types.StringValue(db.Database.Value.Schema.Value),
		IsSchema:      types.BoolValue(db.Database.Value.IsSchema.Value),
		DatabaseType:  types.StringValue(db.Database.Value.Type.Value),
		Archived:      types.BoolValue(db.Database.Value.Archived.Value),
		Version:       types.StringValue(db.Database.Value.Version.Value),
		AllowAttach:   types.BoolValue(db.Database.Value.AllowAttach.Value),
		BlockReads:    types.BoolValue(db.Database.Value.BlockReads.Value),
		BlockWrites:   types.BoolValue(db.Database.Value.BlockWrites.Value),
	}

	return nil
}
