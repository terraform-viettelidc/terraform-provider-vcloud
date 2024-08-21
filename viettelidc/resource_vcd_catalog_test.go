//go:build catalog || ALL || functional

package viettelidc

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	testingTags["catalog"] = "resource_vcd_catalog_test.go"
}

var TestAccVcdCatalogName = "TestAccVcdCatalog"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalog(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":                    testConfig.VCD.Org,
		"CatalogName":            TestAccVcdCatalogName,
		"Description":            TestAccVcdCatalogDescription,
		"StorageProfile":         testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"CatalogItemName":        "TestCatalogItem",
		"CatalogItemNameFromUrl": "Test",
		"DescriptionFromUrl":     "Test",
		"OvaPath":                testConfig.Ova.OvaPath,
		"UploadProgressFromUrl":  testConfig.Ova.UploadProgress,
		"CatalogMediaName":       "TestCatalogMedia",
		"MediaPath":              testConfig.Media.MediaPath,
		"UploadPieceSize":        testConfig.Media.UploadPieceSize,
		"UploadProgress":         testConfig.Media.UploadProgress,
		"Tags":                   "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalog, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step1"
	params["Description"] = "TestAccVcdCatalogBasicDescription-description"
	configText1 := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	// Use field value caching function across multiple test steps to ensure object wasn't recreated (ID did not change)
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision catalog without storage profile and a vApp template and media
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(
						resourceAddress, "metadata.catalog_metadata", "catalog Metadata"),
					resource.TestCheckResourceAttr(
						resourceAddress, "metadata.catalog_metadata2", "catalog Metadata2"),
					resource.TestMatchResourceAttr(resourceAddress, "catalog_version", regexp.MustCompile(`^\d+`)),
					resource.TestMatchResourceAttr(resourceAddress, "owner_name", regexp.MustCompile(`^\S+$`)),
					resource.TestCheckResourceAttr(resourceAddress, "number_of_vapp_templates", "0"),
					resource.TestCheckResourceAttr(resourceAddress, "number_of_media", "0"),
					resource.TestCheckResourceAttr(resourceAddress, "is_shared", "false"),
					resource.TestCheckResourceAttr(resourceAddress, "is_published", "false"),
					resource.TestMatchResourceAttr(resourceAddress, "publish_subscription_type", regexp.MustCompile(`^(PUBLISHED|SUBSCRIBED|UNPUBLISHED)$`)),
				),
			},
			// Set storage profile for existing catalog
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", "TestAccVcdCatalogBasicDescription-description"),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(
						resourceAddress, "metadata.catalog_metadata", "catalog Metadata v2"),
					resource.TestCheckResourceAttr(
						resourceAddress, "metadata.catalog_metadata2", "catalog Metadata2 v2"),
					resource.TestCheckResourceAttr(
						resourceAddress, "metadata.catalog_metadata3", "catalog Metadata3"),
					resource.TestMatchResourceAttr(resourceAddress, "catalog_version", regexp.MustCompile(`^\d+`)),
					resource.TestMatchResourceAttr(resourceAddress, "owner_name", regexp.MustCompile(`^\S+$`)),
					resource.TestCheckResourceAttr(resourceAddress, "number_of_vapp_templates", "1"),
					resource.TestCheckResourceAttr(resourceAddress, "number_of_media", "1"),
				),
			},
			// Remove storage profile just like it was provisioned in step 0
			{

				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			{
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, TestAccVcdCatalogName),
				// These fields can't be retrieved from catalog data
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdCatalogRename ensures that a Catalog can be renamed and the contents of it
// will remain unchanged.
func TestAccVcdCatalogRename(t *testing.T) {
	preTestChecks(t)

	orgName := testConfig.VCD.Org
	vdcName := testConfig.Nsxt.Vdc
	catalogName := t.Name() + "-cat"
	catalogMediaName := t.Name() + "-media"
	vappTemplateName := t.Name() + "-templ"
	vmName := "test-vm"

	var params = StringMap{
		"Org":                orgName,
		"Vdc":                vdcName,
		"CatalogName":        catalogName,
		"NsxtStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"CatalogMediaName":   catalogMediaName,
		"VappTemplateName":   vappTemplateName,
		"Description":        t.Name(),
		"OvaPath":            testConfig.Ova.OvaPath,
		"MediaPath":          testConfig.Media.MediaPath,
		"UploadPieceSize":    testConfig.Media.UploadPieceSize,
		"VmName":             vmName,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogRename, params)

	catalogUpdatedName := catalogName + "_updated"
	params["FuncName"] = t.Name() + "-rename"
	params["CatalogName"] = catalogUpdatedName
	renameText := templateFill(testAccCheckVcdCatalogRename, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] RENAMING CONFIGURATION: %s", renameText)

	resourceCatalog := "vcd_catalog.test-catalog"
	resourceMedia := "vcd_catalog_media.test-media"
	resourcevAppTemplate := "vcd_catalog_vapp_template.test-vapp-template"
	resourceVM1 := "vcd_vm.test-vm-1"
	resourceVM2 := "vcd_vm.test-vm-2"
	// Use field value caching function across multiple test steps to ensure object wasn't recreated (ID did not change)
	cachedCatalogId := &testCachedFieldValue{}
	cachedMediaId := &testCachedFieldValue{}
	cachedvAppTemplateId := &testCachedFieldValue{}
	cachedVMId1 := &testCachedFieldValue{}
	cachedVMId2 := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckCatalogDestroy(orgName, catalogUpdatedName),
			testAccCheckVcdStandaloneVmDestroy(vmName+"-1", orgName, vdcName),
			testAccCheckVcdStandaloneVmDestroy(vmName+"-2", orgName, vdcName),
		),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceCatalog),
					cachedCatalogId.cacheTestResourceFieldValue(resourceCatalog, "id"),
					cachedMediaId.cacheTestResourceFieldValue(resourceMedia, "id"),
					cachedvAppTemplateId.cacheTestResourceFieldValue(resourcevAppTemplate, "id"),
					cachedVMId1.cacheTestResourceFieldValue(resourceVM1, "id"),
					cachedVMId2.cacheTestResourceFieldValue(resourceVM2, "id"),
					testAccCheckCatalogEntityState("vcd_catalog_media", orgName, catalogMediaName, true),
					testAccCheckCatalogEntityState("vcd_catalog_vapp_template", orgName, vappTemplateName, true),
					resource.TestCheckResourceAttr(resourceCatalog, "name", catalogName),
					resource.TestCheckResourceAttr(resourceCatalog, "description", t.Name()),
					resource.TestMatchResourceAttr(resourceCatalog, "catalog_version", regexp.MustCompile(`^\d+`)),
					resource.TestMatchResourceAttr(resourceCatalog, "owner_name", regexp.MustCompile(`^\S+$`)),
				),
			},
			{
				// Intermediate step needed to rename the catalog before checking
				// the vApp template and media depending on it.
				Config: renameText,
				Check: resource.ComposeTestCheckFunc(
					cachedCatalogId.testCheckCachedResourceFieldValue(resourceCatalog, "id"),
					cachedMediaId.testCheckCachedResourceFieldValue(resourceMedia, "id"),
					cachedvAppTemplateId.testCheckCachedResourceFieldValue(resourcevAppTemplate, "id"),
					cachedVMId1.testCheckCachedResourceFieldValue(resourceVM1, "id"),
					cachedVMId2.testCheckCachedResourceFieldValue(resourceVM2, "id"),
					testAccCheckVcdStandaloneVmExists(vmName+"-1", resourceVM1, orgName, vdcName),
					testAccCheckVcdStandaloneVmExists(vmName+"-2", resourceVM2, orgName, vdcName),
				),
			},
			{
				Config: renameText,
				Check: resource.ComposeTestCheckFunc(
					cachedCatalogId.testCheckCachedResourceFieldValue(resourceCatalog, "id"),
					cachedMediaId.testCheckCachedResourceFieldValue(resourceMedia, "id"),
					cachedvAppTemplateId.testCheckCachedResourceFieldValue(resourcevAppTemplate, "id"),
					cachedVMId1.testCheckCachedResourceFieldValue(resourceVM1, "id"),
					cachedVMId2.testCheckCachedResourceFieldValue(resourceVM2, "id"),
					resource.TestCheckResourceAttr(resourceCatalog, "name", catalogUpdatedName),
					resource.TestCheckResourceAttr(resourceCatalog, "number_of_vapp_templates", "1"),
					resource.TestCheckResourceAttr(resourceCatalog, "number_of_media", "1"),
					resource.TestCheckResourceAttrPair(resourcevAppTemplate, "catalog_id", resourceCatalog, "id"),
					resource.TestCheckResourceAttrPair(resourceMedia, "catalog_id", resourceCatalog, "id"),
				),
			},
		},
	})
}

