//go:build vapp || vm || ALL || functional

package viettelidc

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmDhcpWait(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"EdgeGateway":     testConfig.Networking.EdgeGateway,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VAppName":        netVappName,
		"VMName":          netVmName1,
		"Tags":            "vapp vm",
		"DhcpWaitSeconds": 300,
	}
	testParamsNotEmpty(t, params)

	configTextVM := templateFill(testAccCheckVcdVAppVmDhcpWait, params)

	params["FuncName"] = t.Name() + "-step2"
	params["DhcpWaitSeconds"] = 310
	configTextVMDhcpWaitUpdateStep2 := templateFill(testAccCheckVcdVAppVmDhcpWait, params)

	// A step to Power off vApp and VM
	params["FuncName"] = t.Name() + "-step3"
	configTextVMDhcpWaitUpdateStep3 := templateFill(testAccCheckVcdVAppVmDhcpWaitStep3, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	reIp := regexp.MustCompile(`^11.10.0.\d{1,3}$`)
	skipEnvVar := "VCD_SKIP_DHCP_CHECK"
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create with variations of all possible NICs
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "DHCP"),
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", reIp)),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.connected", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_dhcp_wait_seconds", "300"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.connected", "false"),

					// Check data source
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("data.vcd_vapp_vm.ds", "network.0.ip", reIp)),

					skipOnEnvVariable(skipEnvVar, "1", "comparing IPs",
						resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName1, "network.0.ip", "data.vcd_vapp_vm.ds", "network.0.ip")),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.ds", "network_dhcp_wait_seconds", "300"),
				),
			},
			{
				Config: configTextVMDhcpWaitUpdateStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.name", "multinic-net"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip_allocation_mode", "DHCP"),
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("vcd_vapp_vm."+netVmName1, "network.0.ip", reIp)),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+netVmName1, "network.0.mac"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network_dhcp_wait_seconds", "310"),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.ip_allocation_mode", "NONE"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.is_primary", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "network.1.connected", "false"),

					// Check data source
					skipOnEnvVariable(skipEnvVar, "1", "IP regexp "+reIp.String(),
						resource.TestMatchResourceAttr("data.vcd_vapp_vm.ds", "network.0.ip", reIp)),
					skipOnEnvVariable(skipEnvVar, "1", "comparing IPs",
						resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName1, "network.0.ip", "data.vcd_vapp_vm.ds", "network.0.ip")),
					resource.TestCheckResourceAttr("data.vcd_vapp_vm.ds", "network_dhcp_wait_seconds", "310"),
				),
			},
			{
				Config: configTextVMDhcpWaitUpdateStep3,
			},
		},
	})
	postTestChecks(t)
}

// #nosec G101 -- This doesn't contain any credential
const testAccCheckVcdVAppVmDhcpWaitShared = `
# skip-binary-test: vApp networks cannot be removed in a powered on vApp 
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

const testAccCheckVcdVAppVmDhcpWait = testAccCheckVcdVAppVmDhcpWaitShared + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  computer_name = "dhcp-vm"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network_dhcp_wait_seconds = {{.DhcpWaitSeconds}}
  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappNetwork1.org_network_name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
 
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = "false"
  }
}

data "vcd_vapp_vm" "ds" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name                 = vcd_vapp.{{.VAppName}}.name
  name                      = vcd_vapp_vm.{{.VMName}}.name
  network_dhcp_wait_seconds = {{.DhcpWaitSeconds}}
  depends_on                = [vcd_vapp_vm.{{.VMName}}]
}
`

// #nosec G101 -- This doesn't contain any credential
const testAccCheckVcdVAppVmDhcpWaitStep3 = `
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
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  vapp_name        = vcd_vapp.{{.VAppName}}.name
  org_network_name = vcd_network_routed.net.name 
}

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  power_on      = false
  computer_name = "dhcp-vm"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
}
`
