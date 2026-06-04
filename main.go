// Copyright https://github.com/Foxboron/terraform-provider-openwrt/graphs/contributors 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"log"

	"github.com/foxboron/terraform-provider-openwrt/internal/api/luci_http"
	"github.com/foxboron/terraform-provider-openwrt/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/foxboron/openwrt",
		Debug:   debug,
	}

	clientFactory, err := luci_http.NewHTTPRPCFactory(luci_http.HTTPRPCConfiguration{})
	if err != nil {
		log.Fatal(err.Error())
	}

	err = providerserver.Serve(context.Background(), provider.New(version, clientFactory), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
