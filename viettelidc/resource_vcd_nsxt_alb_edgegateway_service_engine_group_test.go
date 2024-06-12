//go:build nsxt || alb || ALL || functional

package viettelidc

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeGatewayServiceEngineGroupDedicated(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ReservationModel":   "DEDICATED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	// Set supported_feature_set for ALB Service Engine Group
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	// Set supported_feature_set for ALB Settings
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSetSettings", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicatedDS, params)
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
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", regexp.MustCompile(`\d*`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(params["EdgeGw"].(string), "first-se"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicatedDS = testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated + `
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

func TestAccVcdNsxtEdgeGatewayServiceEngineGroupShared(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ReservationModel":   "SHARED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	// Set supported_feature_set for ALB Service Engine Group
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	// Set supported_feature_set for ALB Settings
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSetSettings", params, false)

	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupSharedDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

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
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "100"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "30"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "70"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "35"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", "0"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "70"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", "0"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdNsxtEdgeGatewayObject(params["EdgeGw"].(string), "first-se"),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 100
  reserved_virtual_services = 30
}
`

const testAccVcdNsxtAlbEdgeServiceEngineGroupSharedDS = testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared + `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep3 = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 70
  reserved_virtual_services = 35
}
`

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep4 = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 70
  reserved_virtual_services = 0
}
`

// TestAccVcdNsxtEdgeGatewayServiceEngineGroupResourceNotFound checks that deletion of ALB Service
// Engine Group assignment is correctly handled when resource disappears (remove ID by using
// d.SetId("") instead of throwing error) outside of Terraform control.
func TestAccVcdNsxtEdgeGatewayServiceEngineGroupResourceNotFound(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This test invokes go-vcloud-director SDK directly
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ReservationModel":   "DEDICATED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	// Set supported_feature_set for ALB Service Engine Group
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSet", params, false)
	// Set supported_feature_set for ALB Settings
	changeSupportedFeatureSetIfVersionIsLessThan37("LicenseType", "SupportedFeatureSetSettings", params, false)
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	cachedId := &testCachedFieldValue{}

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
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					cachedId.cacheTestResourceFieldValue("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id"),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()

					edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(cachedId.fieldValue)
					if err != nil {
						t.Errorf("error finding ALB Service Engine Group Assignment: %s", err)
					}

					err = edgeAlbServiceEngineGroupAssignment.Delete()
					if err != nil {
						t.Errorf("error deleting ALB Service Engine Group assignment to Edge Gateway: %s", err)
					}
				},
				// Expecting to get a non-empty plan because resource was removed using SDK in
				// PreConfig
				Config:             configText1,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}
