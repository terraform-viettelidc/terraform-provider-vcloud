//go:build network || nsxt || ALL || functional

package viettelidc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdNsxtIpSecVpnTunnel tests out various configurations of IPsec VPN Tunnel configurations without Security
// Profile customization
func TestAccVcdNsxtIpSecVpnTunnel(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtIpSecVpnTunnel1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["ResourceName"] = "test-tunnel-1"
	configText2 := templateFill(testAccNsxtIpSecVpnTunnel1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccNsxtIpSecVpnTunnel2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	params["ResourceName"] = "test-tunnel-1-updated"
	configText4 := templateFill(testAccNsxtIpSecVpnTunnel2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccNsxtIpSecVpnTunnel3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "-step6"
	params["ResourceName"] = "test-tunnel-1"
	configText6 := templateFill(testAccNsxtIpSecVpnTunnel3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["FuncName"] = t.Name() + "-step7"
	configText7 := templateFill(testAccNsxtIpSecVpnTunnel4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	params["FuncName"] = t.Name() + "-step8"
	configText8 := templateFill(testAccNsxtIpSecVpnTunnel4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// ignoreDataSourceFields specifies a field list to ignore for data source comparison with resource. These
	// fields should return the same values, but because 'status' is not controlled and resource and data source are
	// read not at the same time - there is a risk one will have status and other won't
	ignoreDataSourceFields := []string{"status", "ike_service_status", "ike_fail_reason"}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1"),
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "2.2.2.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "3.3.0.0/16"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "2.2.2.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "3.3.0.0/16"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "10.10.10.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "2.2.2.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "3.3.0.0/16"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText8,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "updated-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "50.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "2.3.4.5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "10.10.10.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "1.1.1.1/32"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "2.2.2.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "3.3.0.0/16"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},

			// Test import with IP addresses
			{
				ResourceName:            "vcd_nsxt_ipsec_vpn_tunnel.tunnel1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-tunnel-1"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			// Try to import by UUID
			{
				ResourceName: "vcd_nsxt_ipsec_vpn_tunnel.tunnel1",
				ImportState:  true,
				// Not using pre-built complete ID because ID is not known in advance. This field allows to specify
				// prefix only and the ID itself is automatically suffixed by Terraform test framework
				ImportStateIdPrefix:     testConfig.VCD.Org + ImportSeparator + testConfig.Nsxt.Vdc + ImportSeparator + testConfig.Nsxt.EdgeGateway + ImportSeparator,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtIpSecVpnTunnelDS = `
# skip-binary-test: Data Source test
data "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id
  name            = "{{.ResourceName}}"
}
`

const testAccNsxtIpSecVpnTunnel1 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
}
`
const testAccNsxtIpSecVpnTunnel1DS = testAccNsxtIpSecVpnTunnel1 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnel2 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1-updated"
  
  pre_shared_key    = "updated-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["50.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
}
`

const testAccNsxtIpSecVpnTunnel2DS = testAccNsxtIpSecVpnTunnel2 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnel3 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name    = "test-tunnel-1"
  enabled = false
  
  pre_shared_key    = "updated-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["50.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "2.3.4.5"
  remote_networks   = ["1.1.1.1/32", "2.2.2.0/24", "3.3.0.0/16"]
}
`

const testAccNsxtIpSecVpnTunnel3DS = testAccNsxtIpSecVpnTunnel3 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnel4 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name    = "test-tunnel-1"
  enabled = false
  
  pre_shared_key    = "updated-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["50.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "2.3.4.5"
  remote_id         = "10.10.10.1"
  remote_networks   = ["1.1.1.1/32", "2.2.2.0/24", "3.3.0.0/16"]
}
`

const testAccNsxtIpSecVpnTunnel4DS = testAccNsxtIpSecVpnTunnel4 + testAccNsxtIpSecVpnTunnelDS

func TestAccVcdNsxtIpSecVpnTunnelCustomProfile(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["ResourceName"] = "test-tunnel-1"
	configText2 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	params["ResourceName"] = "test-tunnel-1"
	configText4 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "-step6"
	params["ResourceName"] = "test-tunnel-1"
	configText6 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["FuncName"] = t.Name() + "-step7"
	configText7 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	params["FuncName"] = t.Name() + "-step8"
	params["ResourceName"] = "test-tunnel-1"
	configText8 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	params["FuncName"] = t.Name() + "-step9"
	configText9 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 9: %s", configText9)

	params["FuncName"] = t.Name() + "-step10"
	params["ResourceName"] = "test-tunnel-1"
	configText10 := templateFill(testAccNsxtIpSecVpnTunnelProfileStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 10: %s", configText10)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// ignoreDataSourceFields specifies a field list to ignore for data source comparison with resource. These
	// fields should return the same values, but because 'status' is not controlled and resource and data source are
	// read not at the same time - there is a risk one will have status and other won't
	ignoreDataSourceFields := []string{"status", "ike_service_status", "ike_fail_reason"}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1"),
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					// Security profile customization checks
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile", "CUSTOM"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_version", "IKE_V2"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.*", "AES_128"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_sa_lifetime", "86400"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_pfs_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_df_policy", "COPY"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.*", "AES_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_sa_lifetime", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.dpd_probe_internal", "30"),
				),
			},
			{
				Config: configText2,
				Check:  resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "4"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.30.0/29"),
					// Security profile customization checks
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile", "CUSTOM"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_version", "IKE_V2"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.*", "AES_128"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_sa_lifetime", "86400"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_pfs_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_df_policy", "COPY"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.*", "AES_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_sa_lifetime", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.dpd_probe_internal", "30"),
				),
			},
			{
				Config: configText4,
				Check:  resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile", "CUSTOM"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_version", "IKE_FLEX"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.*", "AES_128"),
				),
			},
			{
				Config: configText6,
				Check:  resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),
			},
			{
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile", "DEFAULT"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText8,
				Check:  resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),
			},
			{
				Config: configText9,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					// Security profile customization checks
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile", "CUSTOM"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_version", "IKE_V2"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_encryption_algorithms.*", "AES_128"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.ike_sa_lifetime", "86400"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_pfs_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_df_policy", "COPY"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_encryption_algorithms.*", "AES_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_digest_algorithms.*", "SHA2_256"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_dh_groups.*", "GROUP14"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.tunnel_sa_lifetime", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.0.dpd_probe_internal", "30"),
				),
			},
			{
				Config: configText10,
				Check:  resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtIpSecVpnTunnelProfileStep1 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address as there is nothing else to peer to
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]

  security_profile_customization {
    ike_version               = "IKE_V2"
    ike_encryption_algorithms = ["AES_128"]
    # ike_encryption_algorithms = ["AES_128", "AES_256"]
    ike_digest_algorithms     = ["SHA2_256"]
    ike_dh_groups             = ["GROUP14"]
    ike_sa_lifetime           = 86400
    
	tunnel_pfs_enabled = true
	tunnel_df_policy = "COPY"
    tunnel_encryption_algorithms = ["AES_256"]
    tunnel_digest_algorithms     = ["SHA2_256"]
    tunnel_dh_groups             = ["GROUP14"]
    tunnel_sa_lifetime           = 3600
    
    dpd_probe_internal = "30"
  }
}
`

const testAccNsxtIpSecVpnTunnelProfileStep1DS = testAccNsxtIpSecVpnTunnelProfileStep1 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnelProfileStep2 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description-updated"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address as there is nothing else to peer to
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28", "192.168.30.0/29"]

  security_profile_customization {
    ike_version               = "IKE_V2"
    ike_encryption_algorithms = ["AES_128"]
    # ike_encryption_algorithms = ["AES_128", "AES_256"]
    ike_digest_algorithms     = ["SHA2_256"]
    ike_dh_groups             = ["GROUP14"]
    ike_sa_lifetime           = 86400
    
	tunnel_pfs_enabled = true
	tunnel_df_policy = "COPY"
    tunnel_encryption_algorithms = ["AES_256"]
    tunnel_digest_algorithms     = ["SHA2_256"]
    tunnel_dh_groups             = ["GROUP14"]
    tunnel_sa_lifetime           = 3600
    
    dpd_probe_internal = "30"
  }
}
`

