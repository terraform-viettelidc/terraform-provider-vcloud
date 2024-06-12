//go:build nsxt || alb || ALL || functional

package viettelidc

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbSettings(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":   "DEDICATED",
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	// Set supported_feature_set for ALB Settings
	isApiLessThanVersion37 := changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSetSettings", params, false)
	// Set supported_feature_set for ALB Service Engine Group
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbGeneralSettings, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	params["IsActive"] = "false"
	params["SupportedFeatureSetSettings"] = " "
	configText2 := templateFill(testAccVcdNsxtAlbGeneralSettings, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSetSettings", params, true)
	configText3 := templateFill(testAccVcdNsxtAlbGeneralSettingsCustomService, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", "192.168.255.1/25"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					checkSupportedFeatureSet("vcd_nsxt_alb_settings.test", false, isApiLessThanVersion37),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_settings.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "false"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_settings.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(params["EdgeGw"].(string)),
				ImportStateVerifyIgnore: []string{"vdc", "supported_feature_set"}, // Ignore supported_feature_set as versions <37.0 don't have it
			},
			// This step will "recreate" the resource because service_network_specification requires a rebuild
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_settings.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", "82.10.10.1/25"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					checkSupportedFeatureSet("vcd_nsxt_alb_settings.test", true, isApiLessThanVersion37),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_settings.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(params["EdgeGw"].(string)),
				ImportStateVerifyIgnore: []string{"vdc", "supported_feature_set"}, // Ignore supported_feature_set as versions <37.0 don't have it
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbProviderPrereqs = `
# Local variable is used to avoid direct reference and cover Terraform core bug https://github.com/hashicorp/terraform/issues/29484
# Even changing NSX-T ALB Controller name in UI, plan will cause to recreate all resources depending 
# on vcd_nsxt_alb_importable_cloud data source if this indirect reference (via local) variable is not used.
locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  {{.LicenseType}}
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                                 = "first-se"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.first.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "{{.ReservationModel}}"
  {{.SupportedFeatureSet}}
}
`

const testAccVcdNsxtAlbGeneralSettings = testAccVcdNsxtAlbProviderPrereqs + `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = {{.IsActive}}
  {{.SupportedFeatureSetSettings}}

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}
`

const testAccVcdNsxtAlbGeneralSettingsCustomService = testAccVcdNsxtAlbProviderPrereqs + `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id               = data.vcd_nsxt_edgegateway.existing.id
  is_active                     = true
  service_network_specification = "82.10.10.1/25"
  {{.SupportedFeatureSetSettings}}
  
  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}
`

func TestAccVcdNsxtAlbSettingsTransparentMode(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipNoNsxtAlbConfiguration(t)
	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"ControllerName":         t.Name(),
		"ControllerUrl":          testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername":     testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword":     testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":        testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":       "DEDICATED",
		"Org":                    testConfig.VCD.Org,
		"NsxtVdc":                testConfig.Nsxt.Vdc,
		"EdgeGw":                 testConfig.Nsxt.EdgeGateway,
		"TransparentModeEnabled": true,
		"LicenseType":            " ",
		"SupportedFeatureSet":    `supported_feature_set = "PREMIUM"`,

		"Tags": "nsxt alb",
	}

	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbGeneralSettingsTransparentMode, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	params["TransparentModeEnabled"] = "false"
	params["SupportedFeatureSetSettings"] = " "
	configText2 := templateFill(testAccVcdNsxtAlbGeneralSettingsTransparentMode, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", "192.168.255.1/25"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "ipv6_service_network_specification", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_transparent_mode_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "supported_feature_set", "PREMIUM"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_settings.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", "192.168.255.1/25"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "ipv6_service_network_specification", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_transparent_mode_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "supported_feature_set", "PREMIUM"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbGeneralSettingsTransparentMode = testAccVcdNsxtAlbProviderPrereqs + `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id       = data.vcd_nsxt_edgegateway.existing.id
  is_active             = {{.IsActive}}

  {{.SupportedFeatureSet}}

  is_transparent_mode_enabled = {{.TransparentModeEnabled}}

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}
`

// TestAccVcdNsxtAlbSettingsDualStackMode tests ALB settings with dual stack mode - IPv4 and IPv6
// addresses set for Service Network Specification
func TestAccVcdNsxtAlbSettingsDualStackMode(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipNoNsxtAlbConfiguration(t)
	if checkVersion(testConfig.Provider.ApiVersion, "< 37.0") {
		t.Skipf("This test tests VCD 10.4.0+ (API V37.0+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"ControllerName":         t.Name(),
		"ControllerUrl":          testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername":     testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword":     testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":        testConfig.Nsxt.NsxtAlbImportableCloud,
		"ReservationModel":       "DEDICATED",
		"Org":                    testConfig.VCD.Org,
		"NsxtVdc":                testConfig.Nsxt.Vdc,
		"EdgeGw":                 testConfig.Nsxt.EdgeGateway,
		"TransparentModeEnabled": true,
		"LicenseType":            " ",
		"SupportedFeatureSet":    `supported_feature_set = "PREMIUM"`,

		"Tags": "nsxt alb",
	}

	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbGeneralSettingsDualStackMode, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	params["IsActive"] = "true"
	configText2 := templateFill(testAccVcdNsxtAlbGeneralSettingsDualStackModeIPv6Only, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1, // Setup prerequisites - configure NSX-T ALB in Provider
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", "10.10.255.225/27"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "ipv6_service_network_specification", "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "supported_feature_set", "PREMIUM"),
				),
			},
			{
				// Recreating from scratch while having only ipv6_service_network_specification set
				// and checking that only IPv6 address is set
				Taint:  []string{"vcd_nsxt_alb_settings.test"},
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_nsxt_alb_importable_cloud.cld", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "service_network_specification", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "ipv6_service_network_specification", "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "is_active", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_settings.test", "supported_feature_set", "PREMIUM"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbGeneralSettingsDualStackMode = testAccVcdNsxtAlbProviderPrereqs + `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = {{.IsActive}}
  
  service_network_specification      = "10.10.255.225/27"
  ipv6_service_network_specification = "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"

  {{.SupportedFeatureSet}}


  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}
`

const testAccVcdNsxtAlbGeneralSettingsDualStackModeIPv6Only = testAccVcdNsxtAlbProviderPrereqs + `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}

resource "vcd_nsxt_alb_settings" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  is_active       = {{.IsActive}}
  
  ipv6_service_network_specification = "2001:0db8:85a3:0000:0000:8a2e:0370:7334/120"

  {{.SupportedFeatureSet}}


  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.first]
}
`

func testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(edgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, rs := range s.RootModule().Resources {
			edgeGatewayName := rs.Primary.Attributes["name"]
			if rs.Type != "vcd_edgegateway" {
				continue
			}
			if edgeGatewayName != edgeName {
				continue
			}
			conn := testAccProvider.Meta().(*VCDClient)
			orgName := rs.Primary.Attributes["org"]
			vdcName := rs.Primary.Attributes["vdc"]

			org, _, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
			}

			egw, err := org.GetNsxtEdgeGatewayByName(edgeName)
			if err != nil {
				return fmt.Errorf("error looking up NSX-T edge gateway %s", edgeName)
			}

			albConfig, err := egw.GetAlbSettings()
			if err != nil {
				return fmt.Errorf("error retrieving NSX-T ALB General Settings: %s", err)
			}

			// Destroy operation of resource should disable Load Balancer
			if albConfig.Enabled {
				return fmt.Errorf("expected NSX-T ALB to be disabled in General Settings")
			}
		}

		return nil
	}
}