const testAccCheckVcdCatalogRename = `
data "vcd_storage_profile" "sp1" {
  org  = "{{.Org}}" 
  vdc  = "{{.Vdc}}"
  name = "{{.NsxtStorageProfile}}"
}

resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp1.id

  delete_force     = "true"
  delete_recursive = "true"
}

resource "vcd_catalog_vapp_template" "test-vapp-template" {
  org     = "{{.Org}}"
  catalog_id = resource.vcd_catalog.test-catalog.id

  name              = "{{.VappTemplateName}}"
  description       = "TestDescription"
  ova_path          = "{{.OvaPath}}"
  upload_piece_size = {{.UploadPieceSize}}
}

resource "vcd_catalog_media"  "test-media" {
  org     = "{{.Org}}"
  catalog_id = resource.vcd_catalog.test-catalog.id

  name              = "{{.CatalogMediaName}}"
  description       = "TestDescription"
  media_path        = "{{.MediaPath}}"
  upload_piece_size = {{.UploadPieceSize}}
}

resource "vcd_vm" "{{.VmName}}-1" {
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  name             = "{{.VmName}}-1"
  vapp_template_id = resource.vcd_catalog_vapp_template.test-vapp-template.id
  description      = "test standalone VM 1"
  power_on         = false
}

resource "vcd_vm" "{{.VmName}}-2" {
  org              = "{{.Org}}"
  vdc              = "{{.Vdc}}"
  name             = "{{.VmName}}-2"
  boot_image_id    = resource.vcd_catalog_media.test-media.id
  description      = "test standalone VM 2"
  computer_name    = "standalone"
  cpus             = 1
  memory           = 1024
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  power_on         = false
}

`

