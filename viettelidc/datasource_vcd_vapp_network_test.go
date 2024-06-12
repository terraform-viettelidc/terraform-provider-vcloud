//go:build vm || ALL || functional

package viettelidc

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVappNetworkDS tests a vApp network data source if a vApp is found in the VDC
// Supports only NSX-V
func TestAccVcdVappNetworkDS(t *testing.T) {
	preTestChecks(t)
	networkName := "TestAccVcdVappNetworkDS"
	description := "Created in test"
	const gateway = "192.168.0.1"
	const netmask = "255.255.255.0"
	const dns1 = "8.8.8.8"
	const dns2 = "1.1.1.1"
	const dnsSuffix = "biz.biz"
	const startAddress = "192.168.0.10"
	const endAddress = "192.168.0.20"
	const dhcpStartAddress = "192.168.0.30"
	const dhcpEndAddress = "192.168.0.40"
	const maxLeaseTime = 3500
	const defaultLeaseTime = 2400
	var guestVlanAllowed = true
	var retainIpMacEnabled = true

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"VDC":                testConfig.VCD.Vdc,
		"vappName":           "TestAccVcdVappNetworkDS",
		"FuncName":           "TestAccVcdVappNetworkDS",
		"vappNetworkName":    networkName,
		"description":        description,
		"gateway":            gateway,
		"netmask":            netmask,
		"dns1":               dns1,
		"dns2":               dns2,
		"dnsSuffix":          dnsSuffix,
		"guestVlanAllowed":   guestVlanAllowed,
		"startAddress":       startAddress,
		"endAddress":         endAddress,
		"maxLeaseTime":       maxLeaseTime,
		"defaultLeaseTime":   defaultLeaseTime,
		"dhcpStartAddress":   dhcpStartAddress,
		"dhcpEndAddress":     dhcpEndAddress,
		"dhcpEnabled":        "true",
		"orgNetwork":         "TestAccVcdVappNetworkDSOrgNetwork",
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"retainIpMacEnabled": retainIpMacEnabled,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(datasourceTestVappNetwork, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("netmask", netmask),
					resource.TestCheckOutput("description", description),
					resource.TestCheckOutput("gateway", gateway),
					resource.TestCheckOutput("dns1", dns1),
					resource.TestCheckOutput("dns2", dns2),
					resource.TestCheckOutput("dnsSuffix", dnsSuffix),
					resource.TestCheckOutput("dhcpStartAddress", dhcpStartAddress),
					resource.TestCheckOutput("dhcpEndAddress", dhcpEndAddress),
					resource.TestCheckOutput("staticIpPoolStartAddress", startAddress),
					resource.TestCheckOutput("staticIpPoolEndAddress", endAddress),
					resource.TestCheckOutput("orgNetwork", params["orgNetwork"].(string)),
					testCheckVappNetworkNonStringOutputs(guestVlanAllowed, retainIpMacEnabled),
				),
			},
		},
	})
	postTestChecks(t)
}

func testCheckVappNetworkNonStringOutputs(guestVlanAllowed, retainIpMacEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["guestVlanAllowed"].Value != fmt.Sprintf("%v", guestVlanAllowed) {
			return fmt.Errorf("guestVlanAllowed value didn't match")
		}

		if outputs["retain_ip_mac_enabled"].Value != fmt.Sprintf("%v", retainIpMacEnabled) {
			return fmt.Errorf("retain_ip_mac_enabled value didn't match")
		}

		return nil
	}
}

const datasourceTestVappNetwork = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.VDC}}"
}

resource "vcd_network_routed" "{{.orgNetwork}}" {
  name         = "{{.orgNetwork}}"
  org          = "{{.Org}}"
  vdc          = "{{.VDC}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp_network" "createdVappNetwork" {
  org                = "{{.Org}}"
  vdc                = "{{.VDC}}"
  name               = "{{.vappNetworkName}}"
  description        = "{{.description}}"
  vapp_name          = vcd_vapp.{{.vappName}}.name
  gateway            = "{{.gateway}}"
  netmask            = "{{.netmask}}"
  dns1               = "{{.dns1}}"
  dns2               = "{{.dns2}}"
  dns_suffix         = "{{.dnsSuffix}}"
  guest_vlan_allowed = {{.guestVlanAllowed}}

  static_ip_pool {
    start_address = "{{.startAddress}}"
    end_address   = "{{.endAddress}}"
  }

  dhcp_pool {
    max_lease_time     = "{{.maxLeaseTime}}"
    default_lease_time = "{{.defaultLeaseTime}}"
    start_address      = "{{.dhcpStartAddress}}"
    end_address        = "{{.dhcpEndAddress}}"
    enabled            = "{{.dhcpEnabled}}"
  }

  org_network_name      = vcd_network_routed.{{.orgNetwork}}.name
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"
}
 

data "vcd_vapp_network" "network-ds" {
  org                = "{{.Org}}"
  vdc                = "{{.VDC}}"

  name       =  vcd_vapp_network.createdVappNetwork.name
  vapp_name  = "{{.vappName}}"
  depends_on =  [vcd_vapp_network.createdVappNetwork]
}

output "netmask" {
  value = data.vcd_vapp_network.network-ds.netmask 
} 
output "description" {
  value = data.vcd_vapp_network.network-ds.description 
} 
output "gateway" {
  value = data.vcd_vapp_network.network-ds.gateway 
} 
output "dns1" {
  value = data.vcd_vapp_network.network-ds.dns1 
} 
output "dns2" {
  value = data.vcd_vapp_network.network-ds.dns2 
} 
output "dnsSuffix" {
  value = data.vcd_vapp_network.network-ds.dns_suffix 
} 
output "guestVlanAllowed" {
  value = data.vcd_vapp_network.network-ds.guest_vlan_allowed
} 
output "dhcpStartAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.dhcp_pool)[0].start_address
}
output "dhcpEndAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.dhcp_pool)[0].end_address
}
output "staticIpPoolStartAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.static_ip_pool)[0].start_address
}
output "staticIpPoolEndAddress" {
  value  = tolist(data.vcd_vapp_network.network-ds.static_ip_pool)[0].end_address
}
output "orgNetwork" {
  value = data.vcd_vapp_network.network-ds.org_network_name
} 
output "retain_ip_mac_enabled" {
  value = data.vcd_vapp_network.network-ds.retain_ip_mac_enabled
}
`
