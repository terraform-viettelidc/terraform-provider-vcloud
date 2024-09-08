//go:build ALL || providerVdc || functional

package vcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceProviderVdc(t *testing.T) {
	// Pre-checks
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// Test configuration
	var params = StringMap{
		"ProviderVdcName": testConfig.VCD.NsxtProviderVdc.Name,
	}
	testParamsNotEmpty(t, params)
	configText := templateFill(testAccVcdDatasourceProviderVdc, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Test cases
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "name", params["ProviderVdcName"].(string)),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "is_enabled", "true"),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "status", "1"),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "compute_provider_scope", "vc1"),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "compute_capacity.0.cpu.0.units", "MHz"),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "compute_capacity.0.is_elastic", "false"),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "compute_capacity.0.is_ha", "false"),
					resource.TestCheckResourceAttr("data.vcd_provider_vdc.pvdc1", "compute_capacity.0.memory.0.units", "MB"),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "external_network_ids.0", getProviderVdcDatasourceAttributeUrnRegex("network")),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "capabilities.0", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "host_ids.0", getProviderVdcDatasourceAttributeUrnRegex("host")),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "storage_profile_ids.0", getProviderVdcDatasourceAttributeUrnRegex("providervdcstorageprofile")),
					resource.TestMatchResourceAttr("data.vcd_provider_vdc.pvdc1", "vcenter_id", getProviderVdcDatasourceAttributeUrnRegex("vimserver")),
				),
			},
		},
	})
	postTestChecks(t)
}

// As the `vcd_provider_vdc` data source has a lot of URNs in its attributes, this function tries to centralize URN checking
// for this test case.
func getProviderVdcDatasourceAttributeUrnRegex(itemType string) *regexp.Regexp {
	return getUuidRegex(fmt.Sprintf("urn:vcloud:%s:", itemType), "$")
}

const testAccVcdDatasourceProviderVdc = `
data "vcd_provider_vdc" "pvdc1" {
    name = "{{.ProviderVdcName}}"
}
`
