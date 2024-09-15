package resource_database

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func DatabaseResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"database": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"allow_attach": schema.BoolAttribute{
						Computed:            true,
						Description:         "The current status for allowing the database to be attached to another.",
						MarkdownDescription: "The current status for allowing the database to be attached to another.",
					},
					"archived": schema.BoolAttribute{
						Computed:            true,
						Description:         "The current status of the database. If `true`, the database is archived and requires a manual unarchive step.",
						MarkdownDescription: "The current status of the database. If `true`, the database is archived and requires a manual unarchive step.",
					},
					"block_reads": schema.BoolAttribute{
						Computed:            true,
						Description:         "The current status for blocked reads.",
						MarkdownDescription: "The current status for blocked reads.",
					},
					"block_writes": schema.BoolAttribute{
						Computed:            true,
						Description:         "The current status for blocked writes.",
						MarkdownDescription: "The current status for blocked writes.",
					},
					"db_id": schema.StringAttribute{
						Computed:            true,
						Description:         "The database universal unique identifier (UUID).",
						MarkdownDescription: "The database universal unique identifier (UUID).",
					},
					"group": schema.StringAttribute{
						Computed:            true,
						Description:         "The name of the group the database belongs to.",
						MarkdownDescription: "The name of the group the database belongs to.",
					},
					"hostname": schema.StringAttribute{
						Computed:            true,
						Description:         "The DNS hostname used for client libSQL and HTTP connections.",
						MarkdownDescription: "The DNS hostname used for client libSQL and HTTP connections.",
					},
					"is_schema": schema.BoolAttribute{
						Computed:            true,
						Description:         "If this database controls other child databases then this will be `true`. See [Multi-DB Schemas](/features/multi-db-schemas).",
						MarkdownDescription: "If this database controls other child databases then this will be `true`. See [Multi-DB Schemas](/features/multi-db-schemas).",
					},
					"name": schema.StringAttribute{
						Computed:            true,
						Description:         "The database name, **unique** across your organization.",
						MarkdownDescription: "The database name, **unique** across your organization.",
					},
					"primary_region": schema.StringAttribute{
						Computed:            true,
						Description:         "The primary region location code the group the database belongs to.",
						MarkdownDescription: "The primary region location code the group the database belongs to.",
					},
					"regions": schema.ListAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "A list of regions for the group the database belongs to.",
						MarkdownDescription: "A list of regions for the group the database belongs to.",
					},
					"schema": schema.StringAttribute{
						Computed:            true,
						Description:         "The name of the parent database that owns the schema for this database. See [Multi-DB Schemas](/features/multi-db-schemas).",
						MarkdownDescription: "The name of the parent database that owns the schema for this database. See [Multi-DB Schemas](/features/multi-db-schemas).",
					},
					"type": schema.StringAttribute{
						Computed:            true,
						Description:         "The string representing the object type.",
						MarkdownDescription: "The string representing the object type.",
						Default:             stringdefault.StaticString("logical"),
					},
					"version": schema.StringAttribute{
						Computed:            true,
						Description:         "The current libSQL version the database is running.",
						MarkdownDescription: "The current libSQL version the database is running.",
					},
				},
				CustomType: DatabaseType{
					ObjectType: types.ObjectType{
						AttrTypes: DatabaseValue{}.AttributeTypes(ctx),
					},
				},
				Computed: true,
			},
			"group": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the group where the database should be created. **The group must already exist.**",
				MarkdownDescription: "The name of the group where the database should be created. **The group must already exist.**",
			},
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The name of the database.",
				MarkdownDescription: "The name of the database.",
			},
			"is_schema": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Mark this database as the parent schema database that updates child databases with any schema changes. See [Multi-DB Schemas](/features/multi-db-schemas).",
				MarkdownDescription: "Mark this database as the parent schema database that updates child databases with any schema changes. See [Multi-DB Schemas](/features/multi-db-schemas).",
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the new database. Must contain only lowercase letters, numbers, dashes. No longer than 64 characters.",
				MarkdownDescription: "The name of the new database. Must contain only lowercase letters, numbers, dashes. No longer than 64 characters.",
			},
			"schema": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The name of the parent database to use as the schema. See [Multi-DB Schemas](/features/multi-db-schemas).",
				MarkdownDescription: "The name of the parent database to use as the schema. See [Multi-DB Schemas](/features/multi-db-schemas).",
			},
			"seed": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The name of the existing database when `database` is used as a seed type.",
						MarkdownDescription: "The name of the existing database when `database` is used as a seed type.",
					},
					"timestamp": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "A formatted [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) recovery point to create a database from. This must be within the last 24 hours, or 30 days on the scaler plan.",
						MarkdownDescription: "A formatted [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) recovery point to create a database from. This must be within the last 24 hours, or 30 days on the scaler plan.",
					},
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The type of seed to be used to create a new database.",
						MarkdownDescription: "The type of seed to be used to create a new database.",
						Validators: []validator.String{
							stringvalidator.OneOf(
								"database",
								"dump",
							),
						},
					},
					"url": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The URL returned by [upload dump](/api-reference/databases/upload-dump) can be used with the `dump` seed type.",
						MarkdownDescription: "The URL returned by [upload dump](/api-reference/databases/upload-dump) can be used with the `dump` seed type.",
					},
				},
				CustomType: SeedType{
					ObjectType: types.ObjectType{
						AttrTypes: SeedValue{}.AttributeTypes(ctx),
					},
				},
				Optional: true,
				Computed: true,
			},
			"allow_attach": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Allow or disallow attaching databases to the current database.",
				MarkdownDescription: "Allow or disallow attaching databases to the current database.",
			},
			"block_reads": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Block all database reads.",
				MarkdownDescription: "Block all database reads.",
			},
			"block_writes": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Block all database writes.",
				MarkdownDescription: "Block all database writes.",
			},
			"size_limit": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The maximum size of the database in bytes. Values with units are also accepted, e.g. 1mb, 256mb, 1gb.",
				MarkdownDescription: "The maximum size of the database in bytes. Values with units are also accepted, e.g. 1mb, 256mb, 1gb.",
			},
		},
	}
}

