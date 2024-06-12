//go:build network || nsxt || ALL || functional

package viettelidc

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNetworkImportedDS(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"NsxtVdc":           testConfig.Nsxt.Vdc,
		"EdgeGw":            testConfig.Nsxt.EdgeGateway,
		"NsxtImportSegment": testConfig.Nsxt.NsxtImportSegment,
		"NetworkName":       t.Name(),
		"Tags":              "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-DS"
	configText := templateFill(testAccVcdNetworkImportedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-DS-step2"
	configText2 := templateFill(testAccVcdNetworkImportedNsxtDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText)

	params["FuncName"] = t.Name() + "-DS-step3"
	configText3 := templateFill(testAccVcdNetworkImportedNsxtDSStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-DS-step4"
	configText4 := templateFill(testAccVcdNetworkImportedNsxtDSStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-imported-test-initial"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id")),
			},

			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_nsxt_network_imported.net1", "data.vcd_nsxt_network_imported.ds", []string{"%", "nsxt_logical_switch_name"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_nsxt_network_imported.net1", "data.vcd_nsxt_network_imported.ds", []string{"%", "nsxt_logical_switch_name"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_nsxt_network_imported.net1", "data.vcd_nsxt_network_imported.ds", []string{"%", "nsxt_logical_switch_name"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkImportedNsxtDS = testAccVcdNetworkImportedV2NsxtStep1 + `
data "vcd_nsxt_network_imported" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "nsxt-imported-test-initial"
}
`

const testAccVcdNetworkImportedNsxtDSStep3 = testAccVcdNetworkImportedV2NsxtStep1 + `
data "vcd_nsxt_network_imported" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  filter {
	name_regex = "^nsxt-imported"
  }
}
`

const testAccVcdNetworkImportedNsxtDSStep4 = testAccVcdNetworkImportedV2NsxtStep1 + `
data "vcd_nsxt_network_imported" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  filter {
	ip = "1.1.1"
  }
}
`
