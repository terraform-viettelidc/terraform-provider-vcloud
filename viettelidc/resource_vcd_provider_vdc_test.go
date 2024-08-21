//go:build ALL || providerVdc || functional

package viettelidc

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"
)

func init() {
	testingTags["providerVdc"] = "resource_vcd_provider_vdc_test.go"
}

func TestAccVcdResourceProviderVdc(t *testing.T) {
	// Note: you need to have at least one free resource pool to test provider VDC creation,
	// and at least two of them to test update. They should be indicated in
	// testConfig.Vsphere.ResourcePoolForVcd1 and testConfig.Vsphere.ResourcePoolForVcd2

	// Pre-checks
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	providerVdcName1 := t.Name()
	newNetworkPoolName := t.Name()
	orgVdcName := "TestOrgVdcNewPvdc"
	orgName := testConfig.VCD.Org
	providerVdcDescription1 := t.Name() + "description"
	providerVdcName2 := t.Name() + "-2"
	providerVdcDescription2 := t.Name() + "description 2"
	// Test configuration
	var params = StringMap{
		"OrgName":                 orgName,
		"OrgVdcName":              orgVdcName,
		"ProviderVdcName1":        providerVdcName1,
		"ProviderVdcDescription1": providerVdcDescription1,
		"ProviderVdcName2":        providerVdcName2,
		"ProviderVdcDescription2": providerVdcDescription2,
		"ResourcePool1":           testConfig.VSphere.ResourcePoolForVcd1,
		"ResourcePool2":           testConfig.VSphere.ResourcePoolForVcd2,
		"NsxtManager":             testConfig.Nsxt.Manager,
		"NewNsxtNetworkPool":      newNetworkPoolName,
		"StorageProfile1":         testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"StorageProfile2":         testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Vcenter":                 testConfig.Networking.Vcenter,
		"FuncName":                t.Name() + "_step1",
	}
	testParamsNotEmpty(t, params)
	params["FuncName"] = t.Name() + "_no-network-pool"
	params["SkipMessage"] = "# "
	configTextNoNetworkPool := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcNoNetworkPool, params)
	debugPrintf("#[DEBUG] Configuration no network pool: %s", configTextNoNetworkPool)

	params["FuncName"] = t.Name()
	params["SkipMessage"] = "# skip-binary-test: redundant"
	configText := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcManagerPrerequisites+
		testAccVcdResourceProviderVdcStep1, params)
	debugPrintf("#[DEBUG] Configuration: %s", configText)

	params["FuncName"] = t.Name() + "_step2"
	configTextRename := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcManagerPrerequisites+
		testAccVcdResourceProviderVdcStep2, params)
	debugPrintf("#[DEBUG] Rename 1: %s", configTextRename)

	params["FuncName"] = t.Name() + "_step3"
	configTextUpdate := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcManagerPrerequisites+
		testAccVcdResourceProviderVdcStep3, params)
	debugPrintf("#[DEBUG] Update 1: %s", configTextUpdate)

	params["FuncName"] = t.Name() + "_step4"
	configTextDisable := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcManagerPrerequisites+
		testAccVcdResourceProviderVdcStep4, params)
	debugPrintf("#[DEBUG] disable: %s", configTextDisable)

	params["SkipMessage"] = ""
	params["FuncName"] = t.Name() + "_step5"
	configTextOrgVdc := templateFill(testAccVcdResourceProviderVdcVcenterPrerequisites+
		testAccVcdResourceProviderVdcManagerPrerequisites+
		testAccVcdResourceProviderVdcStep1+
		testAccVcdResourceProviderVdcStep5, params)
	debugPrintf("#[DEBUG] Add VDC: %s", configTextOrgVdc)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceDef := "vcd_provider_vdc.pvdc1"
	count := 0
	makeFunc := func(label string) func() {
		return func() {
			if vcdTestVerbose {
				fmt.Printf("step %2d - %s\n", count, label)
			}
			count++
		}
	}
	// Test cases
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			checkProviderVdcExists(providerVdcName1, false),
			checkProviderVdcExists(providerVdcName1+"-no-network-pool", false),
			checkNetworkPoolExists(newNetworkPoolName, false),
			checkOrgVdcExists(orgName, orgVdcName, false),
		),
		Steps: []resource.TestStep{
			// step 0 - Create provider VDC without network pool
			{
				Config:    configTextNoNetworkPool,
				PreConfig: makeFunc("create (no-network pool)"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1+"-no-network-pool", true),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "name", providerVdcName1+"-no-network-pool"),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "description", providerVdcDescription1+" no network pool"),
					resource.TestMatchResourceAttr(resourceDef+"nnp", "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "status", "1"),
					resource.TestMatchResourceAttr(resourceDef+"nnp", "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "network_pool_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceDef+"nnp", "storage_profile_names.#", "1"),
				),
			},
			// step 1 - Create provider VDC with network pool
			{
				Config:    configText,
				PreConfig: makeFunc("create"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					checkNetworkPoolExists(newNetworkPoolName, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 2 - Rename the provider VDC
			{
				Config:    configTextRename,
				PreConfig: makeFunc("rename"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName2, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName2),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription2),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 3 - Rename back to original name and description
			{
				Config:    configText,
				PreConfig: makeFunc("rename back"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 4 - Add resource pool and storage profile
			{
				Config:    configTextUpdate,
				PreConfig: makeFunc("add resource pool and storage profile"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "2"),
				),
			},
			// step 5 - remove resource pool and storage profile
			{
				Config:    configText,
				PreConfig: makeFunc("remove resource pool and storage profile"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 6 -Disable provider VDC
			{
				Config:    configTextDisable,
				PreConfig: makeFunc("disable provider VDC"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 7 - Enable provider VDC
			{
				Config:    configText,
				PreConfig: makeFunc("enable provider VDC"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
				),
			},
			// step 8 - Add Org VDC
			{
				Config:    configTextOrgVdc,
				PreConfig: makeFunc("create depending org VDC"),
				Check: resource.ComposeTestCheckFunc(
					checkProviderVdcExists(providerVdcName1, true),
					checkOrgVdcExists(orgName, orgVdcName, true),
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName1),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription1),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
					resource.TestCheckResourceAttr(resourceDef, "resource_pool_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceDef, "storage_profile_names.#", "1"),
					resource.TestCheckResourceAttr("vcd_org_vdc.testVdc", "name", orgVdcName),
				),
			},
			// step 9 - import
			{
				PreConfig:         makeFunc("import provider VDC"),
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(providerVdcName1),
				// These fields can't be retrieved reliably from user data
				ImportStateVerifyIgnore: []string{"network_pool_ids"},
			},
		},
	})
	postTestChecks(t)
}

func checkOrgVdcExists(orgName, vdcName string, wantExisting bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_org_vdc" {
				continue
			}
			org, err := conn.GetOrg(orgName)
			if err != nil {
				return err
			}
			_, err = org.GetVDCByName(vdcName, false)

			if wantExisting {
				if err != nil {
					return fmt.Errorf("org vdc %s not found: %s ", vdcName, err)
				}
			} else {
				if err == nil {
					return fmt.Errorf("org vdc %s not deleted yet", vdcName)
				} else {
					return nil
				}
			}
		}
		return nil
	}
}