const testAccNsxtIpSecVpnTunnelProfileStep2DS = testAccNsxtIpSecVpnTunnelProfileStep2 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnelProfileStep3 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address as there is nothing else to peer to
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]

  security_profile_customization {
    ike_version               = "IKE_FLEX"
    ike_encryption_algorithms = ["AES_128"]
    ike_digest_algorithms     = ["SHA2_256"]
    ike_dh_groups             = ["GROUP19"]
    ike_sa_lifetime           = 21600 # 4 hours
    
	tunnel_pfs_enabled = true
	tunnel_df_policy = "COPY"
    tunnel_encryption_algorithms = ["AES_128"]
    tunnel_digest_algorithms     = ["SHA2_512"]
    tunnel_dh_groups             = ["GROUP19"]
    tunnel_sa_lifetime           = 6000 # 10 minutes
    
    dpd_probe_internal = "30"
  }
}
`

const testAccNsxtIpSecVpnTunnelProfileStep3DS = testAccNsxtIpSecVpnTunnelProfileStep3 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnelProfileStep4 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address as there is nothing else to peer to
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]

}
`

const testAccNsxtIpSecVpnTunnelProfileStep4DS = testAccNsxtIpSecVpnTunnelProfileStep4 + testAccNsxtIpSecVpnTunnelDS

