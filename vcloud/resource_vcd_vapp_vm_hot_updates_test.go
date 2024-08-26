//go:build vapp || vm || ALL || functional

package vcloud

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppHotUpdateVm(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		hotVappName string = t.Name()
		hotVmName1  string = t.Name() + "VM"
	)

	if testConfig.Media.MediaName == "" {
		fmt.Println("Warning: `MediaName` is not configured: boot image won't be tested.")
	}

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VAppName":        hotVappName,
		"VMName":          hotVmName1,
		"Tags":            "vapp vm",
		"Media":           testConfig.Media.MediaName,
		"StorageProfile":  testConfig.VCD.ProviderVdc.StorageProfile,
		"StorageProfile2": testConfig.VCD.ProviderVdc.StorageProfile2,
	}
	testParamsNotEmpty(t, params)

	configTextVM := templateFill(testAccCheckVcdVAppHotUpdateVm, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppHotUpdateVmStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep1)

	params["FuncName"] = t.Name() + "-step2"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppHotUpdateVmStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep2)

	params["FuncName"] = t.Name() + "-step3"
	configTextVMUpdateStep3 := templateFill(testAccCheckVcdVAppHotUpdateVmStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep3)

	params["FuncName"] = t.Name() + "-step4"
	configTextVMUpdateStep4 := templateFill(testAccCheckVcdVAppHotUpdateVmStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep4)

	params["FuncName"] = t.Name() + "-step5"
	configTextVMUpdateStep5 := templateFill(testAccCheckVcdVAppHotUpdateVmStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep5)

	params["FuncName"] = t.Name() + "-step6"
	configTextVMUpdateStep6 := templateFill(testAccCheckVcdVAppHotUpdateVmStep6PowerOff, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVMUpdateStep6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var step5Check resource.TestCheckFunc

	step4func := resource.TestStep{
		Config: configTextVMUpdateStep4,
		Check: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdVmNotRestarted("vcd_vapp_vm."+hotVmName1, hotVappName, hotVmName1),
		),
	}
	step5Check = testAccCheckVcdVmNotRestarted("vcd_vapp_vm."+hotVmName1, hotVappName, hotVmName1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(hotVappName),
		Steps: []resource.TestStep{
			// Step 0 - create
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "2048"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "1"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+hotVmName1, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "metadata.mediaItem_metadata", "data 1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "metadata.mediaItem_metadata2", "data 2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "metadata.mediaItem_metadata3", "data 3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `guest_properties.guest.another.subkey`, "another-value"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `storage_profile`, params["StorageProfile"].(string)),

					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`), ">= 38.1"),
					testCheckResourceAttrSetWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.name", ">= 38.1"),
					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`), ">= 38.1"),
				),
			},
			// Step 1 - update - network changes
			{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+hotVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "metadata.mediaItem_metadata", "data 1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "metadata.mediaItem_metadata2", "data 3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `guest_properties.guest.hostname`, "test-host2"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `storage_profile`, params["StorageProfile2"].(string)),

					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`), ">= 38.1"),
					testCheckResourceAttrSetWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.name", ">= 38.1"),
					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`), ">= 38.1"),

					testAccCheckVcdVmNotRestarted("vcd_vapp_vm."+hotVmName1, hotVappName, hotVmName1),
				),
			},
			// Step 2 - update
			{
				Config:      configTextVMUpdateStep2,
				ExpectError: regexp.MustCompile(`update stopped: VM needs to power off to change properties.*`),
			},
			// Step 3 - update - add new network section
			{
				Config: configTextVMUpdateStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+hotVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.2.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.2.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.2.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `storage_profile`, params["StorageProfile2"].(string)),

					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`), ">= 38.1"),
					testCheckResourceAttrSetWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.name", ">= 38.1"),
					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`), ">= 38.1"),

					testAccCheckVcdVmNotRestarted("vcd_vapp_vm."+hotVmName1, hotVappName, hotVmName1),
				),
			},
			// Step 4 - update - remove network section
			step4func,
			// Step 5 - update - network changes
			{
				Config: configTextVMUpdateStep5,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(hotVappName, hotVmName1, "vcd_vapp_vm."+hotVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "name", hotVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory_hot_add_enabled", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "memory", "3072"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "cpus", "3"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.0.connected", "false"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+hotVmName1, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, "network.1.connected", "true"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+hotVmName1, `storage_profile`, params["StorageProfile2"].(string)),

					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`), ">= 38.1"),
					testCheckResourceAttrSetWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.name", ">= 38.1"),
					testMatchResourceAttrWhenVersionMatches("vcd_vapp_vm."+hotVmName1, "inherited_metadata.vm.origin.type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`), ">= 38.1"),

					step5Check,
				),
			},
			{
				Config: configTextVMUpdateStep6,
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdVmNotRestarted(n string, vappName, vmName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		org, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(vappName, false)
		if err != nil {
			return err
		}

		vm, err := vapp.GetVMByName(vmName, false)
		if err != nil {
			return err
		}

		tasks, err := org.GetTaskList()
		if err != nil {
			return err
		}

		for _, task := range tasks.Task {
			if strings.Contains(task.Operation, "Stopped") && task.Owner.ID == vm.VM.ID {
				return fmt.Errorf("found task which stopped VM")
			}
		}

		return nil
	}
}