// TestAccVcdCatalogWithStorageProfile is very similar to TestAccVcdCatalog, but it ensure that a catalog can be created
// using specific storage profile
func TestAccVcdCatalogWithStorageProfile(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.Nsxt.Vdc,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogWithStorageProfile, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	dataSourceAddress := "data.vcd_storage_profile.sp"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision with storage profile
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					resource.TestCheckResourceAttrPair(resourceAddress, "storage_profile_id", dataSourceAddress, "id"),
					checkStorageProfileOriginatesInParentVdc(dataSourceAddress,
						params["StorageProfile"].(string),
						params["Org"].(string),
						params["Vdc"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdCatalogPublished = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabled}}"
  cache_enabled                 = "{{.CacheEnabled}}"
  preserve_identity_information = "{{.PreserveIdentityInformation}}"
  password                      = "superUnknown"
}
`

const testAccCheckVcdCatalogPublishedUpdate1 = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabledUpdate1}}"
  cache_enabled                 = "{{.CacheEnabledUpdate1}}"
  preserve_identity_information = "{{.PreserveIdentityInformationUpdate1}}"
  password                      = "superUnknown"
}
`

const testAccCheckVcdCatalogPublishedUpdate2 = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabledUpdate2}}"
  cache_enabled                 = "{{.CacheEnabledUpdate2}}"
  preserve_identity_information = "{{.PreserveIdentityInformationUpdate2}}"
  password                      = ""
}
`

// TestAccVcdCatalogPublishedToExternalOrg is very similar to TestAccVcdCatalog, but it ensures that a catalog can be
// published to external Org
func TestAccVcdCatalogPublishedToExternalOrg(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":                                testConfig.VCD.Org,
		"Vdc":                                testConfig.Nsxt.Vdc,
		"CatalogName":                        TestAccVcdCatalogName,
		"Description":                        TestAccVcdCatalogDescription,
		"Tags":                               "catalog",
		"PublishEnabled":                     true,
		"PublishEnabledUpdate1":              true,
		"PublishEnabledUpdate2":              false,
		"CacheEnabled":                       true,
		"CacheEnabledUpdate1":                false,
		"CacheEnabledUpdate2":                false,
		"PreserveIdentityInformation":        true,
		"PreserveIdentityInformationUpdate1": false,
		"PreserveIdentityInformationUpdate2": false,
	}
	testParamsNotEmpty(t, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// TODO - This code snippet is to avoid having the org catalog publishing settings set to disable.
	// There are some bugs in VCD that disable those options. This code snippet will be removed
	// as soon as those bugs are solved.
	vcdClient := createSystemTemporaryVCDConnection()
	adminOrg, err := vcdClient.GetAdminOrg(testConfig.VCD.Org)
	if err != nil {
		t.Errorf("couldn't retrieve the adminOrg for setting workaround for VCD bug - %s", err)
	}

	adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishCatalogs = true
	adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanPublishExternally = true
	adminOrg.AdminOrg.OrgSettings.OrgGeneralSettings.CanSubscribe = true
	task, err := adminOrg.Update()
	if err != nil {
		t.Errorf("couldn't update the adminOrg settings for workaround for VCD bug - %s", err)
	}

	err = task.WaitTaskCompletion()
	if err != nil {
		t.Errorf("the task that performs the VCD bug workaround didn't finish successfully - %s", err)
	}
	// End of the workaround

	configText := templateFill(testAccCheckVcdCatalogPublished, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "step1"
	configTextUpd1 := templateFill(testAccCheckVcdCatalogPublishedUpdate1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextUpd1)
	params["FuncName"] = t.Name() + "step2"
	configTextUpd2 := templateFill(testAccCheckVcdCatalogPublishedUpdate2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextUpd2)

	resourceAddress := "vcd_catalog.test-catalog"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabled"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformation"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabled"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
			{
				Config: configTextUpd1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabledUpdate1"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformationUpdate1"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabledUpdate1"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
			{
				Config: configTextUpd2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabledUpdate2"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformationUpdate2"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabledUpdate2"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdCatalogExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Org ID is set")
		}

		if testAccProvider == nil || testAccProvider.Meta() == nil {
			return fmt.Errorf("testAccProvider is not initialised")
		}
		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%s)", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckCatalogDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog" && rs.Primary.Attributes["name"] != TestAccVcdCatalogName {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("catalog %s still exists", rs.Primary.ID)
		}

	}

	return nil
}

const testAccCheckVcdCatalog = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  metadata = {
    catalog_metadata  = "catalog Metadata"
    catalog_metadata2 = "catalog Metadata2"
  }
}

resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = resource.vcd_catalog.test-catalog.name

  name                 = "{{.CatalogItemName}}"
  description          = "TestDescription"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}

resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = resource.vcd_catalog.test-catalog.name

  name                 = "{{.CatalogMediaName}}"
  description          = "TestDescription"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}

`

