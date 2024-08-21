//go:build vapp || ALL || functional

package viettelidc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppRaw_Basic(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"EdgeGateway": testConfig.Nsxt.EdgeGateway,
		"NetworkName": "TestAccVcdVAppRawNet",
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VappName":    "TestAccVcdVAppRawVapp",
		"VmName":      "TestAccVcdVAppRawVm",
		"Tags":        "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVAppRaw_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppRawDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppRawExists(fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), &vapp),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("vcd_vapp.%s", params["VappName"].(string)), "name", params["VappName"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdVAppRawExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		newVapp, err := vdc.GetVAppByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return err
		}

		*vapp = *newVapp

		return nil
	}
}

func testAccCheckVcdVAppRawDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetVAppByNameOrId(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVAppRaw_basic = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.VappName}}"
  depends_on   = ["vcd_network_routed_v2.{{.NetworkName}}"]
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed_v2.{{.NetworkName}}.name 
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = false
  memory        = 1024
  cpus          = 1

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "POOL"
  }
}
`
