package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-viettelidc/terraform-provider-viettelidc/v3/viettelidc"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: viettelidc.Provider})
}