const testAccCheckVcdCatalogStep1 = `
data "vcd_storage_profile" "sp" {
	name = "{{.StorageProfile}}"
}

resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp.id

  delete_force      = "true"
  delete_recursive  = "true"

  metadata = {
    catalog_metadata  = "catalog Metadata v2"
    catalog_metadata2 = "catalog Metadata2 v2"
    catalog_metadata3 = "catalog Metadata3"
  }
}

resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = resource.vcd_catalog.test-catalog.name

  name                 = "{{.CatalogItemName}}"
  description          = "TestDescription"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}

resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = resource.vcd_catalog.test-catalog.name

  name                 = "{{.CatalogMediaName}}"
  description          = "TestDescription"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"
}
`

const testAccCheckVcdCatalogWithStorageProfile = `
data "vcd_storage_profile" "sp" {
	name = "{{.StorageProfile}}"
}

resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp.id

  delete_force      = "true"
  delete_recursive  = "true"

  metadata = {
    catalog_metadata  = "catalog Metadata v2"
    catalog_metadata2 = "catalog Metadata2 v2"
    catalog_metadata3 = "catalog Metadata3"
  }
}
`

// TestAccVcdCatalogSharedAccess is a test to cover bugfix when Organization Administrator is not able to lookup shared
// catalog from another Org
// Because of limited Terraform acceptance test functionality it uses go-vcloud-director SDK to pre-configure
// environment explicitly using System Org (even if it the test is run as Org user). The following objects are created
// using SDK (their cleanup is deferred):
// * Org
// * Vdc inside newly created Org
// * Catalog inside newly created Org. This catalog is shared with Org defined in testConfig.VCD.Org variable
// * Uploads A minimal vApp template to save on upload / VM spawn time
//
// After these objects are pre-created using SDK, terraform definition is used to spawn a VM by using template in a
// catalog from another Org. This test works in both System and Org admin roles but the bug (which was introduced in SDK
// v2.12.0 and terraform-provider-viettelidc v3.3.0) occurred only for Organization Administrator user.
//
// Original issue -  https://github.com/vmware/terraform-provider-vcd/issues/689
func TestAccVcdCatalogSharedAccess(t *testing.T) {
	preTestChecks(t)
	// This test manipulates VCD during runtime using SDK and is not possible to run as binary or upgrade test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// initiate System client ignoring value of `VCD_TEST_ORG_USER` flag and create pre-requisite objects
	systemClient := createSystemTemporaryVCDConnection()
	catalog, vdc, oldOrg, newOrg, err := spawnTestOrgVdcSharedCatalog(systemClient, t.Name())
	if err != nil {
		testOrgVdcSharedCatalogCleanUp(catalog, vdc, oldOrg, newOrg, t)
		t.Fatalf("%s", err)
	}
	// call cleanup ath the end of the test with any of the entities that have been created up to that point
	defer func() { testOrgVdcSharedCatalogCleanUp(catalog, vdc, oldOrg, newOrg, t) }()

	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.Nsxt.Vdc,
		"TestName":          t.Name(),
		"SharedCatalog":     t.Name(),
		"SharedCatalogItem": "vapp-template",
		"Tags":              "catalog",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccCheckVcdCatalogShared, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVcdVAppVmDestroy(t.Name()),
			testAccCheckVcdStandaloneVmDestroy("test-standalone-vm", "", ""),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					// There is no need to check for much resources - the main point is to have the VMs created
					// without failures for catalog lookups or similar problems
					resource.TestCheckResourceAttrSet("vcd_vm.test-vm", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp.singleVM", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test-vm", "id"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdCatalogShared = `
resource "vcd_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "test-standalone-vm"
  catalog_name  = "{{.SharedCatalog}}"
  template_name = "{{.SharedCatalogItem}}"
  power_on      = false
}

resource "vcd_vapp" "singleVM" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name = "{{.TestName}}"
}

resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.singleVM.name
  name          = "test-vapp-vm"
  catalog_name  = "{{.SharedCatalog}}"
  template_name = "{{.SharedCatalogItem}}"
  power_on      = false

  depends_on = [vcd_vapp.singleVM]
}
`

// spawnTestOrgVdcSharedCatalog spawns an Org to be used in tests
func spawnTestOrgVdcSharedCatalog(client *VCDClient, name string) (govcd.AdminCatalog, *govcd.Vdc, *govcd.AdminOrg, *govcd.AdminOrg, error) {
	fmt.Println("# Setting up prerequisites using SDK (non Terraform definitions)")
	fmt.Printf("# Using user 'System' (%t) to prepare environment\n", client.Client.IsSysAdmin)

	existingOrg, err := client.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return govcd.AdminCatalog{}, nil, nil, nil, fmt.Errorf("error getting existing Org '%s': %s", testConfig.VCD.Org, err)
	}
	task, err := govcd.CreateOrg(client.VCDClient, name, name, name, existingOrg.AdminOrg.OrgSettings, true)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("error creating Org '%s': %s", name, err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("task failed for Org '%s' creation: %s", name, err)
	}
	newAdminOrg, err := client.GetAdminOrgByName(name)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("error getting new Org '%s': %s", name, err)
	}
	fmt.Printf("# Created new Org '%s'\n", newAdminOrg.AdminOrg.Name)

	existingVdc, err := existingOrg.GetAdminVDCByName(testConfig.Nsxt.Vdc, false)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, newAdminOrg, fmt.Errorf("error retrieving existing VDC '%s': %s", testConfig.Nsxt.Vdc, err)

	}
	vdcConfiguration := &types.VdcConfiguration{
		Name:            name + "-VDC",
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{{
			Enabled: addrOf(true),
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: getVdcProviderVdcStorageProfileHref(client, existingVdc.AdminVdc.ProviderVdcReference.HREF),
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: existingVdc.AdminVdc.NetworkPoolReference.HREF,
		},
		ProviderVdcReference: &types.Reference{
			HREF: existingVdc.AdminVdc.ProviderVdcReference.HREF,
		},
		IsEnabled:             true,
		IsThinProvision:       true,
		UsesFastProvisioning:  true,
		IsElastic:             addrOf(true),
		IncludeMemoryOverhead: addrOf(true),
	}

	vdc, err := newAdminOrg.CreateOrgVdc(vdcConfiguration)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Created new Vdc '%s' inside Org '%s'\n", vdc.Vdc.Name, newAdminOrg.AdminOrg.Name)

	catalog, err := newAdminOrg.CreateCatalog(name, name)
	if err != nil {
		return govcd.AdminCatalog{}, vdc, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Created new Catalog '%s' inside Org '%s'\n", catalog.AdminCatalog.Name, newAdminOrg.AdminOrg.Name)

	// Share new Catalog in newOrgName1 with default test Org vcd.Org
	readOnly := "ReadOnly"
	accessControl := &types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: &readOnly,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{{
				Subject: &types.LocalSubject{
					HREF: existingOrg.AdminOrg.HREF,
					Name: existingOrg.AdminOrg.Name,
					Type: types.MimeOrg,
				},
				AccessLevel: "ReadOnly",
			}},
		},
	}
	err = catalog.SetAccessControl(accessControl, false)
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Shared new Catalog '%s' with existing Org '%s'\n",
		catalog.AdminCatalog.Name, existingOrg.AdminOrg.Name)

	uploadTask, err := catalog.UploadOvf(testConfig.Ova.OvaPath, "vapp-template", "upload from test", 1024)
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, fmt.Errorf("error uploading template: %s", err)
	}

	err = uploadTask.WaitTaskCompletion()
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, fmt.Errorf("error in uploading template task: %s", err)
	}
	fmt.Printf("# Uploaded vApp template '%s' to shared new Catalog '%s' in new Org '%s' with existing Org '%s'\n",
		"vapp-template", catalog.AdminCatalog.Name, newAdminOrg.AdminOrg.Name, existingOrg.AdminOrg.Name)

	return catalog, vdc, existingOrg, newAdminOrg, nil
}

