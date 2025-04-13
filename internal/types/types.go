package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type StringValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuable = StringValue{}

func (t *StringValue) UnmarshalJSON(data []byte) error {
	var v *string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v == nil {
		t.StringValue = basetypes.NewStringNull()
	} else {
		t.StringValue = basetypes.NewStringValue(*v)
	}

	return nil
}

func (t StringValue) MarshalJSON() ([]byte, error) {
	if t.IsNull() || t.IsUnknown() {
		return []byte("null"), nil
	}
	return json.Marshal(t.StringValue.ValueString())
}

func (v StringValue) Equal(o attr.Value) bool {
	// fmt.Println("------ STRINGVALUE EQUAL")
	other, ok := o.(StringValue)

	if !ok {
		// fmt.Println("FALSE")
		return false
	}

	// fmt.Println(v.StringValue)
	// fmt.Println(other.StringValue)
	// fmt.Println(v.StringValue.Equal(other.StringValue))

	return v.StringValue.Equal(other.StringValue)
}

func (v StringValue) StringSemanticEquals(ctx context.Context, sv basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := sv.(StringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", sv),
		)
		return false, diags
	}
	return newValue.Equal(v), diags
}

func (v StringValue) Type(ctx context.Context) attr.Type {
	return StringType{}
}

type StringType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = StringType{}

func (t StringType) Equal(o attr.Type) bool {
	// fmt.Println("STRINGTYPE EQUAL")
	other, ok := o.(StringType)

	if !ok {
		// fmt.Println("FALSE?")
		return false
	}

	// fmt.Println(t.StringType)
	// fmt.Println(other.StringType)
	// fmt.Println(t.StringType.Equal(other.StringType))
	// fmt.Println("----------------")

	return t.StringType.Equal(other.StringType)
}

func (t StringType) String() string {
	return "StringType"
}

func (t StringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := StringValue{
		StringValue: in,
	}
	return value, nil
}

func (t StringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t StringType) ValueType(ctx context.Context) attr.Value {
	return StringValue{}
}

type BoolValue struct {
	basetypes.BoolValue
}

var _ basetypes.BoolValuable = BoolValue{}

func (b *BoolValue) UnmarshalJSON(data []byte) error {
	var v *bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v == nil {
		b.BoolValue = basetypes.NewBoolNull()
	} else {
		b.BoolValue = basetypes.NewBoolValue(*v)
	}
	return nil
}

func (b BoolValue) MarshalJSON() ([]byte, error) {
	if b.IsNull() || b.IsUnknown() {
		return []byte("null"), nil
	}
	return json.Marshal(b.BoolValue.ValueBool())
}

func (v BoolValue) Equal(o attr.Value) bool {
	other, ok := o.(BoolValue)

	if !ok {
		return false
	}

	return v.BoolValue.Equal(other.BoolValue)
}

func (v BoolValue) Type(ctx context.Context) attr.Type {
	return BoolType{}
}

type BoolType struct {
	basetypes.BoolType
}

var _ basetypes.BoolTypable = BoolType{}

func (t BoolType) Equal(o attr.Type) bool {
	other, ok := o.(BoolType)

	if !ok {
		return false
	}

	return t.BoolType.Equal(other.BoolType)
}

func (t BoolType) String() string {
	return "BoolType"
}

func (t BoolType) ValueFromBool(ctx context.Context, in basetypes.BoolValue) (basetypes.BoolValuable, diag.Diagnostics) {
	value := BoolValue{
		BoolValue: in,
	}
	return value, nil
}

func (t BoolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	boolValue, ok := attrValue.(basetypes.BoolValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	boolValuable, diags := t.ValueFromBool(ctx, boolValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting BoolValue to BoolValuable: %v", diags)
	}

	return boolValuable, nil
}

func (t BoolType) ValueType(ctx context.Context) attr.Value {
	return BoolValue{}
}
