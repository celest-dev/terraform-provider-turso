// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package datasource_groups

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func GroupsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"archived": schema.BoolAttribute{
							Computed:            true,
							Description:         "Groups on the free tier get archived after some inactivity.",
							MarkdownDescription: "Groups on the free tier get archived after some inactivity.",
						},
						"locations": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "An array of location keys the group is located.",
							MarkdownDescription: "An array of location keys the group is located.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "The group name, unique across your organization.",
							MarkdownDescription: "The group name, unique across your organization.",
						},
						"primary": schema.StringAttribute{
							Computed:            true,
							Description:         "The primary location key.",
							MarkdownDescription: "The primary location key.",
						},
						"uuid": schema.StringAttribute{
							Computed:            true,
							Description:         "The group universal unique identifier (UUID).",
							MarkdownDescription: "The group universal unique identifier (UUID).",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							Description:         "The current libSQL server version the databases in that group are running.",
							MarkdownDescription: "The current libSQL server version the databases in that group are running.",
						},
					},
					CustomType: GroupsType{
						ObjectType: types.ObjectType{
							AttrTypes: GroupsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Computed: true,
			},
		},
	}
}

type GroupsModel struct {
	Groups types.List `tfsdk:"groups"`
}

var _ basetypes.ObjectTypable = GroupsType{}

type GroupsType struct {
	basetypes.ObjectType
}

