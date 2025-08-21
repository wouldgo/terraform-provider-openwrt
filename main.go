package main

import (
	"context"
	"flag"
	"log"

	"github.com/foxboron/terraform-provider-openwrt/internal/api"
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
		Address: "registry.opentofu.org/foxboron/openwrt",
		Debug:   debug,
	}

	clientFactory, err := api.NewClientFactory()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = providerserver.Serve(context.Background(), provider.New(version, clientFactory), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
