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

type BoolValue struct {
	basetypes.BoolValue
}

func NewBoolValue(b bool) BoolValue {
	return BoolValue{
		BoolValue: basetypes.NewBoolValue(b),
	}
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
