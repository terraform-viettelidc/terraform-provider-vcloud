//go:build ALL || nsxt || functional

package viettelidc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceNsxtManager(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)

	var params = StringMap{
		"FuncName":    t.Name(),
		"NsxtManager": testConfig.Nsxt.Manager,
		"Tags":        "nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdNsxtManager, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// ID must match URN 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc'
					resource.TestMatchResourceAttr("data.vcd_nsxt_manager.nsxt", "id",
						getUuidRegex("urn:vcloud:nsxtmanager:", "$")),
					resource.TestCheckResourceAttr("data.vcd_nsxt_manager.nsxt", "name", params["NsxtManager"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdNsxtManager = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}
`
