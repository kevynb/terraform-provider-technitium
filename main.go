package main

// docs generation + dependencies moved to tools

import (
	"context"
	"flag"
	"log"

	"github.com/kevynb/terraform-provider-technitium/internal/client"
	"github.com/kevynb/terraform-provider-technitium/internal/model"
	"github.com/kevynb/terraform-provider-technitium/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// will be set by the goreleaser
	// see https://goreleaser.com/cookbooks/using-main.version/
	// also set to "test" and "unittest" by acceptance and unit tests
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/dscain/technitium",
		Debug:   debug,
	}

	apiClientFactory := func(apiURL, apiToken string, skipCertificateVerification bool) (model.DNSApiClient, error) {
		return client.NewClient(apiURL, apiToken, skipCertificateVerification)
	}

	err := providerserver.Serve(context.Background(), provider.New(version, apiClientFactory), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
