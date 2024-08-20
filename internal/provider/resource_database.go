package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoadmin"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	client *tursoadmin.Client
}

type DatabaseResourceModel struct {
	Name        types.String `tfsdk:"name"`
	Group       types.String `tfsdk:"group"`
	Seed        types.Object `tfsdk:"seed"`
	SizeLimit   types.String `tfsdk:"size_limit"`
	IsSchema    types.Bool   `tfsdk:"is_schema"`
	Schema      types.String `tfsdk:"schema"`
	AllowAttach types.Bool   `tfsdk:"allow_attach"`
	BlockReads  types.Bool   `tfsdk:"block_reads"`
	BlockWrites types.Bool   `tfsdk:"block_writes"`

	// Computed
	DbId          types.String `tfsdk:"db_id"`
	Hostname      types.String `tfsdk:"hostname"`
	Type          types.String `tfsdk:"type"`
	Version       types.String `tfsdk:"version"`
	PrimaryRegion types.String `tfsdk:"primary_region"`
	Instances     types.Map    `tfsdk:"instances"`
}

type DatabaseSeedModel struct {
	Type      types.String      `tfsdk:"type"`
	Name      types.String      `tfsdk:"name"`
	URL       types.String      `tfsdk:"url"`
	Timestamp timetypes.RFC3339 `tfsdk:"timestamp"`
}

func (DatabaseSeedModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":      types.StringType,
		"name":      types.StringType,
		"url":       types.StringType,
		"timestamp": timetypes.RFC3339Type{},
	}
}