type DatabaseModel struct {
	Database    DatabaseValue `tfsdk:"database"`
	Group       types.String  `tfsdk:"group"`
	Id          types.String  `tfsdk:"id"`
	IsSchema    types.Bool    `tfsdk:"is_schema"`
	Name        types.String  `tfsdk:"name"`
	Schema      types.String  `tfsdk:"schema"`
	Seed        SeedValue     `tfsdk:"seed"`
	AllowAttach types.Bool    `tfsdk:"allow_attach"`
	BlockReads  types.Bool    `tfsdk:"block_reads"`
	BlockWrites types.Bool    `tfsdk:"block_writes"`
	SizeLimit   types.String  `tfsdk:"size_limit"`
}

var _ basetypes.ObjectTypable = DatabaseType{}

type DatabaseType struct {
	basetypes.ObjectType
}

func (t DatabaseType) Equal(o attr.Type) bool {
	other, ok := o.(DatabaseType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t DatabaseType) String() string {
	return "DatabaseType"
}

func (t DatabaseType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	allowAttachAttribute, ok := attributes["allow_attach"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`allow_attach is missing from object`)

		return nil, diags
	}

	allowAttachVal, ok := allowAttachAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`allow_attach expected to be basetypes.BoolValue, was: %T`, allowAttachAttribute))
	}

	archivedAttribute, ok := attributes["archived"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`archived is missing from object`)

		return nil, diags
	}

	archivedVal, ok := archivedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`archived expected to be basetypes.BoolValue, was: %T`, archivedAttribute))
	}

	blockReadsAttribute, ok := attributes["block_reads"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`block_reads is missing from object`)

		return nil, diags
	}

	blockReadsVal, ok := blockReadsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`block_reads expected to be basetypes.BoolValue, was: %T`, blockReadsAttribute))
	}

	blockWritesAttribute, ok := attributes["block_writes"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`block_writes is missing from object`)

		return nil, diags
	}

	blockWritesVal, ok := blockWritesAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`block_writes expected to be basetypes.BoolValue, was: %T`, blockWritesAttribute))
	}

	dbIdAttribute, ok := attributes["db_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`db_id is missing from object`)

		return nil, diags
	}

	dbIdVal, ok := dbIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`db_id expected to be basetypes.StringValue, was: %T`, dbIdAttribute))
	}

	groupAttribute, ok := attributes["group"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`group is missing from object`)

		return nil, diags
	}

	groupVal, ok := groupAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`group expected to be basetypes.StringValue, was: %T`, groupAttribute))
	}

	hostnameAttribute, ok := attributes["hostname"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`hostname is missing from object`)

		return nil, diags
	}

	hostnameVal, ok := hostnameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`hostname expected to be basetypes.StringValue, was: %T`, hostnameAttribute))
	}

	isSchemaAttribute, ok := attributes["is_schema"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`is_schema is missing from object`)

		return nil, diags
	}

	isSchemaVal, ok := isSchemaAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`is_schema expected to be basetypes.BoolValue, was: %T`, isSchemaAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	primaryRegionAttribute, ok := attributes["primary_region"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`primary_region is missing from object`)

		return nil, diags
	}

	primaryRegionVal, ok := primaryRegionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`primary_region expected to be basetypes.StringValue, was: %T`, primaryRegionAttribute))
	}

	regionsAttribute, ok := attributes["regions"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`regions is missing from object`)

		return nil, diags
	}

	regionsVal, ok := regionsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`regions expected to be basetypes.ListValue, was: %T`, regionsAttribute))
	}

	schemaAttribute, ok := attributes["schema"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`schema is missing from object`)

		return nil, diags
	}

	schemaVal, ok := schemaAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`schema expected to be basetypes.StringValue, was: %T`, schemaAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	versionAttribute, ok := attributes["version"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`version is missing from object`)

		return nil, diags
	}

	versionVal, ok := versionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`version expected to be basetypes.StringValue, was: %T`, versionAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return DatabaseValue{
		AllowAttach:   allowAttachVal,
		Archived:      archivedVal,
		BlockReads:    blockReadsVal,
		BlockWrites:   blockWritesVal,
		DbId:          dbIdVal,
		Group:         groupVal,
		Hostname:      hostnameVal,
		IsSchema:      isSchemaVal,
		Name:          nameVal,
		PrimaryRegion: primaryRegionVal,
		Regions:       regionsVal,
		Schema:        schemaVal,
		DatabaseType:  typeVal,
		Version:       versionVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewDatabaseValueNull() DatabaseValue {
	return DatabaseValue{
		state: attr.ValueStateNull,
	}
}

func NewDatabaseValueUnknown() DatabaseValue {
	return DatabaseValue{
		state: attr.ValueStateUnknown,
	}
}

func NewDatabaseValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (DatabaseValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing DatabaseValue Attribute Value",
				"While creating a DatabaseValue value, a missing attribute value was detected. "+
					"A DatabaseValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("DatabaseValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid DatabaseValue Attribute Type",
				"While creating a DatabaseValue value, an invalid attribute value was detected. "+
					"A DatabaseValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("DatabaseValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("DatabaseValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra DatabaseValue Attribute Value",
				"While creating a DatabaseValue value, an extra attribute value was detected. "+
					"A DatabaseValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra DatabaseValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewDatabaseValueUnknown(), diags
	}

	allowAttachAttribute, ok := attributes["allow_attach"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`allow_attach is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	allowAttachVal, ok := allowAttachAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`allow_attach expected to be basetypes.BoolValue, was: %T`, allowAttachAttribute))
	}

	archivedAttribute, ok := attributes["archived"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`archived is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	archivedVal, ok := archivedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`archived expected to be basetypes.BoolValue, was: %T`, archivedAttribute))
	}

	blockReadsAttribute, ok := attributes["block_reads"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`block_reads is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	blockReadsVal, ok := blockReadsAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`block_reads expected to be basetypes.BoolValue, was: %T`, blockReadsAttribute))
	}

	blockWritesAttribute, ok := attributes["block_writes"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`block_writes is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	blockWritesVal, ok := blockWritesAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`block_writes expected to be basetypes.BoolValue, was: %T`, blockWritesAttribute))
	}

	dbIdAttribute, ok := attributes["db_id"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`db_id is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	dbIdVal, ok := dbIdAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`db_id expected to be basetypes.StringValue, was: %T`, dbIdAttribute))
	}

	groupAttribute, ok := attributes["group"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`group is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	groupVal, ok := groupAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`group expected to be basetypes.StringValue, was: %T`, groupAttribute))
	}

	hostnameAttribute, ok := attributes["hostname"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`hostname is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	hostnameVal, ok := hostnameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`hostname expected to be basetypes.StringValue, was: %T`, hostnameAttribute))
	}

	isSchemaAttribute, ok := attributes["is_schema"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`is_schema is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	isSchemaVal, ok := isSchemaAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`is_schema expected to be basetypes.BoolValue, was: %T`, isSchemaAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	primaryRegionAttribute, ok := attributes["primary_region"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`primary_region is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	primaryRegionVal, ok := primaryRegionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`primary_region expected to be basetypes.StringValue, was: %T`, primaryRegionAttribute))
	}

	regionsAttribute, ok := attributes["regions"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`regions is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	regionsVal, ok := regionsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`regions expected to be basetypes.ListValue, was: %T`, regionsAttribute))
	}

	schemaAttribute, ok := attributes["schema"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`schema is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	schemaVal, ok := schemaAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`schema expected to be basetypes.StringValue, was: %T`, schemaAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	versionAttribute, ok := attributes["version"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`version is missing from object`)

		return NewDatabaseValueUnknown(), diags
	}

	versionVal, ok := versionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`version expected to be basetypes.StringValue, was: %T`, versionAttribute))
	}

	if diags.HasError() {
		return NewDatabaseValueUnknown(), diags
	}

	return DatabaseValue{
		AllowAttach:   allowAttachVal,
		Archived:      archivedVal,
		BlockReads:    blockReadsVal,
		BlockWrites:   blockWritesVal,
		DbId:          dbIdVal,
		Group:         groupVal,
		Hostname:      hostnameVal,
		IsSchema:      isSchemaVal,
		Name:          nameVal,
		PrimaryRegion: primaryRegionVal,
		Regions:       regionsVal,
		Schema:        schemaVal,
		DatabaseType:  typeVal,
		Version:       versionVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewDatabaseValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) DatabaseValue {
	object, diags := NewDatabaseValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewDatabaseValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t DatabaseType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewDatabaseValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewDatabaseValueUnknown(), nil
	}

	if in.IsNull() {
		return NewDatabaseValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewDatabaseValueMust(DatabaseValue{}.AttributeTypes(ctx), attributes), nil
}

func (t DatabaseType) ValueType(ctx context.Context) attr.Value {
	return DatabaseValue{}
}

var _ basetypes.ObjectValuable = DatabaseValue{}

type DatabaseValue struct {
	AllowAttach   basetypes.BoolValue   `tfsdk:"allow_attach"`
	Archived      basetypes.BoolValue   `tfsdk:"archived"`
	BlockReads    basetypes.BoolValue   `tfsdk:"block_reads"`
	BlockWrites   basetypes.BoolValue   `tfsdk:"block_writes"`
	DbId          basetypes.StringValue `tfsdk:"db_id"`
	Group         basetypes.StringValue `tfsdk:"group"`
	Hostname      basetypes.StringValue `tfsdk:"hostname"`
	IsSchema      basetypes.BoolValue   `tfsdk:"is_schema"`
	Name          basetypes.StringValue `tfsdk:"name"`
	PrimaryRegion basetypes.StringValue `tfsdk:"primary_region"`
	Regions       basetypes.ListValue   `tfsdk:"regions"`
	Schema        basetypes.StringValue `tfsdk:"schema"`
	DatabaseType  basetypes.StringValue `tfsdk:"type"`
	Version       basetypes.StringValue `tfsdk:"version"`
	state         attr.ValueState
}

func (v DatabaseValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 14)

	var val tftypes.Value
	var err error

	attrTypes["allow_attach"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["archived"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["block_reads"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["block_writes"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["db_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["group"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["hostname"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["is_schema"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["primary_region"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["regions"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["schema"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["version"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 14)

		val, err = v.AllowAttach.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["allow_attach"] = val

		val, err = v.Archived.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["archived"] = val

		val, err = v.BlockReads.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["block_reads"] = val

		val, err = v.BlockWrites.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["block_writes"] = val

		val, err = v.DbId.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["db_id"] = val

		val, err = v.Group.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["group"] = val

		val, err = v.Hostname.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["hostname"] = val

		val, err = v.IsSchema.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["is_schema"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.PrimaryRegion.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["primary_region"] = val

		val, err = v.Regions.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["regions"] = val

		val, err = v.Schema.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["schema"] = val

		val, err = v.DatabaseType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["type"] = val

		val, err = v.Version.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["version"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v DatabaseValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v DatabaseValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v DatabaseValue) String() string {
	return "DatabaseValue"
}

func (v DatabaseValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	regionsVal, d := types.ListValue(types.StringType, v.Regions.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"allow_attach":   basetypes.BoolType{},
			"archived":       basetypes.BoolType{},
			"block_reads":    basetypes.BoolType{},
			"block_writes":   basetypes.BoolType{},
			"db_id":          basetypes.StringType{},
			"group":          basetypes.StringType{},
			"hostname":       basetypes.StringType{},
			"is_schema":      basetypes.BoolType{},
			"name":           basetypes.StringType{},
			"primary_region": basetypes.StringType{},
			"regions": basetypes.ListType{
				ElemType: types.StringType,
			},
			"schema":  basetypes.StringType{},
			"type":    basetypes.StringType{},
			"version": basetypes.StringType{},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"allow_attach":   basetypes.BoolType{},
		"archived":       basetypes.BoolType{},
		"block_reads":    basetypes.BoolType{},
		"block_writes":   basetypes.BoolType{},
		"db_id":          basetypes.StringType{},
		"group":          basetypes.StringType{},
		"hostname":       basetypes.StringType{},
		"is_schema":      basetypes.BoolType{},
		"name":           basetypes.StringType{},
		"primary_region": basetypes.StringType{},
		"regions": basetypes.ListType{
			ElemType: types.StringType,
		},
		"schema":  basetypes.StringType{},
		"type":    basetypes.StringType{},
		"version": basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"allow_attach":   v.AllowAttach,
			"archived":       v.Archived,
			"block_reads":    v.BlockReads,
			"block_writes":   v.BlockWrites,
			"db_id":          v.DbId,
			"group":          v.Group,
			"hostname":       v.Hostname,
			"is_schema":      v.IsSchema,
			"name":           v.Name,
			"primary_region": v.PrimaryRegion,
			"regions":        regionsVal,
			"schema":         v.Schema,
			"type":           v.DatabaseType,
			"version":        v.Version,
		})

	return objVal, diags
}

func (v DatabaseValue) Equal(o attr.Value) bool {
	other, ok := o.(DatabaseValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.AllowAttach.Equal(other.AllowAttach) {
		return false
	}

	if !v.Archived.Equal(other.Archived) {
		return false
	}

	if !v.BlockReads.Equal(other.BlockReads) {
		return false
	}

	if !v.BlockWrites.Equal(other.BlockWrites) {
		return false
	}

	if !v.DbId.Equal(other.DbId) {
		return false
	}

	if !v.Group.Equal(other.Group) {
		return false
	}

	if !v.Hostname.Equal(other.Hostname) {
		return false
	}

	if !v.IsSchema.Equal(other.IsSchema) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.PrimaryRegion.Equal(other.PrimaryRegion) {
		return false
	}

	if !v.Regions.Equal(other.Regions) {
		return false
	}

	if !v.Schema.Equal(other.Schema) {
		return false
	}

	if !v.DatabaseType.Equal(other.DatabaseType) {
		return false
	}

	if !v.Version.Equal(other.Version) {
		return false
	}

	return true
}

func (v DatabaseValue) Type(ctx context.Context) attr.Type {
	return DatabaseType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v DatabaseValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"allow_attach":   basetypes.BoolType{},
		"archived":       basetypes.BoolType{},
		"block_reads":    basetypes.BoolType{},
		"block_writes":   basetypes.BoolType{},
		"db_id":          basetypes.StringType{},
		"group":          basetypes.StringType{},
		"hostname":       basetypes.StringType{},
		"is_schema":      basetypes.BoolType{},
		"name":           basetypes.StringType{},
		"primary_region": basetypes.StringType{},
		"regions": basetypes.ListType{
			ElemType: types.StringType,
		},
		"schema":  basetypes.StringType{},
		"type":    basetypes.StringType{},
		"version": basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = SeedType{}

type SeedType struct {
	basetypes.ObjectType
}

func (t SeedType) Equal(o attr.Type) bool {
	other, ok := o.(SeedType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t SeedType) String() string {
	return "SeedType"
}

func (t SeedType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return nil, diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	timestampAttribute, ok := attributes["timestamp"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`timestamp is missing from object`)

		return nil, diags
	}

	timestampVal, ok := timestampAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`timestamp expected to be basetypes.StringValue, was: %T`, timestampAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	urlAttribute, ok := attributes["url"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`url is missing from object`)

		return nil, diags
	}

	urlVal, ok := urlAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`url expected to be basetypes.StringValue, was: %T`, urlAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return SeedValue{
		Name:      nameVal,
		Timestamp: timestampVal,
		SeedType:  typeVal,
		Url:       urlVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewSeedValueNull() SeedValue {
	return SeedValue{
		state: attr.ValueStateNull,
	}
}

func NewSeedValueUnknown() SeedValue {
	return SeedValue{
		state: attr.ValueStateUnknown,
	}
}

func NewSeedValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (SeedValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing SeedValue Attribute Value",
				"While creating a SeedValue value, a missing attribute value was detected. "+
					"A SeedValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SeedValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid SeedValue Attribute Type",
				"While creating a SeedValue value, an invalid attribute value was detected. "+
					"A SeedValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("SeedValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("SeedValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra SeedValue Attribute Value",
				"While creating a SeedValue value, an extra attribute value was detected. "+
					"A SeedValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra SeedValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewSeedValueUnknown(), diags
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewSeedValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	timestampAttribute, ok := attributes["timestamp"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`timestamp is missing from object`)

		return NewSeedValueUnknown(), diags
	}

	timestampVal, ok := timestampAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`timestamp expected to be basetypes.StringValue, was: %T`, timestampAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewSeedValueUnknown(), diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	urlAttribute, ok := attributes["url"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`url is missing from object`)

		return NewSeedValueUnknown(), diags
	}

	urlVal, ok := urlAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`url expected to be basetypes.StringValue, was: %T`, urlAttribute))
	}

	if diags.HasError() {
		return NewSeedValueUnknown(), diags
	}

	return SeedValue{
		Name:      nameVal,
		Timestamp: timestampVal,
		SeedType:  typeVal,
		Url:       urlVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewSeedValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) SeedValue {
	object, diags := NewSeedValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewSeedValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t SeedType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewSeedValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewSeedValueUnknown(), nil
	}

	if in.IsNull() {
		return NewSeedValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewSeedValueMust(SeedValue{}.AttributeTypes(ctx), attributes), nil
}

func (t SeedType) ValueType(ctx context.Context) attr.Value {
	return SeedValue{}
}

var _ basetypes.ObjectValuable = SeedValue{}

type SeedValue struct {
	Name      basetypes.StringValue `tfsdk:"name"`
	Timestamp basetypes.StringValue `tfsdk:"timestamp"`
	SeedType  basetypes.StringValue `tfsdk:"type"`
	Url       basetypes.StringValue `tfsdk:"url"`
	state     attr.ValueState
}

func (v SeedValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 4)

	var val tftypes.Value
	var err error

	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["timestamp"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["url"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 4)

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.Timestamp.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["timestamp"] = val

		val, err = v.SeedType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["type"] = val

		val, err = v.Url.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["url"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v SeedValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v SeedValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v SeedValue) String() string {
	return "SeedValue"
}

func (v SeedValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"name":      basetypes.StringType{},
		"timestamp": basetypes.StringType{},
		"type":      basetypes.StringType{},
		"url":       basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"name":      v.Name,
			"timestamp": v.Timestamp,
			"type":      v.SeedType,
			"url":       v.Url,
		})

	return objVal, diags
}

func (v SeedValue) Equal(o attr.Value) bool {
	other, ok := o.(SeedValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.Timestamp.Equal(other.Timestamp) {
		return false
	}

	if !v.SeedType.Equal(other.SeedType) {
		return false
	}

	if !v.Url.Equal(other.Url) {
		return false
	}

	return true
}

func (v SeedValue) Type(ctx context.Context) attr.Type {
	return SeedType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v SeedValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"name":      basetypes.StringType{},
		"timestamp": basetypes.StringType{},
		"type":      basetypes.StringType{},
		"url":       basetypes.StringType{},
	}
}