const testAccNsxtIpSecVpnTunnelProfileStep5 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address as there is nothing else to peer to
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]

  security_profile_customization {
    ike_version               = "IKE_V2"
    ike_encryption_algorithms = ["AES_128"]
    # ike_encryption_algorithms = ["AES_128", "AES_256"]
    ike_digest_algorithms     = ["SHA2_256"]
    ike_dh_groups             = ["GROUP14"]
    ike_sa_lifetime           = 86400
    
	tunnel_pfs_enabled = true
	tunnel_df_policy = "COPY"
    tunnel_encryption_algorithms = ["AES_256"]
    tunnel_digest_algorithms     = ["SHA2_256"]
    tunnel_dh_groups             = ["GROUP14"]
    tunnel_sa_lifetime           = 3600
    
    dpd_probe_internal = "30"
  }
}
`

const testAccNsxtIpSecVpnTunnelProfileStep5DS = testAccNsxtIpSecVpnTunnelProfileStep5 + testAccNsxtIpSecVpnTunnelDS

func testAccCheckNsxtIpSecVpnTunnelDestroy(ipSecVpnTunnelIdentifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		egw, err := conn.GetNsxtEdgeGateway(testConfig.VCD.Org, testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, testConfig.Nsxt.EdgeGateway)
		}

		_, errByName := egw.GetIpSecVpnTunnelByName(ipSecVpnTunnelIdentifier)
		_, errById := egw.GetIpSecVpnTunnelById(ipSecVpnTunnelIdentifier)

		if errByName == nil {
			return fmt.Errorf("got no errors for NSX-T IPsec VPN Tunnel lookup Name")
		}

		if errById == nil {
			return fmt.Errorf("got no errors for NSX-T IPsec VPN Tunnel lookup by ID")
		}

		return nil
	}
}

func TestAccVcdNsxtIpSecVpnTunnelCertAuth(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate1PrivateKeyPath == "" || testConfig.Certificates.Certificate1Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path, " +
			"Certificates.Certificate1PrivateKeyPath, Certificates.Certificate1Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"EdgeGw":              testConfig.Nsxt.EdgeGateway,
		"NetworkName":         t.Name(),
		"CertificatePath":     testConfig.Certificates.Certificate1Path,
		"PrivateKey":          testConfig.Certificates.Certificate1PrivateKeyPath,
		"PassPhrase":          testConfig.Certificates.Certificate1Pass,
		"RootCertificatePath": testConfig.Certificates.RootCertificatePath,
		"Alias":               t.Name(),
		"AliasCa":             t.Name() + "-ca",
		"Tags":                "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtIpSecVpnTunnelCertAuth1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["ResourceName"] = "test-tunnel-1"
	configText2 := templateFill(testAccNsxtIpSecVpnTunnelCertAuth1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// ignoreDataSourceFields specifies a field list to ignore for data source comparison with resource. These
	// fields should return the same values, but because 'status' is not controlled and resource and data source are
	// read not at the same time - there is a risk one will have status and other won't
	ignoreDataSourceFields := []string{"status", "ike_service_status", "ike_fail_reason"}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1"),
			testAccCheckNsxtIpSecVpnTunnelDestroy("test-tunnel-1-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "cert-san"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "data.vcd_nsxt_ipsec_vpn_tunnel.tunnel1", ignoreDataSourceFields),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "name", "test-tunnel-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "description", "test-tunnel-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "pre_shared_key", "test-psk"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_ip_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "30.30.30.0/28"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "local_networks.*", "40.40.40.1/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_ip_address", "1.2.3.4"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_id", "cert-san"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.#", "3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.1.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "remote_networks.*", "192.168.20.0/28"),
					resource.TestCheckResourceAttr("vcd_nsxt_ipsec_vpn_tunnel.tunnel1", "security_profile_customization.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtIpSecVpnTunnelCertAuth1 = testAccNsxtIpSetPrereqs + `
resource "vcd_library_certificate" "ca-cert" {
  org         = "{{.Org}}"
  alias       = "{{.AliasCa}}"
  description = "CA certificate"
  certificate = file("{{.RootCertificatePath}}")
}

resource "vcd_library_certificate" "cert" {
  org                    = "{{.Org}}"
  alias                  = "{{.Alias}}"
  description            = "Certificate"
  certificate            = file("{{.CertificatePath}}")
  private_key            = file("{{.PrivateKey}}")
  private_key_passphrase = "{{.PassPhrase}}"
}

resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-tunnel-1"
  description = "test-tunnel-description"

  authentication_mode = "CERTIFICATE"
  certificate_id      = vcd_library_certificate.cert.id
  ca_certificate_id   = vcd_library_certificate.ca-cert.id
  
  pre_shared_key    = "test-psk"
  # Primary IP address of Edge Gateway
  local_ip_address  = tolist(data.vcd_nsxt_edgegateway.existing_gw.subnet)[0].primary_ip
  local_networks    = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
  remote_id         = "cert-san"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
}
`
const testAccNsxtIpSecVpnTunnelCertAuth1DS = testAccNsxtIpSecVpnTunnelCertAuth1 + testAccNsxtIpSecVpnTunnelDS
