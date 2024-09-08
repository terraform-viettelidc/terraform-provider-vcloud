package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/vmware/terraform-provider-vcd/v3/vcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcloud.Provider})
}
