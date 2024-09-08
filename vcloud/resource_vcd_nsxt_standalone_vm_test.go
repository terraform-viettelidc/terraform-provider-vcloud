//go:build (vm || nsxt || standaloneVm || ALL || functional) && !skipStandaloneVm

package vcloud

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	testingTags["standaloneVm"] = "resource_vcd_vapp_vm_test.go"
}

// TestAccVcdNsxtStandaloneVmTemplate tests NSX-T Routed network DHCP pools, static pools and manual IP assignment
// Note. This test triggers a bug in 10.2.2.17855680 and fails. Because of this reason it is skipped on exactly this
// version.
// Note. Usage of vcd_nsxt_network_imported network in VM, using CDS triggers:
// "The operation failed because no suitable resource was found. Out of 0 candidate hubs: NO_FEASIBLE_PLACEMENT_SOLUTION"
// which observed in UI also
func TestAccVcdNsxtStandaloneVmTemplate(t *testing.T) {
	preTestChecks(t)
	if noTestCredentials() {
		t.Skip("Skipping test run as no credentials are provided and this test needs to lookup VCD version")
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can create Imported networks")
	}

	skipTestForVcdExactVersion(t, "10.2.2.17855680", "removal of standalone VM with NICs fails")

	// making sure the VM name is unique
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())
	var diskResourceName = fmt.Sprintf("%s_disk", t.Name())
	var diskName = fmt.Sprintf("%s-disk", t.Name())

	orgName := testConfig.VCD.Org
	vdcName := testConfig.Nsxt.Vdc
	var params = StringMap{
		"Org":                orgName,
		"Vdc":                vdcName,
		"EdgeGateway":        testConfig.Nsxt.EdgeGateway,
		"NetworkName":        "TestAccVcdNsxtStandaloneVmNet",
		"ImportSegment":      testConfig.Nsxt.NsxtImportSegment,
		"Catalog":            testSuiteCatalogName,
		"CatalogItem":        testSuiteCatalogOVAItem,
		"VmName":             standaloneVmName,
		"ComputerName":       standaloneVmName + "-unique",
		"diskName":           diskName,
		"size":               "5",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": "*",
		"diskResourceName":   diskResourceName,
		"Tags":               "vm standaloneVm nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdNsxtStandaloneVm_basic, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccCheckVcdNsxtStandaloneVm_basicCleanup, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, orgName, vdcName),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, orgName, vdcName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "vm_type", string(standaloneVmType)),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "description", "test standalone VM"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "computer_name", standaloneVmName+"-unique"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "network.0.ip", "10.10.102.161"),
					// Due to DHCP nature it is not guaranteed to be reported quickly enough
					// resource.TestCheckResourceAttrSet(
					// 	"vcd_vm."+standaloneVmName, "network.1.ip"),
					resource.TestMatchResourceAttr(
						"vcd_vm."+standaloneVmName, "network.2.ip", regexp.MustCompile(`^10\.10\.102\.`)),
					resource.TestMatchResourceAttr(
						"vcd_vm."+standaloneVmName, "network.3.ip", regexp.MustCompile(`^110\.10\.102\.`)),
					resource.TestMatchResourceAttr(
						"vcd_vm."+standaloneVmName, "network.4.ip", regexp.MustCompile(`^12\.12\.2\.`)),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "metadata.vm_metadata", "VM Metadata."),
					resource.TestCheckOutput("disk", diskName),
					resource.TestCheckOutput("disk_bus_number", "1"),
					resource.TestCheckOutput("disk_unit_number", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm."+standaloneVmName, "disk.*", map[string]string{
						"size_in_mb": "5",
					}),
					testMatchResourceAttrWhenVersionMatches("vcd_vm."+standaloneVmName, "inherited_metadata.vm.origin.id", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`), ">= 38.1"),
					testCheckResourceAttrSetWhenVersionMatches("vcd_vm."+standaloneVmName, "inherited_metadata.vm.origin.name", ">= 38.1"),
					testMatchResourceAttrWhenVersionMatches("vcd_vm."+standaloneVmName, "inherited_metadata.vm.origin.type", regexp.MustCompile(`^com\.vmware\.vcloud\.entity\.\w+$`), ">= 38.1"),
				),
			},
			{
				ResourceName:      "vcd_vm." + standaloneVmName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(standaloneVmName),
				// These fields can't be retrieved from user data
				// "network.1.ip" is a DHCP value. It may happen so that during "read" IP is still not reported, but
				// it is reported during import
				// "network_dhcp_wait_seconds" is a user setting and cannot be imported
				ImportStateVerifyIgnore: []string{"template_name", "catalog_name",
					"accept_all_eulas", "power_on", "computer_name", "prevent_update_power_off", "network.1.ip",
					"network_dhcp_wait_seconds", "consolidate_disks_on_create", "imported", "vapp_template_id"},
			},
			// This step ensures that VM and disk are removed, but networks are left
			{
				Config: configText2,
			},
			// This step gives 10-second sleep timer so that cleanup bug is not hit in VCD 10.3
			{
				Config:    configText2,
				PreConfig: func() { time.Sleep(10 * time.Second) },
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdNsxtStandaloneEmptyVm(t *testing.T) {
	preTestChecks(t)

	// making sure the VM name is unique
	standaloneVmName := fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	if testConfig.Media.NsxtBackedMediaName == "" {
		fmt.Println("Warning: `NsxtBackedMediaName` is not configured: boot image won't be tested.")
	}

	orgName := testConfig.VCD.Org
	vdcName := testConfig.Nsxt.Vdc
	var params = StringMap{
		"Org":         orgName,
		"Vdc":         vdcName,
		"EdgeGateway": testConfig.Nsxt.EdgeGateway,
		"Catalog":     testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"VMName":      standaloneVmName,
		"Tags":        "vm standaloneVm",
		"Media":       testConfig.Media.NsxtBackedMediaName,
	}
	testParamsNotEmpty(t, params)

	// Create objects for testing field values across update steps
	nic0Mac := testCachedFieldValue{}
	nic1Mac := testCachedFieldValue{}

	configTextVM := templateFill(testAccCheckVcdNsxtStandaloneEmptyVm, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, orgName, vdcName),
		Steps: []resource.TestStep{
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, orgName, vdcName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.name", "multinic-net2"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.ip", "12.10.0.152"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.adapter_type", "PCNet32"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.0.connected", "true"),
					nic0Mac.cacheTestResourceFieldValue("vcd_vm."+standaloneVmName, "network.0.mac"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttrSet("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "network.1.adapter_type", "VMXNET3"),
					nic1Mac.cacheTestResourceFieldValue("vcd_vm."+standaloneVmName, "network.1.mac"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "description", "test empty standalone VM"),

					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "cpu_hot_add_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "memory_hot_add_enabled", "true"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdNsxtStandaloneVm_basic = `
# skip-binary-test: removing NSX-T Org networks right after standalone VM fails in 10.3
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.NetworkName}}" {
  name            = "{{.NetworkName}}"
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.200"
  }

  depends_on = [vcd_nsxt_network_imported.imported-test]
}

