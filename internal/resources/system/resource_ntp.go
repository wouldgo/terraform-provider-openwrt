// Copyright (c) https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors
// SPDX-License-Identifier: MPL-2.0

package system

import "github.com/foxboron/terraform-provider-openwrt/internal/types"

// NTPModel resource data model that matches the schema.
type NTPModel struct {
	Id        types.StringValue `tfsdk:"id" json:".name,omitempty"`
	Anonymous types.BoolValue   `tfsdk:"anonymous" json:".anonymous,omitzero,omitempty"`
	Type      types.StringValue `tfsdk:"type" json:".type,omitzero,omitempty"`

	Server       types.StringValue `tfsdk:"server" json:"server,omitzero"`
	EnableServer types.BoolValue   `tfsdk:"enable_server" json:"enable_server,omitzero"`
	Interface    types.StringValue `tfsdk:"interface" json:"interface,omitzero"`
	UseDHCP      types.BoolValue   `tfsdk:"use_dhcp" json:"use_dhcp,omitzero"`
}