func (t GroupsType) Equal(o attr.Type) bool {
	other, ok := o.(GroupsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t GroupsType) String() string {
	return "GroupsType"
}

func (t GroupsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

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

	locationsAttribute, ok := attributes["locations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`locations is missing from object`)

		return nil, diags
	}

	locationsVal, ok := locationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`locations expected to be basetypes.ListValue, was: %T`, locationsAttribute))
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

	primaryAttribute, ok := attributes["primary"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`primary is missing from object`)

		return nil, diags
	}

	primaryVal, ok := primaryAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`primary expected to be basetypes.StringValue, was: %T`, primaryAttribute))
	}

	uuidAttribute, ok := attributes["uuid"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`uuid is missing from object`)

		return nil, diags
	}

	uuidVal, ok := uuidAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`uuid expected to be basetypes.StringValue, was: %T`, uuidAttribute))
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

	return GroupsValue{
		Archived:  archivedVal,
		Locations: locationsVal,
		Name:      nameVal,
		Primary:   primaryVal,
		Uuid:      uuidVal,
		Version:   versionVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewGroupsValueNull() GroupsValue {
	return GroupsValue{
		state: attr.ValueStateNull,
	}
}

func NewGroupsValueUnknown() GroupsValue {
	return GroupsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewGroupsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (GroupsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing GroupsValue Attribute Value",
				"While creating a GroupsValue value, a missing attribute value was detected. "+
					"A GroupsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("GroupsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid GroupsValue Attribute Type",
				"While creating a GroupsValue value, an invalid attribute value was detected. "+
					"A GroupsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("GroupsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("GroupsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra GroupsValue Attribute Value",
				"While creating a GroupsValue value, an extra attribute value was detected. "+
					"A GroupsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra GroupsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewGroupsValueUnknown(), diags
	}

	archivedAttribute, ok := attributes["archived"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`archived is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	archivedVal, ok := archivedAttribute.(basetypes.BoolValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`archived expected to be basetypes.BoolValue, was: %T`, archivedAttribute))
	}

	locationsAttribute, ok := attributes["locations"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`locations is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	locationsVal, ok := locationsAttribute.(basetypes.ListValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`locations expected to be basetypes.ListValue, was: %T`, locationsAttribute))
	}

	nameAttribute, ok := attributes["name"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`name is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	nameVal, ok := nameAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`name expected to be basetypes.StringValue, was: %T`, nameAttribute))
	}

	primaryAttribute, ok := attributes["primary"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`primary is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	primaryVal, ok := primaryAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`primary expected to be basetypes.StringValue, was: %T`, primaryAttribute))
	}

	uuidAttribute, ok := attributes["uuid"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`uuid is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	uuidVal, ok := uuidAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`uuid expected to be basetypes.StringValue, was: %T`, uuidAttribute))
	}

	versionAttribute, ok := attributes["version"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`version is missing from object`)

		return NewGroupsValueUnknown(), diags
	}

	versionVal, ok := versionAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`version expected to be basetypes.StringValue, was: %T`, versionAttribute))
	}

	if diags.HasError() {
		return NewGroupsValueUnknown(), diags
	}

	return GroupsValue{
		Archived:  archivedVal,
		Locations: locationsVal,
		Name:      nameVal,
		Primary:   primaryVal,
		Uuid:      uuidVal,
		Version:   versionVal,
		state:     attr.ValueStateKnown,
	}, diags
}

func NewGroupsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) GroupsValue {
	object, diags := NewGroupsValue(attributeTypes, attributes)

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

		panic("NewGroupsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t GroupsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewGroupsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewGroupsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewGroupsValueNull(), nil
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

	return NewGroupsValueMust(GroupsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t GroupsType) ValueType(ctx context.Context) attr.Value {
	return GroupsValue{}
}

var _ basetypes.ObjectValuable = GroupsValue{}

type GroupsValue struct {
	Archived  basetypes.BoolValue   `tfsdk:"archived"`
	Locations basetypes.ListValue   `tfsdk:"locations"`
	Name      basetypes.StringValue `tfsdk:"name"`
	Primary   basetypes.StringValue `tfsdk:"primary"`
	Uuid      basetypes.StringValue `tfsdk:"uuid"`
	Version   basetypes.StringValue `tfsdk:"version"`
	state     attr.ValueState
}

func (v GroupsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 6)

	var val tftypes.Value
	var err error

	attrTypes["archived"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["locations"] = basetypes.ListType{
		ElemType: types.StringType,
	}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["primary"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["uuid"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["version"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 6)

		val, err = v.Archived.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["archived"] = val

		val, err = v.Locations.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["locations"] = val

		val, err = v.Name.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["name"] = val

		val, err = v.Primary.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["primary"] = val

		val, err = v.Uuid.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["uuid"] = val

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

func (v GroupsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v GroupsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v GroupsValue) String() string {
	return "GroupsValue"
}

func (v GroupsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	locationsVal, d := types.ListValue(types.StringType, v.Locations.Elements())

	diags.Append(d...)

	if d.HasError() {
		return types.ObjectUnknown(map[string]attr.Type{
			"archived": basetypes.BoolType{},
			"locations": basetypes.ListType{
				ElemType: types.StringType,
			},
			"name":    basetypes.StringType{},
			"primary": basetypes.StringType{},
			"uuid":    basetypes.StringType{},
			"version": basetypes.StringType{},
		}), diags
	}

	attributeTypes := map[string]attr.Type{
		"archived": basetypes.BoolType{},
		"locations": basetypes.ListType{
			ElemType: types.StringType,
		},
		"name":    basetypes.StringType{},
		"primary": basetypes.StringType{},
		"uuid":    basetypes.StringType{},
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
			"archived":  v.Archived,
			"locations": locationsVal,
			"name":      v.Name,
			"primary":   v.Primary,
			"uuid":      v.Uuid,
			"version":   v.Version,
		})

	return objVal, diags
}

func (v GroupsValue) Equal(o attr.Value) bool {
	other, ok := o.(GroupsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Archived.Equal(other.Archived) {
		return false
	}

	if !v.Locations.Equal(other.Locations) {
		return false
	}

	if !v.Name.Equal(other.Name) {
		return false
	}

	if !v.Primary.Equal(other.Primary) {
		return false
	}

	if !v.Uuid.Equal(other.Uuid) {
		return false
	}

	if !v.Version.Equal(other.Version) {
		return false
	}

	return true
}

func (v GroupsValue) Type(ctx context.Context) attr.Type {
	return GroupsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v GroupsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"archived": basetypes.BoolType{},
		"locations": basetypes.ListType{
			ElemType: types.StringType,
		},
		"name":    basetypes.StringType{},
		"primary": basetypes.StringType{},
		"uuid":    basetypes.StringType{},
		"version": basetypes.StringType{},
	}
}