func checkProviderVdcExists(providerVdcName string, wantExisting bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_provider_vdc" {
				continue
			}
			_, err := conn.GetProviderVdcExtendedByName(providerVdcName)
			if wantExisting {
				if err != nil {
					return fmt.Errorf("provider vdc %s not found: %s ", providerVdcName, err)
				}
			} else {
				if err == nil {
					return fmt.Errorf("provider vdc %s not deleted yet", providerVdcName)
				} else {
					return nil
				}
			}
		}
		return nil
	}
}

const testAccVcdResourceProviderVdcVcenterPrerequisites = `
data "vcd_vcenter" "vcenter1" {
  name = "{{.Vcenter}}"
}

data "vcd_resource_pool" "rp1" {
  name       = "{{.ResourcePool1}}"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}

data "vcd_resource_pool" "rp2" {
  name       = "{{.ResourcePool2}}"
  vcenter_id = data.vcd_vcenter.vcenter1.id
}
`
const testAccVcdResourceProviderVdcManagerPrerequisites = `
data "vcd_nsxt_manager" "mgr1" {
  name = "{{.NsxtManager}}"
}

resource "vcd_network_pool" "np1" {
  name                = "{{.NewNsxtNetworkPool}}"
  network_provider_id = data.vcd_nsxt_manager.mgr1.id
  type                = "GENEVE" # provider VDC needs either a GENEVE (NSX-T) or a VXLAN (NSX-V) network pool

  backing_selection_constraint = "use-first-available"
}
`
const testAccVcdResourceProviderVdcNoNetworkPool = `
{{.SkipMessage}}
resource "vcd_provider_vdc" "pvdc1nnp" {
  name                               = "{{.ProviderVdcName1}}-no-network-pool"
  description                        = "{{.ProviderVdcDescription1}} no network pool"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["{{.StorageProfile1}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`

