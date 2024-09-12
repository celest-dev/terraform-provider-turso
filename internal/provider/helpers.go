package provider

import (
	"cmp"
	"log"
	"slices"

	"github.com/celest-dev/terraform-provider-turso/internal/tursoclient"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func optString(s basetypes.StringValue) tursoclient.OptString {
	if s.IsNull() || s.IsUnknown() {
		return tursoclient.OptString{}
	}
	return tursoclient.NewOptString(s.ValueString())
}

func optBool(b basetypes.BoolValue) tursoclient.OptBool {
	if b.IsNull() || b.IsUnknown() {
		return tursoclient.OptBool{}
	}
	return tursoclient.NewOptBool(b.ValueBool())
}

func decodeStringSet(v basetypes.SetValue) []string {
	elements := v.Elements()
	result := make([]string, len(elements))
	for i, e := range elements {
		v, ok := e.(basetypes.StringValue)
		if !ok {
			log.Panicf("unexpected type in string set: %T", e)
		}
		result[i] = v.ValueString()
	}
	return result
}

func encodeStringSet(v []string) basetypes.SetValue {
	elements := make([]attr.Value, len(v))
	for i, s := range v {
		elements[i] = basetypes.NewStringValue(s)
	}
	return basetypes.NewSetValueMust(basetypes.StringType{}, elements)
}

func encodeStringList(v []string) basetypes.ListValue {
	elements := make([]attr.Value, len(v))
	for i, s := range v {
		elements[i] = basetypes.NewStringValue(s)
	}
	return basetypes.NewListValueMust(basetypes.StringType{}, elements)
}

func mergeLists[S ~[]E, E cmp.Ordered](a S, b S) S {
	set := make(map[E]struct{})
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		set[v] = struct{}{}
	}
	result := make(S, 0, len(set))
	for v := range set {
		result = append(result, v)
	}
	slices.Sort(result)
	return result
}
