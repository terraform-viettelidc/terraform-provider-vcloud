//go:build ALL || functional

package viettelidc

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVcenter(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Vcenter": testConfig.Networking.Vcenter,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(datasourceTestVcenter, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_vcenter.vc", "id", regexp.MustCompile("^urn:vcloud:vimserver:.*")),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "vcenter_version"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "vcenter_host"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "status"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "is_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "connection_status"),
				),
			},
		},
	})
	postTestChecks(t)
}

const datasourceTestVcenter = `
data "vcd_vcenter" "vc" {
	name = "{{.Vcenter}}"
  }
`
