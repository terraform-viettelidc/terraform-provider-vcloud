package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/terraform-viettelidc/terraform-provider-vcloud/v3/vcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcloud.Provider})
}