const testSharedHotUpdate = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name     = "{{.VAppName}}"
  power_on = true
}

resource "vcd_network_routed" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "multinic-net"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "11.10.0.1"

  dhcp_pool {
    start_address = "11.10.0.2"
    end_address   = "11.10.0.100"
  }

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.254"
  }
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VAppName}}.name
  org_network_name   = vcd_network_routed.net.name 
}
`

const testAccCheckVcdVAppHotUpdateVm = testSharedHotUpdate + `
# skip-binary-test: cannot remove vApp network from a powered on vApp
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "compNameUp"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"

  memory        = 2048
  cpus          = 1

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }
 
  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }

  metadata = {
    mediaItem_metadata = "data 1"
    mediaItem_metadata2 = "data 2"
    mediaItem_metadata3 = "data 3"
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }

  storage_profile = "{{.StorageProfile}}"
 }
`

const testAccCheckVcdVAppHotUpdateVmStep1 = `# skip-binary-test: only for updates
` + testSharedHotUpdate + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  metadata = {
    mediaItem_metadata = "data 1"
    mediaItem_metadata2 = "data 3"
  }

  guest_properties = {
	"guest.hostname"       = "test-host2"
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdVAppHotUpdateVmStep2 = `# skip-binary-test: only for updates
` + testSharedHotUpdate + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = false
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  storage_profile = "{{.StorageProfile2}}"
}
`
const testAccCheckVcdVAppHotUpdateVmStep3 = `# skip-binary-test: only for updates
` + testSharedHotUpdate + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdVAppHotUpdateVmStep4 = `# skip-binary-test: only for updates
` + testSharedHotUpdate + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = true

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = true
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = false
  }

  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdVAppHotUpdateVmStep5 = `# skip-binary-test: only for updates
` + testSharedHotUpdate + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = false

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = false
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
 
  storage_profile = "{{.StorageProfile2}}"
}
`

const testAccCheckVcdVAppHotUpdateVmStep6PowerOff = `
resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name     = "{{.VAppName}}"
  power_on = false
}

resource "vcd_network_routed" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name         = "multinic-net"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "11.10.0.1"

  dhcp_pool {
    start_address = "11.10.0.2"
    end_address   = "11.10.0.100"
  }

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.254"
  }
}

resource "vcd_vapp_org_network" "vappNetwork1" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = vcd_vapp.{{.VAppName}}.name
  org_network_name   = vcd_network_routed.net.name 
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  computer_name = "compNameUp"
  name          = "{{.VMName}}"
  power_on      = false

  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
 
  memory        = 3072
  cpus          = 3

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  prevent_update_power_off = false

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
    is_primary         = false
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
 
  storage_profile = "{{.StorageProfile2}}"
}
`
