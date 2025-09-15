// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package system

import "github.com/foxboron/terraform-provider-openwrt/internal/types"

// WatchcatModel resource data model that matches the schema.
type WatchcatModel struct {
	Id        types.StringValue `tfsdk:"id" json:".name,omitempty"`
	Anonymous types.BoolValue   `tfsdk:"anonymous" json:".anonymous,omitzero,omitempty"`
	Type      types.StringValue `tfsdk:"type" json:".type,omitzero,omitempty"`

	Mode        types.StringValue `tfsdk:"mode" json:"mode,omitzero"`
	Period      types.StringValue `tfsdk:"period" json:"period,omitzero"`
	PingHosts   types.StringValue `tfsdk:"pinghosts" json:"pinghosts,omitzero"`
	PingSize    types.StringValue `tfsdk:"pingsize" json:"pingsize,omitzero"`
	Interface   types.StringValue `tfsdk:"interface" json:"interface,omitzero"`
	ForceDelay  types.StringValue `tfsdk:"forceDelay" json:"forcedelay,omitzero"`
	Mmifacename types.StringValue `tfsdk:"mmifacename" json:"mmifacename,omitzero"`
	Unlockbands types.StringValue `tfsdk:"unlockbands" json:"unlockbands,omitzero"`
}