func testOrgVdcSharedCatalogCleanUp(catalog govcd.AdminCatalog, vdc *govcd.Vdc, existingOrg, newAdminOrg *govcd.AdminOrg, t *testing.T) {
	fmt.Println("# Cleaning up")
	var err error
	if catalog != (govcd.AdminCatalog{}) {
		timeout := 30 * time.Second
		start := time.Now()
		attempts := 0
		for time.Since(start) < timeout {
			err = catalog.Delete(true, true)
			if err == nil {
				break
			}
			attempts++
			fmt.Printf("## deletion attempt %d - error: %s - elapsed: %s\n", attempts, err, time.Since(start))
			time.Sleep(200 * time.Millisecond)
		}
		if err != nil {
			t.Errorf("error cleaning up catalog: %s", err)
		}
		// The catalog.Delete ignores the task returned and does not wait for the operation to complete. This code
		// was made for a particular bugfix therefore other parts of code were not altered/fixed.
		for i := 0; i < 30; i++ {
			_, err := existingOrg.GetAdminCatalogById(catalog.AdminCatalog.ID, true)
			if govcd.ContainsNotFound(err) {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
	}

	if vdc != nil {
		err = vdc.DeleteWait(true, true)
		if err != nil {
			t.Errorf("error cleaning up VDC: %s", err)
		}
	}

	if newAdminOrg != nil {
		err = newAdminOrg.Refresh()
		if err != nil {
			t.Errorf("error refreshing Org: %s", err)
		}
		err = newAdminOrg.Delete(true, true)
		if err != nil {
			t.Errorf("error cleaning up Org: %s", err)
		}
	}
}

// TestAccVcdCatalogMetadata tests metadata CRUD on catalogs
func TestAccVcdCatalogMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdCatalogMetadata, "vcd_catalog.test-catalog",
		testAccCheckVcdCatalogMetadataDatasource, "data.vcd_catalog.test-catalog-ds",
		nil, true)
}

