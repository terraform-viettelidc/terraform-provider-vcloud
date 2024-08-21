//go:build vdc || nsxt || ALL || functional

package vcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcNsxtNetworkProfile(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VdcName": testConfig.Nsxt.Vdc,
		"OrgName": testConfig.VCD.Org,

		"EdgeCluster": testConfig.Nsxt.NsxtEdgeCluster,

		"TestName":                   t.Name(),
		"NsxtManager":                testConfig.Nsxt.Manager,
		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,

		"Tags": "vdc",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdOrgVdcNsxtNetworkProfile, params)
	params["FuncName"] = t.Name() + "-step2DS"
	configText2 := templateFill(testAccVcdOrgVdcNsxtNetworkProfileDS, params)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdOrgVdcNsxtNetworkProfileRemove, params)

	debugPrintf("#[DEBUG] CONFIGURATION - Step1: %s", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION - Step2: %s", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION - Step4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_org_vdc_nsxt_network_profile.nsxt", "vdc_networks_default_segment_profile_template_id"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc_nsxt_network_profile.nsxt", "vapp_networks_default_segment_profile_template_id"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc_nsxt_network_profile.nsxt", "edge_cluster_id"),
					resource.TestCheckResourceAttrPair("vcd_org_vdc_nsxt_network_profile.nsxt", "edge_cluster_id", "data.vcd_org_vdc.nsxt2", "edge_cluster_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_org_vdc_nsxt_network_profile.nsxt", "vcd_org_vdc_nsxt_network_profile.nsxt", nil),
				),
			},
			{
				ResourceName:      "vcd_org_vdc_nsxt_network_profile.nsxt",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.Nsxt.Vdc,
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_org_vdc_nsxt_network_profile.nsxt", "vdc_networks_default_segment_profile_template_id", ""),
					resource.TestCheckResourceAttr("vcd_org_vdc_nsxt_network_profile.nsxt", "vapp_networks_default_segment_profile_template_id", ""),
					resource.TestCheckResourceAttr("vcd_org_vdc_nsxt_network_profile.nsxt", "edge_cluster_id", ""),
					resource.TestCheckResourceAttrPair("vcd_org_vdc_nsxt_network_profile.nsxt", "edge_cluster_id", "data.vcd_org_vdc.nsxt2", "edge_cluster_id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdOrgVdcNsxtNetworkProfileCommon = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name       = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name       = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name       = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name       = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name       = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

resource "vcd_nsxt_segment_profile_template" "complete" {
  name        = "{{.TestName}}-complete"
  description = "description"

  nsxt_manager_id             = data.vcd_nsxt_manager.nsxt.id
  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}

`

const testAccVcdOrgVdcNsxtNetworkProfile = testAccVcdOrgVdcNsxtNetworkProfileCommon + `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"
}

data "vcd_nsxt_edge_cluster" "first" {
  org    = "{{.OrgName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
  name   = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc_nsxt_network_profile" "nsxt" {
  org = "{{.OrgName}}"
  vdc = "{{.VdcName}}"

  edge_cluster_id                                   = data.vcd_nsxt_edge_cluster.first.id
  vdc_networks_default_segment_profile_template_id  = vcd_nsxt_segment_profile_template.complete.id
  vapp_networks_default_segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}

data "vcd_org_vdc" "nsxt2" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"

  depends_on = [vcd_org_vdc_nsxt_network_profile.nsxt]
}
`

const testAccVcdOrgVdcNsxtNetworkProfileDS = testAccVcdOrgVdcNsxtNetworkProfile + `
data "vcd_org_vdc_nsxt_network_profile" "nsxt" {
  org = vcd_org_vdc_nsxt_network_profile.nsxt.org
  vdc = vcd_org_vdc_nsxt_network_profile.nsxt.vdc
}
`

const testAccVcdOrgVdcNsxtNetworkProfileRemove = testAccVcdOrgVdcNsxtNetworkProfileCommon + `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"
}

data "vcd_nsxt_edge_cluster" "first" {
  org    = "{{.OrgName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
  name   = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc_nsxt_network_profile" "nsxt" {
  org = "{{.OrgName}}"
  vdc = "{{.VdcName}}"
}

data "vcd_org_vdc" "nsxt2" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"

  depends_on = [vcd_org_vdc_nsxt_network_profile.nsxt]
}
`