const testAccVcdResourceProviderVdcStep1 = `
{{.SkipMessage}}
resource "vcd_provider_vdc" "pvdc1" {
  name                               = "{{.ProviderVdcName1}}"
  description                        = "{{.ProviderVdcDescription1}}"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["{{.StorageProfile1}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`

const testAccVcdResourceProviderVdcStep2 = `
# skip-binary-test: used for updates
resource "vcd_provider_vdc" "pvdc1" {
  name                               = "{{.ProviderVdcName2}}"
  description                        = "{{.ProviderVdcDescription2}}"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["{{.StorageProfile1}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`
const testAccVcdResourceProviderVdcStep3 = `
# skip-binary-test: used for updates
resource "vcd_provider_vdc" "pvdc1" {
  name                               = "{{.ProviderVdcName1}}"
  description                        = "{{.ProviderVdcDescription1}}"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id, data.vcd_resource_pool.rp2.id]
  storage_profile_names              = ["{{.StorageProfile1}}","{{.StorageProfile2}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`

const testAccVcdResourceProviderVdcStep4 = `
# skip-binary-test: used for updates
resource "vcd_provider_vdc" "pvdc1" {
  name                               = "{{.ProviderVdcName1}}"
  description                        = "{{.ProviderVdcDescription1}}"
  is_enabled                         = false
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["{{.StorageProfile1}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`
const testAccVcdResourceProviderVdcStep5 = `
resource "vcd_org_vdc" "testVdc" {
  org               = "{{.OrgName}}"
  name              = "{{.OrgVdcName}}"
  allocation_model  = "ReservationPool"
  network_pool_name = "NSX-T Overlay 1"
  provider_vdc_name = vcd_provider_vdc.pvdc1.name
  compute_capacity {
    cpu {
      allocated = 2048
    }
    memory {
      allocated = 2048
    }
  }
  storage_profile {
    name    = "{{.StorageProfile1}}"
    limit   = 10240
    default = true
  }
  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}
`

// TestAccVcdProviderVdcMetadata tests metadata CRUD on Provider VDCs
func TestAccVcdProviderVdcMetadata(t *testing.T) {
	skipIfNotSysAdmin(t)
	testMetadataEntryCRUD(t,
		testAccCheckVcdProviderVdcMetadata, "vcd_provider_vdc.test-provider-vdc",
		testAccCheckVcdProviderVdcMetadataDatasource, "data.vcd_provider_vdc.test-provider-vdc-ds",
		StringMap{
			"Vcenter":         testConfig.Networking.Vcenter,
			"ResourcePool":    testConfig.VSphere.ResourcePoolForVcd1,
			"NsxtManager":     testConfig.Nsxt.Manager,
			"NsxtNetworkPool": testConfig.VCD.NsxtProviderVdc.NetworkPool,
			"StorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		}, false)
}

const testAccCheckVcdProviderVdcMetadata = `
data "vcd_vcenter" "vcenter" {
  name = "{{.Vcenter}}"
}

data "vcd_resource_pool" "rp" {
  name       = "{{.ResourcePool}}"
  vcenter_id = data.vcd_vcenter.vcenter.id
}

data "vcd_nsxt_manager" "mgr" {
  name = "{{.NsxtManager}}"
}

data "vcd_network_pool" "np" {
  name = "{{.NsxtNetworkPool}}"
}

resource "vcd_provider_vdc" "test-provider-vdc" {
  name                               = "{{.Name}}"
  description                        = "{{.Name}}"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr.id
  network_pool_ids                   = [data.vcd_network_pool.np.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp.id]
  storage_profile_names              = ["{{.StorageProfile}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp.hardware_version
  {{.Metadata}}
}
`

const testAccCheckVcdProviderVdcMetadataDatasource = `
data "vcd_provider_vdc" "test-provider-vdc-ds" {
  name = vcd_provider_vdc.test-provider-vdc.name
}
`

func TestAccVcdProviderVdcMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		providerVdc, err := vcdClient.GetProviderVdcById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Provider VDC '%s': %s", id, err)
		}

		return providerVdc, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdProviderVdcMetadata, "vcd_provider_vdc.test-provider-vdc",
		testAccCheckVcdProviderVdcMetadataDatasource, "data.vcd_provider_vdc.test-provider-vdc-ds",
		getObjectById, StringMap{
			"Vcenter":         testConfig.Networking.Vcenter,
			"ResourcePool":    testConfig.VSphere.ResourcePoolForVcd1,
			"NsxtManager":     testConfig.Nsxt.Manager,
			"NsxtNetworkPool": testConfig.VCD.NsxtProviderVdc.NetworkPool,
			"StorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}