const testAccCheckVcdCatalogMetadata = `
resource "vcd_catalog" "test-catalog" {
  org              = "{{.Org}}"
  name             = "{{.Name}}"
  delete_force     = "true"
  delete_recursive = "true"
  {{.Metadata}}
}
`

const testAccCheckVcdCatalogMetadataDatasource = `
data "vcd_catalog" "test-catalog-ds" {
  org  = vcd_catalog.test-catalog.org
  name = vcd_catalog.test-catalog.name
}
`

func TestAccVcdCatalogMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetAdminCatalogById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", id, err)
		}
		return catalog, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogMetadata, "vcd_catalog.test-catalog",
		testAccCheckVcdCatalogMetadataDatasource, "data.vcd_catalog.test-catalog-ds",
		getObjectById, nil)
}

func getVdcProviderVdcStorageProfileHref(client *VCDClient, pvdcReference string) string {
	// Filtering by name and in correct pVdc to avoid picking NSX-V VDC storage profile
	results, _ := client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdcStorageProfile",
		"filter": fmt.Sprintf("name==%s;providerVdc==%s", testConfig.VCD.NsxtProviderVdc.StorageProfile, pvdcReference),
	})
	providerVdcStorageProfileHref := results.Results.ProviderVdcStorageProfileRecord[0].HREF
	return providerVdcStorageProfileHref
}
