//go:build vdc || nsxt || ALL || functional

package vcloud

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestAccVcdVdc = "TestAccVcdVdcBasic"

func runOrgVdcTest(t *testing.T, params StringMap, allocationModel string) {

	skipIfNotSysAdmin(t)
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVdc_basic, params)
	params["SecondStorageProfile"] = ""

	// If a second storage profile is defined in the configuration, we add its parameters in the update
	secondStorageProfileMapValue, exist := params["ProviderVdcStorageProfile2"]
	secondStorageProfile := ""
	if exist {
		secondStorageProfile = secondStorageProfileMapValue.(string)
	}
	if secondStorageProfile != "" {
		unfilledTemplate := template.Must(template.New("").Parse(additionalStorageProfile))
		buf := &bytes.Buffer{}
		err := unfilledTemplate.Execute(buf, map[string]interface{}{
			"StorageProfileName":    secondStorageProfile,
			"StorageProfileDefault": false,
		})
		if err == nil {
			fmt.Printf("[INFO] second storage profile will be used in test\n")
			params["SecondStorageProfile"] = buf.String()
		} else {
			fmt.Printf("[WARNING] error reported while filling second storage profile details: %s\n", err)
		}
	} else {
		fmt.Printf("[WARNING] second storage profile will not be used in test\n")
	}

	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdc_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateText)
	secondUpdateText := strings.Replace(updateText, "#START_STORAGE_PROFILE", "/*", 1)
	secondUpdateText = strings.Replace(secondUpdateText, "#END_STORAGE_PROFILE", "*/", 1)
	cachedVdcNumber := &testCachedFieldValue{}

	resourceDef := "vcd_org_vdc." + params["VdcName"].(string)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						resourceDef, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						resourceDef, "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_thin_provisioning", "true"),
					// All VDCs in this test should be NSX-T, and as such the property enable_nsxv_distributed_firewall is false
					resource.TestCheckResourceAttr(
						resourceDef, "enable_nsxv_distributed_firewall", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.name", params["ProviderVdcStorageProfile"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),
					cachedVdcNumber.cacheTestResourceFieldValue("data.vcd_org.org_before", "number_of_vdcs"),
				),
			},
			{
				Config: updateText,
				PreConfig: func() {
					// Increment number of VDCs, as the Organization is now supposed to have one more
					numberOfVdcs, err := strconv.Atoi(cachedVdcNumber.String())
					if err != nil {
						panic("invalid number of VDCs detected")
					}
					numberOfVdcs++
					cachedVdcNumber.fieldValue = fmt.Sprintf("%d", numberOfVdcs)
				},
				Check: resource.ComposeTestCheckFunc(
					testVcdVdcUpdated("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						resourceDef, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						resourceDef, "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "enabled", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_thin_provisioning", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_force", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_recursive", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "memory_guaranteed", params["MemoryGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "cpu_guaranteed", params["CpuGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata2", "VDC Metadata2"),
					testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
						"name":    params["ProviderVdcStorageProfile"].(string),
						"enabled": "true",
						"default": "true",
						"limit":   "20480",
					}),
					// This test runs only if we have a second storage profile
					// It retrieves the details of the second storage profile
					testConditionalCheck(secondStorageProfile != "",
						testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
							"name":    secondStorageProfile,
							"enabled": "false",
							"default": "false",
							"limit":   "20480",
						})),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "elasticity", regexp.MustCompile(`^`+params["ElasticityUpdateValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadUpdateValueForAssert"].(string)+`$`)),
					// This test runs only if we have a second storage profile
					// This check makes sure we have 2 storage profiles
					testConditionalCheck(secondStorageProfile != "",
						resource.TestCheckResourceAttr(resourceDef, "storage_profile.#", "2")),
					cachedVdcNumber.testCheckCachedResourceFieldValue("data.vcd_org.org_after", "number_of_vdcs"),
				),
			},
			// Test removal of second storage profile
			{
				Config: secondUpdateText,
				// This test runs only if we have a second storage profile
				Check: testConditionalCheck(secondStorageProfile != "", resource.ComposeTestCheckFunc(
					// After the removal, we will only have one storage profile
					resource.TestCheckResourceAttr(resourceDef, "storage_profile.#", "1"),
					// This check will find only the first storage profile, since the second one will have been deleted
					testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
						"name":    params["ProviderVdcStorageProfile"].(string),
						"enabled": "true",
						"default": "true",
						"limit":   "20480",
					}),
				)),
			},
			{
				ResourceName:      "vcd_org_vdc." + params["VdcName"].(string),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, params["VdcName"].(string)),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
}

