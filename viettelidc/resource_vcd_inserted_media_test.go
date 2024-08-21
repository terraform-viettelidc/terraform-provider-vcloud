//go:build catalog || ALL || functional

package viettelidc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var TestAccVcdMediaInsert = "TestAccVcdMediaInsertBasic"
var vappNameForInsert = "TestAccVcdVAppForInsert"
var vmNameForInsert = "TestAccVcdVAppVmForInsert"
var TestAccVcdCatalogMediaForInsert = "TestAccVcdCatalogMediaBasicForInsert"
var TestAccVcdCatalogMediaDescriptionForInsert = "TestAccVcdCatalogMediaBasicDescriptionForInsert"
var TestAccVcdVAppVmNetForInsert = "TestAccVcdVAppVmNetForInsert"

func TestAccVcdMediaInsertBasic(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":                        testConfig.VCD.Org,
		"Vdc":                        testConfig.Nsxt.Vdc,
		"EdgeGateway":                testConfig.Nsxt.EdgeGateway,
		"Catalog":                    testSuiteCatalogName,
		"CatalogItem":                testSuiteCatalogOVAItem,
		"VappName":                   vappNameForInsert,
		"VmName":                     vmNameForInsert,
		"NsxtBackedCataloName":       testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"NsxtBackedCatalogMediaName": testConfig.Media.NsxtBackedMediaName,
		"CatalogMediaName":           TestAccVcdCatalogMediaForInsert,
		"Description":                TestAccVcdCatalogMediaDescriptionForInsert,
		"MediaPath":                  testConfig.Media.MediaPath,
		"UploadPieceSize":            testConfig.Media.UploadPieceSize,
		"UploadProgress":             testConfig.Media.UploadProgress,
		"InsertMediaName":            TestAccVcdMediaInsert,
		"NetworkName":                TestAccVcdVAppVmNetForInsert,
		"EjectForce":                 true,
		"Tags":                       "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdInsertEjectBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccResourcesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInserted("vcd_inserted_media."+TestAccVcdMediaInsert, params),
					testAccCheckMediaEjected("vcd_inserted_media."+TestAccVcdMediaInsert, params),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckMediaInserted(itemName string, params StringMap) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no media insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(params["VappName"].(string), false)
		if err != nil {
			return err
		}

		vm, err := vapp.GetVMByName(params["VmName"].(string), false)

		if err != nil {
			return err
		}

		for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
			if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
				return nil
			}
		}
		return fmt.Errorf("no media inserted found")
	}
}

func testAccCheckMediaEjected(itemName string, params StringMap) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no media insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(params["VappName"].(string), false)
		if err != nil {
			return err
		}

		vm, err := vapp.GetVMByName(params["VmName"].(string), false)

		if err != nil {
			return err
		}

		for _, hardwareItem := range vm.VM.VirtualHardwareSection.Item {
			if hardwareItem.ResourceSubType == types.VMsCDResourceSubType {
				return nil
			}
		}
		return fmt.Errorf("no media inserted found")
	}
}

func testAccResourcesDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		itemName := rs.Primary.Attributes["name"]
		if rs.Type != "vcd_inserted_media" && itemName != TestAccVcdMediaInsert {
			continue
		}

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetVAppByName(vappNameForInsert, false)
		if err == nil {
			return fmt.Errorf("vapp %s still exist", itemName)
		}

		_, err = vdc.GetOrgVdcNetworkByName(TestAccVcdVAppVmNetForInsert, false)
		if err == nil {
			return fmt.Errorf("network %s still exist and error: %#v", itemName, err)
		}
	}
	return nil
}

const testAccCheckVcdInsertEjectBasic = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  name       = "{{.VappName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed.{{.NetworkName}}.name 
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  power_on      = "false"
  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "POOL"
  }
}

resource "vcd_catalog_media" "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = "{{.NsxtBackedCataloName}}"
  name    = "{{.CatalogMediaName}}"

  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}

resource "vcd_inserted_media" "{{.InsertMediaName}}" {
  org     = "{{.Org}}"
  vdc     = "{{.Vdc}}"
  catalog = "{{.NsxtBackedCataloName}}"
  name    = "{{.CatalogMediaName}}"

  vapp_name  = vcd_vapp.{{.VappName}}.name
  vm_name    = vcd_vapp_vm.{{.VmName}}.name
  depends_on = [vcd_vapp_vm.{{.VmName}}, vcd_catalog_media.{{.CatalogMediaName}}]

  eject_force = "{{.EjectForce}}"
}
`

func TestAccVcdMediaLockTest(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":                        testConfig.VCD.Org,
		"Vdc":                        testConfig.Nsxt.Vdc,
		"EdgeGateway":                testConfig.Nsxt.EdgeGateway,
		"Catalog":                    testConfig.VCD.Catalog.Name,
		"CatalogItem":                testConfig.VCD.Catalog.CatalogItem,
		"VappName":                   t.Name(),
		"VmName":                     t.Name(),
		"NsxtBackedCataloName":       testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"NsxtBackedCatalogMediaName": testConfig.Media.NsxtBackedMediaName,
		"Description":                t.Name() + "_description",
		"MediaPath":                  testConfig.Media.MediaPath,
		"UploadPieceSize":            testConfig.Media.UploadPieceSize,
		"UploadProgress":             testConfig.Media.UploadProgress,
		"NetworkName":                t.Name(),
		"EjectForce":                 true,
		"Tags":                       "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdInsertEjectLock, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccResourcesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMediaInserted("vcd_inserted_media.test", params),
					testAccCheckMediaEjected("vcd_inserted_media.test", params),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdInsertEjectLock = `
resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp" "{{.VappName}}" {
  name       = "{{.VappName}}"
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VappName}}.name
  org_network_name   = vcd_network_routed.{{.NetworkName}}.name 
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  power_on      = "false"
  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "POOL"
  }
}

resource "vcd_inserted_media" "test" {
  org     = "{{.Org}}"
  vdc     = "{{.Vdc}}"
  catalog = "{{.NsxtBackedCataloName}}"
  name    = "{{.NsxtBackedCatalogMediaName}}"

  vapp_name  = vcd_vapp.{{.VappName}}.name
  vm_name    = vcd_vapp_vm.{{.VmName}}.name

  eject_force = "{{.EjectForce}}"

}

resource "vcd_vm_internal_disk" "internal_disk1" {
  org     = "{{.Org}}"
  vdc     = "{{.Vdc}}"

  vapp_name  = vcd_vapp.{{.VappName}}.name
  vm_name    = vcd_vapp_vm.{{.VmName}}.name

  bus_type    = "paravirtual"
  size_in_mb  = 10 * 1024
  bus_number  = 1
  unit_number = 0

}
`
