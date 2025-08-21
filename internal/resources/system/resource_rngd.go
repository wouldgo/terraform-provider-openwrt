package system

import "github.com/foxboron/terraform-provider-openwrt/internal/types"

// ProjectModel resource data model that matches the schema.
type RngdModel struct {
	Id        types.StringValue `tfsdk:"id" json:".name,omitempty"`
	Anonymous types.BoolValue   `tfsdk:"anonymous" json:".anonymous,omitzero,omitempty"`
	Type      types.StringValue `tfsdk:"type" json:".type,omitzero,omitempty"`

	Enabled       types.BoolValue   `tfsdk:"enabled" json:"enabled,omitzero"`
	Device        types.StringValue `tfsdk:"device" json:"device,omitzero"`
	PreCMD        types.StringValue `tfsdk:"precmd" json:"precmd,omitzero"`
	FillWatermark types.StringValue `tfsdk:"fill_watermark" json:"fill_watermark,omitzero"`
}