resource "vcd_network_isolated_v2" "net-test" {
  name            = "{{.NetworkName}}-isolated"
  org             = "{{.Org}}"
  
  gateway         = "110.10.102.1"
  prefix_length   = 26

  static_ip_pool {
    start_address = "110.10.102.2"
    end_address   = "110.10.102.20"
  }
}

resource "vcd_nsxt_network_dhcp" "{{.NetworkName}}-dhcp" {
  org             = "{{.Org}}"
  
  org_network_id  = vcd_network_routed_v2.{{.NetworkName}}.id

  pool {
    start_address = "10.10.102.210"
    end_address   = "10.10.102.220"
  }

  pool {
    start_address = "10.10.102.230"
    end_address   = "10.10.102.240"
  }
}

resource "vcd_nsxt_network_imported" "imported-test" {
  name            = "{{.NetworkName}}-imported"
  org             = "{{.Org}}"
  gateway         = "12.12.2.1"
  prefix_length   = 24

  nsxt_logical_switch_name = "{{.ImportSegment}}"

  static_ip_pool {
    start_address = "12.12.2.10"
    end_address   = "12.12.2.15"
  }

  depends_on = [vcd_network_isolated_v2.net-test]
}

resource "vcd_independent_disk" "{{.diskResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.diskName}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"

  depends_on = [vcd_network_routed_v2.{{.NetworkName}}, vcd_network_isolated_v2.net-test, 
                vcd_nsxt_network_imported.imported-test]
}

resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  description   = "test standalone VM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  storage_profile = "{{.storageProfileName}}"

  metadata = {
    vm_metadata = "VM Metadata."
  }

  # This is short enough to not wait for IP retrieval, but is useful to check that
  # DHCP lookup functions internally do not fail. It may emit WARNING message during
  # test run, but these are just to inform that 10 seconds is too short time.
  network_dhcp_wait_seconds = 10

  network {
    type               = "org"
    name               = vcd_network_routed_v2.{{.NetworkName}}.name
    ip_allocation_mode = "MANUAL"
    ip                 = "10.10.102.161"
  }

  network {
    type               = "org"
    name               = vcd_network_routed_v2.{{.NetworkName}}.name
    ip_allocation_mode = "DHCP"
  }

  network {
    type               = "org"
    name               = vcd_network_routed_v2.{{.NetworkName}}.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name               = vcd_network_isolated_v2.net-test.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name               = vcd_nsxt_network_imported.imported-test.name
    ip_allocation_mode = "POOL"
  }

  disk {
    name        = vcd_independent_disk.{{.diskResourceName}}.name
    bus_number  = 1
    unit_number = 0
  }

  depends_on = [vcd_independent_disk.{{.diskResourceName}}]
}

output "disk" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].name
}
output "disk_bus_number" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].bus_number
}
output "disk_unit_number" {
  value = tolist(vcd_vm.{{.VmName}}.disk)[0].unit_number
}
output "vm" {
  value = vcd_vm.{{.VmName}}
  
  sensitive = true
}
`

const testAccCheckVcdNsxtStandaloneVm_basicCleanup = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "{{.NetworkName}}" {
  name            = "{{.NetworkName}}"
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.200"
  }

  depends_on = [vcd_nsxt_network_imported.imported-test]
}

resource "vcd_network_isolated_v2" "net-test" {
  name            = "{{.NetworkName}}-isolated"
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  
  gateway         = "110.10.102.1"
  prefix_length   = 26

  static_ip_pool {
    start_address = "110.10.102.2"
    end_address   = "110.10.102.20"
  }
}

resource "vcd_nsxt_network_dhcp" "{{.NetworkName}}-dhcp" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  
  org_network_id  = vcd_network_routed_v2.{{.NetworkName}}.id

  pool {
    start_address = "10.10.102.210"
    end_address   = "10.10.102.220"
  }

  pool {
    start_address = "10.10.102.230"
    end_address   = "10.10.102.240"
  }
}

resource "vcd_nsxt_network_imported" "imported-test" {
  name            = "{{.NetworkName}}-imported"
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  gateway         = "12.12.2.1"
  prefix_length   = 24

  nsxt_logical_switch_name = "{{.ImportSegment}}"

  static_ip_pool {
    start_address = "12.12.2.10"
    end_address   = "12.12.2.15"
  }

  depends_on = [vcd_network_isolated_v2.net-test]
}
`

const testAccCheckVcdNsxtStandaloneEmptyVmNetworkShared = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.EdgeGateway}}"
}

resource "vcd_network_routed_v2" "net" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name            = "multinic-net"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  prefix_length   = 24
  gateway         = "11.10.0.1"

  static_ip_pool {
    start_address = "11.10.0.152"
    end_address   = "11.10.0.254"
  }
}

resource "vcd_network_routed_v2" "net2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name            = "multinic-net2"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "12.10.0.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "12.10.0.152"
    end_address   = "12.10.0.254"
  }
}
`

const testAccCheckVcdNsxtStandaloneEmptyVm = testAccCheckVcdNsxtStandaloneEmptyVmNetworkShared + `
resource "vcd_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  # You cannot remove NICs from an active virtual machine on which no operating system is installed.
  power_on = false

  description   = "test empty standalone VM"
  name          = "{{.VMName}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1 
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  boot_image                     = "{{.Media}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  cpu_hot_add_enabled    = true
  memory_hot_add_enabled = true

  network {
    type               = "org"
    name               = vcd_network_routed_v2.net2.name
    ip_allocation_mode = "POOL"
    is_primary         = false
	  adapter_type       = "PCNet32"
  }

  network {
    type               = "org"
    name               = vcd_network_routed_v2.net.name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }
}
`