type DatabaseInstanceModel struct {
	UUID     types.String `tfsdk:"uuid"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Region   types.String `tfsdk:"region"`
	Hostname types.String `tfsdk:"hostname"`
}

func (DatabaseInstanceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"uuid":     types.StringType,
		"name":     types.StringType,
		"type":     types.StringType,
		"region":   types.StringType,
		"hostname": types.StringType,
	}
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Database resource",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the new database. Must contain only lowercase letters, numbers, dashes. No longer than 32 characters.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group": schema.StringAttribute{
				MarkdownDescription: "The name of the group where the database should be created. The group must already exist.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"seed": schema.SingleNestedAttribute{
				MarkdownDescription: "Seed configuration for the new database.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of seed to be used to create a new database.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(string(tursoadmin.DatabaseSeedDatabase), string(tursoadmin.DatabaseSeedDump)),
						},
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the existing database when `database` is used as a seed type.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("url")),
							stringvalidator.AtLeastOneOf(path.MatchRelative(), path.MatchRelative().AtParent().AtName("url")),
						},
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "The URL returned by [upload dump](https://docs.turso.tech/api-reference/databases/upload-dump) can be used with the `dump` seed type.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
							stringvalidator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("name"), path.MatchRelative()),
						},
					},
					"timestamp": schema.StringAttribute{
						MarkdownDescription: "A formatted ISO 8601 recovery point to create a database from. This must be within the last 24 hours, or 30 days on the scaler plan.",
						Optional:            true,
						CustomType:          timetypes.RFC3339Type{},
					},
				},
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"size_limit": schema.StringAttribute{
				MarkdownDescription: "The maximum size of the database in bytes. Values with units are also accepted, e.g. 1mb, 256mb, 1gb.",
				Optional:            true,
			},
			"is_schema": schema.BoolAttribute{
				MarkdownDescription: "Mark this database as the parent schema database that updates child databases with any schema changes.",
				Optional:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The name of the parent database to use as the schema.",
				Optional:            true,
			},
			"allow_attach": schema.BoolAttribute{
				MarkdownDescription: "Allow attaching databases to this database.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"block_reads": schema.BoolAttribute{
				MarkdownDescription: "Block all read queries to the database.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"block_writes": schema.BoolAttribute{
				MarkdownDescription: "Block all write queries to the database.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"db_id": schema.StringAttribute{
				MarkdownDescription: "The database universal unique identifier (UUID).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The DNS hostname used for client libSQL and HTTP connections.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The string representing the object type. Default: `logical`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The current libSQL version the database is running.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_region": schema.StringAttribute{
				MarkdownDescription: "The location code for the primary region this database is located.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instances": schema.MapNestedAttribute{
				MarkdownDescription: "The instance configurations for the database.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							MarkdownDescription: "The instance universal unique identifier (UUID).",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the instance (location code).",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of database instance. One of: `primary` or `replica`.",
							Computed:            true,
						},
						"region": schema.StringAttribute{
							MarkdownDescription: "The location code for the region this instance is located.",
							Computed:            true,
						},
						"hostname": schema.StringAttribute{
							MarkdownDescription: "The DNS hostname used for client libSQL and HTTP connections (specific to this instance only).",
							Computed:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*tursoadmin.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *tursoadmin.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "create database plan", map[string]interface{}{
		"plan": plan,
	})
	fmt.Printf("create database plan: %+v\n", plan)

	var dbSeed *tursoadmin.DatabaseSeed
	if !plan.Seed.IsNull() && !plan.Seed.IsUnknown() {
		var seedModel DatabaseSeedModel
		resp.Diagnostics.Append(plan.Seed.As(ctx, &seedModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})...)
		if resp.Diagnostics.HasError() {
			return
		}

		dbSeed = &tursoadmin.DatabaseSeed{
			Type: tursoadmin.DatabaseSeedType(seedModel.Type.ValueString()),
			Name: seedModel.Name.ValueString(),
			URL:  seedModel.URL.ValueString(),
		}

		// timestamp := seedModel.Timestamp
		// if !timestamp.IsNull() && !timestamp.IsUnknown() {
		// 	ts, diags := timestamp.ValueRFC3339Time()
		// 	resp.Diagnostics.Append(diags...)
		// 	if resp.Diagnostics.HasError() {
		// 		return
		// 	}
		// 	dbSeed.Timestamp = ts
		// }
	}

	createReq := tursoadmin.CreateDatabaseRequest{
		Name:           plan.Name.ValueString(),
		Group:          plan.Group.ValueString(),
		Seed:           dbSeed,
		SizeLimit:      plan.SizeLimit.ValueString(),
		IsSchema:       plan.IsSchema.ValueBool(),
		SchemaDatabase: plan.Schema.ValueString(),
	}
	tflog.Trace(ctx, "creating database", map[string]interface{}{
		"request": createReq,
	})
	fmt.Printf("creating database: %+v\n", createReq)
	db, err := r.client.CreateDatabase(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("error creating database", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readDatabase(ctx, db.Name, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created database resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel
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
	var data DatabaseResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Printf("update database: %+v\n", data)
	config, err := r.client.UpdateDatabaseConfiguration(ctx, data.Name.ValueString(), tursoadmin.DatabaseConfig{
		SizeLimit:   data.SizeLimit.ValueStringPointer(),
		AllowAttach: data.AllowAttach.ValueBoolPointer(),
		BlockReads:  data.BlockReads.ValueBoolPointer(),
		BlockWrites: data.BlockWrites.ValueBoolPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("size_limit"), config.SizeLimit)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("allow_attach"), config.AllowAttach)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("block_reads"), config.BlockReads)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("block_writes"), config.BlockWrites)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Printf("delete database: %+v\n", data)
	err := r.client.DeleteDatabase(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("client error", err.Error())
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Trace(ctx, fmt.Sprintf("importing database: %s", req.ID))

	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 {
		resp.Diagnostics.AddError("invalid import ID", fmt.Sprintf("expected format: group_name/database_name, got: %s", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
}

func (r *DatabaseResource) readDatabase(ctx context.Context, name string, data *DatabaseResourceModel) diag.Diagnostics {
	db, err := r.client.GetDatabase(ctx, name)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", err.Error()),
		}
	}
	fmt.Printf("read database: %+v\n", db)

	data.DbId = types.StringValue(db.ID)
	data.Hostname = types.StringValue(db.Hostname)
	if data.Seed.IsUnknown() {
		data.Seed = types.ObjectNull(DatabaseSeedModel{}.AttributeTypes())
	}
	if data.SizeLimit.IsUnknown() {
		data.SizeLimit = types.StringNull()
	}
	if !data.IsSchema.IsNull() && !data.IsSchema.IsUnknown() {
		data.IsSchema = types.BoolValue(db.IsSchema)
	} else {
		data.IsSchema = types.BoolNull()
	}
	if data.Schema.IsUnknown() {
		data.Schema = types.StringNull()
	}
	data.PrimaryRegion = types.StringValue(db.PrimaryRegion)
	data.Type = types.StringValue(db.Type)
	data.Version = types.StringValue(db.Version)
	data.AllowAttach = types.BoolValue(db.AllowAttach)
	data.BlockReads = types.BoolValue(db.BlockReads)
	data.BlockWrites = types.BoolValue(db.BlockWrites)

	instances, err := r.client.ListDatabaseInstances(ctx, db.Name)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("client error", err.Error()),
		}
	}
	instanceElements := make(map[string]attr.Value)
	for _, instance := range instances {
		model := DatabaseInstanceModel{
			UUID:     types.StringValue(instance.UUID),
			Name:     types.StringValue(instance.Name),
			Type:     types.StringValue(string(instance.Type)),
			Region:   types.StringValue(instance.Region),
			Hostname: types.StringValue(instance.Hostname),
		}
		element, diags := types.ObjectValueFrom(ctx, model.AttributeTypes(), model)
		if diags.HasError() {
			return diags
		}
		instanceElements[instance.Region] = element
	}
	instancesValue, diags := types.MapValue(
		types.ObjectType{AttrTypes: DatabaseInstanceModel{}.AttributeTypes()},
		instanceElements,
	)
	if diags.HasError() {
		return diags
	}
	data.Instances = instancesValue

	return nil
}
