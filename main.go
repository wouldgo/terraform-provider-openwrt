package main

import (
	"context"
	"log"

	"github.com/foxboron/terraform-provider-openwrt/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.opentofu.org/foxboron/openwrt",
		Debug:   true,
	}

	err := providerserver.Serve(context.Background(), provider.New("test"), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