// testConditionalCheck runs the wanted check only if the preliminary condition is true
func testConditionalCheck(condition bool, f resource.TestCheckFunc) resource.TestCheckFunc {
	if condition {
		return f
	}
	return func(s *terraform.State) error { return nil }
}

func testAccCheckVcdVdcExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		return nil
	}
}

func testVcdVdcUpdated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		updateVdc, err := adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		if updateVdc.Vdc.IsEnabled != false {
			return fmt.Errorf("VDC update failed - VDC still enabled")
		}
		return nil
	}
}

func testAccCheckVdcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org_vdc" && rs.Primary.Attributes["name"] != TestAccVcdVdc {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("vdc %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCheckVcdVdc_basic = `
data "vcd_org" "org_before" {
  name = "{{.OrgName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}        {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}
}
`

const testAccCheckVcdVdc_update = `
# skip-binary-test: only for updates

data "vcd_org" "org_after" {
  name = "{{.OrgName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
    }

    memory {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 20480
    default = true
  }

  {{.SecondStorageProfile}}

  metadata = {
    vdc_metadata  = "VDC Metadata"
    vdc_metadata2 = "VDC Metadata2"
  }

  cpu_guaranteed             = {{.CpuGuaranteed}}
  memory_guaranteed          = {{.MemoryGuaranteed}}
  enabled                    = false
  enable_thin_provisioning   = false
  enable_fast_provisioning   = false
  delete_force               = false
  delete_recursive           = false
  {{.FlexElasticKey}}        {{.equalsChar}} {{.FlexElasticValueUpdate}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValueUpdate}}
}
`

// additionalStorageProfile is a component that allows the insertion of a second storage profile
// when one was defined in the configuration file.
// The start/end labels will be replaced by comment markers, thus eliminating the
// second storage profile from the script, so that we can test the removal of the storage profile.
const additionalStorageProfile = `
  #START_STORAGE_PROFILE
  storage_profile {
    name    = "{{.StorageProfileName}}"
    enabled = false
    limit   = 20480
    default = {{.StorageProfileDefault}}
  }
  #END_STORAGE_PROFILE
`

// TestAccVcdVdcMetadata tests metadata CRUD on VDCs
func TestAccVcdVdcMetadata(t *testing.T) {
	skipIfNotSysAdmin(t)
	testMetadataEntryCRUD(t,
		testAccCheckVcdVdcMetadata, "vcd_org_vdc.test-vdc",
		testAccCheckVcdVdcMetadataDatasource, "data.vcd_org_vdc.test-vdc-ds",
		StringMap{
			"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
			"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		}, true)
}

const testAccCheckVcdVdcMetadata = `
resource "vcd_org_vdc" "test-vdc" {
  org              = "{{.Org}}"
  name             = "{{.Name}}"
  allocation_model  = "AllocationVApp"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = 0
      limit     = 0
    }

    memory {
      allocated = 0
      limit     = 0
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    limit   = 100
    default = true
  }

  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.Metadata}}
}
`

const testAccCheckVcdVdcMetadataDatasource = `
data "vcd_org_vdc" "test-vdc-ds" {
  org  = vcd_org_vdc.test-vdc.org
  name = vcd_org_vdc.test-vdc.name
}
`

func TestAccVcdVdcMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetAdminVDCById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", id, err)
		}
		return vdc, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVdcMetadata, "vcd_org_vdc.test-vdc",
		testAccCheckVcdVdcMetadataDatasource, "data.vcd_org_vdc.test-vdc-ds",
		getObjectById, StringMap{
			"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
			"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}
