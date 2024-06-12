//go:build network || nsxt || ALL || functional

package viettelidc

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNatRuleDnat(t *testing.T) {
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

	configText1 := templateFill(testAccNsxtNatDnat, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["RuleName"] = "test-dnat-rule"
	configText2 := templateFill(testAccNsxtNatDnatDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccNsxtNatDnatStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	params["RuleName"] = "test-dnat-rule-updated"
	configText4 := templateFill(testAccNsxtNatDnatStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccNsxtNatDnatStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	natRuleId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtNatRuleDestroy("test-dnat-rule"),
			testAccCheckNsxtNatRuleDestroy("test-dnat-rule-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_nat_rule.dnat", "edge_gateway_id", regexp.MustCompile(`^urn:vcloud:gateway:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "name", "test-dnat-rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "rule_type", "DNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "description", "description"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "external_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "internal_address", "11.11.11.2"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "logging", "false"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_nat_rule.dnat", "app_port_profile_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "snat_destination_address", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "enabled", "true"),
					natRuleId.cacheTestResourceFieldValue("vcd_nsxt_nat_rule.dnat", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "id"),
					resourceFieldsEqual("vcd_nsxt_nat_rule.dnat", "data.vcd_nsxt_nat_rule.nat", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_nat_rule.dnat", "edge_gateway_id", regexp.MustCompile(`^urn:vcloud:gateway:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "name", "test-dnat-rule-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "rule_type", "DNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "description", "updated-description"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "external_address"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "app_port_profile_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "internal_address", "11.11.11.0/32"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "dnat_external_port", "8888"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "snat_destination_address", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat", "enabled", "false"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "id"),
					resourceFieldsEqual("vcd_nsxt_nat_rule.dnat", "data.vcd_nsxt_nat_rule.nat", nil),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat", "id"),
					// Ensure custom created Application Port Profile ID is consumed
					resource.TestCheckResourceAttrPair("vcd_nsxt_nat_rule.dnat", "app_port_profile_id", "vcd_nsxt_app_port_profile.custom-app", "id"),
				),
			},
			// Try to import by Name
			{
				ResourceName:            "vcd_nsxt_nat_rule.dnat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-dnat-rule-updated"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			// Try to import by rule UUID
			{
				ResourceName: "vcd_nsxt_nat_rule.dnat",
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

const natRuleDataSourceDefinition = `
# skip-binary-test: Data Source test
data "vcd_nsxt_nat_rule" "nat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  name            = "{{.RuleName}}"
}
`

const testAccNsxtNatDnat = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "dnat" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-dnat-rule"
  rule_type   = "DNAT"
  description = "description"

  # Using primary_ip from edge gateway
  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address = "11.11.11.2"
}
`

const testAccNsxtNatDnatDS = testAccNsxtNatDnat + natRuleDataSourceDefinition

const testAccNsxtNatDnatStep2 = testAccNsxtSecurityGroupPrereqsEmpty + `
data "vcd_nsxt_app_port_profile" "custom" {
  scope = "SYSTEM"
  name  = "SSH"
}

resource "vcd_nsxt_nat_rule" "dnat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-dnat-rule-updated"
  rule_type  = "DNAT"
  description = "updated-description"
  
  # Using primary_ip from edge gateway
  external_address  = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address  = "11.11.11.0/32"
  dnat_external_port  = 8888
  app_port_profile_id = data.vcd_nsxt_app_port_profile.custom.id
  
  enabled = false
}
`

const testAccNsxtNatDnatStep2DS = testAccNsxtNatDnatStep2 + natRuleDataSourceDefinition

const testAccNsxtNatDnatStep5 = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_app_port_profile" "custom-app" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name        = "custom app profile"
  description = "Application port profile for custom application"
  scope       = "TENANT"
  # NAT rules accept only Application Port Profiles with 1 Port
  app_port {
    protocol = "TCP"
    port     = ["2000"]
  }
}


resource "vcd_nsxt_nat_rule" "dnat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-dnat-rule-updated"
  rule_type  = "DNAT"
  description = "updated-description"
  
  # Using primary_ip from edge gateway
  external_address  = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address  = "11.11.11.0/32"
  dnat_external_port  = 8888
  app_port_profile_id = vcd_nsxt_app_port_profile.custom-app.id
  
  logging = false
  enabled = false
}
`

func TestAccVcdNsxtNatRuleNoDnat(t *testing.T) {
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

	configText := templateFill(testAccNsxtNatNoDnat, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtNatRuleDestroy("test-no-dnat-rule"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.no-dnat", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "name", "test-no-dnat-rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "rule_type", "NO_DNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.no-dnat", "external_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "internal_address", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "logging", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-dnat", "dnat_external_port", "7777"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.no-dnat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-no-dnat-rule"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtNatNoDnat = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "no-dnat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name      = "test-no-dnat-rule"
  rule_type = "NO_DNAT"

  
  # Using primary_ip from edge gateway
  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  dnat_external_port = 7777
}
`

func TestAccVcdNsxtNatRuleSnat(t *testing.T) {
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

	configText1 := templateFill(testAccNsxtNatSnat, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["RuleName"] = "test-snat-rule"
	configText2 := templateFill(testAccNsxtNatSnatDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccNsxtNatSnat2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	params["RuleName"] = "test-snat-rule-updated"
	configText4 := templateFill(testAccNsxtNatSnat2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtNatRuleDestroy("test-snat-rule"),
			testAccCheckNsxtNatRuleDestroy("test-snat-rule-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_nat_rule.snat", "edge_gateway_id", regexp.MustCompile(`^urn:vcloud:gateway:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "name", "test-snat-rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "rule_type", "SNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "description", "description"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "external_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "internal_address", "11.11.11.2"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_nat_rule.snat", "app_port_profile_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "snat_destination_address", "8.8.8.8"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "logging", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "id"),
					resourceFieldsEqual("vcd_nsxt_nat_rule.snat", "data.vcd_nsxt_nat_rule.nat", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_nat_rule.snat", "edge_gateway_id", regexp.MustCompile(`^urn:vcloud:gateway:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "name", "test-snat-rule-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "rule_type", "SNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "external_address"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "internal_address", "10.10.10.0/24"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_nat_rule.snat", "app_port_profile_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "snat_destination_address", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.snat", "logging", "false"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.snat", "id"),
					resourceFieldsEqual("vcd_nsxt_nat_rule.snat", "data.vcd_nsxt_nat_rule.nat", nil),
				),
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.snat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-snat-rule-updated"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtNatSnat = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "snat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
	
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-snat-rule"
  rule_type   = "SNAT"
  description = "description"
  
  # Using primary_ip from edge gateway
  external_address         = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address         = "11.11.11.2"
  snat_destination_address = "8.8.8.8"
}
`
const testAccNsxtNatSnatDS = testAccNsxtNatSnat + natRuleDataSourceDefinition

const testAccNsxtNatSnat2 = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "snat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
	
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-snat-rule-updated"
  rule_type   = "SNAT"
  description = ""
  
  # Using primary_ip from edge gateway
  external_address         = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address         = "10.10.10.0/24"
  logging = false
}
`

const testAccNsxtNatSnat2DS = testAccNsxtNatSnat2 + natRuleDataSourceDefinition

func TestAccVcdNsxtNatRuleNoSnat(t *testing.T) {
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

	configText := templateFill(testAccNsxtNatNoSnat, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtNatRuleDestroy("test-no-snat-rule"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.no-snat", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-snat", "name", "test-no-snat-rule"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-snat", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-snat", "rule_type", "NO_SNAT"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.no-snat", "internal_address", "11.11.11.0/24"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.no-snat",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-no-snat-rule"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtNatNoSnat = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "no-snat" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-no-snat-rule"
  rule_type   = "NO_SNAT"
  description = "description"
  
  # Using primary_ip from edge gateway
  internal_address         = "11.11.11.0/24"
}
`

// TestAccVcdNsxtNatRuleFirewallMatchPriority explicitly tests support for two new fields introduced
// in API 35.2 (VCD 10.2.2) firewall_match and priority.
func TestAccVcdNsxtNatRuleFirewallMatchPriority(t *testing.T) {
	preTestChecks(t)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"NsxtVdc":       testConfig.Nsxt.Vdc,
		"EdgeGw":        testConfig.Nsxt.EdgeGateway,
		"NetworkName":   t.Name(),
		"Tags":          "network nsxt",
		"FirewallMatch": "MATCH_INTERNAL_ADDRESS",
		"Priority":      "10",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtNatFirewallMatchPriority, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["FirewallMatch"] = "MATCH_EXTERNAL_ADDRESS"
	params["Priority"] = "30"
	configText2 := templateFill(testAccNsxtNatFirewallMatchPriority, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	params["FirewallMatch"] = "BYPASS"
	params["Priority"] = "0"
	configText3 := templateFill(testAccNsxtNatFirewallMatchPriority, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtNatRuleDestroy("test-dnat-rule-match-and-priority"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat-match", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "name", "test-dnat-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "firewall_match", "MATCH_INTERNAL_ADDRESS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "priority", "10"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat-match", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "name", "test-dnat-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "firewall_match", "MATCH_EXTERNAL_ADDRESS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "priority", "30"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.dnat-match", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "name", "test-dnat-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "firewall_match", "BYPASS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.dnat-match", "priority", "0"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.dnat-match",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-dnat-rule-match-and-priority"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtNatFirewallMatchPriority = testAccNsxtSecurityGroupPrereqsEmpty + `
resource "vcd_nsxt_nat_rule" "dnat-match" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-dnat-rule-match-and-priority"
  rule_type   = "DNAT"
  description = "description"

  # Using primary_ip from edge gateway
  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address = "11.11.11.2"

  firewall_match = "{{.FirewallMatch}}"
  priority       = "{{.Priority}}"
}
`

func TestAccVcdNsxtNatRuleReflexive(t *testing.T) {
	preTestChecks(t)
	if noTestCredentials() {
		t.Skip("Skipping test run as no credentials are provided and this test needs to lookup VCD version")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"NsxtVdc":       testConfig.Nsxt.Vdc,
		"EdgeGw":        testConfig.Nsxt.EdgeGateway,
		"NetworkName":   t.Name(),
		"Tags":          "network nsxt",
		"FirewallMatch": "MATCH_INTERNAL_ADDRESS",
		"Priority":      "10",
		"SkipNotice":    " ",
	}

	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccNsxtNatRuleReflexive, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["FirewallMatch"] = "MATCH_EXTERNAL_ADDRESS"
	params["Priority"] = "30"
	configText2 := templateFill(testAccNsxtNatRuleReflexive, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	params["FirewallMatch"] = "BYPASS"
	params["Priority"] = "0"
	configText3 := templateFill(testAccNsxtNatRuleReflexive, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtNatRuleDestroy("test-dnat-rule-match-and-priority"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.reflexive", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "name", "test-reflexive-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "rule_type", "REFLEXIVE"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "firewall_match", "MATCH_INTERNAL_ADDRESS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "priority", "10"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.reflexive", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "name", "test-reflexive-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "rule_type", "REFLEXIVE"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "firewall_match", "MATCH_EXTERNAL_ADDRESS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "priority", "30"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_nat_rule.reflexive", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "name", "test-reflexive-rule-match-and-priority"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "rule_type", "REFLEXIVE"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "firewall_match", "BYPASS"),
					resource.TestCheckResourceAttr("vcd_nsxt_nat_rule.reflexive", "priority", "0"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_nat_rule.reflexive",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(testConfig.Nsxt.EdgeGateway, "test-reflexive-rule-match-and-priority"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtNatRuleReflexive = testAccNsxtSecurityGroupPrereqsEmpty + `
{{.SkipNotice}}
resource "vcd_nsxt_nat_rule" "reflexive" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "test-reflexive-rule-match-and-priority"
  rule_type   = "REFLEXIVE"
  description = "description"

  # Using primary_ip from edge gateway
  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address = "11.11.11.2"

  firewall_match = "{{.FirewallMatch}}"
  priority       = "{{.Priority}}"
}
`

func testAccCheckNsxtNatRuleDestroy(natRuleIdentifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		egw, err := conn.GetNsxtEdgeGateway(testConfig.VCD.Org, testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, testConfig.Nsxt.EdgeGateway)
		}

		_, errByName := egw.GetNatRuleByName(natRuleIdentifier)
		_, errById := egw.GetNatRuleById(natRuleIdentifier)

		if errByName == nil {
			return fmt.Errorf("got no errors for NSX-T NAT rule lookup by Name")
		}

		if errById == nil {
			return fmt.Errorf("got no errors for NSX-T NAT rule lookup by ID")
		}

		return nil
	}
}
