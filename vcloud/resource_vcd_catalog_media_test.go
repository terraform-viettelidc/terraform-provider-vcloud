//go:build catalog || ALL || functional

package vcloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdCatalogMediaBasic(t *testing.T) {
	preTestChecks(t)

	var TestAccVcdCatalogMedia = "TestAccVcdCatalogMediaBasic"
	var TestAccVcdCatalogMediaDescription = "TestAccVcdCatalogMediaBasicDescription"
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Catalog":          testConfig.VCD.Catalog.Name,
		"CatalogMediaName": TestAccVcdCatalogMedia,
		"Description":      TestAccVcdCatalogMediaDescription,
		"MediaPath":        testConfig.Media.MediaPath,
		"MediaPath2":       getCurrentDir() + "/resource_vcd_catalog_media.go",
		"UploadPieceSize":  testConfig.Media.UploadPieceSize,
		"UploadProgress":   testConfig.Media.UploadProgress,
		"Tags":             "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogMediaBasic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVcdCatalogMediaUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCatalogMediaDestroy(TestAccVcdCatalogMedia),
			testAccCheckCatalogMediaDestroy(TestAccVcdCatalogMedia+"2"),
			testAccCheckCatalogMediaDestroy(TestAccVcdCatalogMedia+"2-update"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdCatalogMedia),
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media.text_file"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "name", TestAccVcdCatalogMedia),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media.text_file", "name", TestAccVcdCatalogMedia+"-2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "description", TestAccVcdCatalogMediaDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media.text_file", "description", TestAccVcdCatalogMediaDescription+" 2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata", "mediaItem Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata2", "mediaItem Metadata2"),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("creation_date", regexp.MustCompile(`^^\d{4}-\d{2}-\d{2}.*`)),
					resource.TestCheckOutput("status", "RESOLVED"),
					resource.TestMatchOutput("storage_profile_name", regexp.MustCompile(`^\S+`)),
					testCheckMediaNonStringOutputs(),
				),
			},
			{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdCatalogMedia),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "name", TestAccVcdCatalogMedia),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media.text_file", "name", TestAccVcdCatalogMedia+"-2-update"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "description", TestAccVcdCatalogMediaDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata", "mediaItem Metadata v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata2", "mediaItem Metadata2 v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata3", "mediaItem Metadata3"),
				),
			},
			{
				ResourceName:      "vcd_catalog_media." + TestAccVcdCatalogMedia,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgCatalogObject(TestAccVcdCatalogMedia),
				// These fields can't be retrieved from catalog media data
				ImportStateVerifyIgnore: []string{"media_path", "upload_piece_size", "show_upload_progress", "upload_any_file"},
			},
		},
	})
	postTestChecks(t)
}

func testCheckMediaNonStringOutputs() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["is_iso"].Value != "true" {
			return fmt.Errorf("is_iso value didn't match")
		}

		if outputs["is_published"].Value != "false" {
			return fmt.Errorf("is_published value didn't match")
		}

		if !regexp.MustCompile(`^\d+$`).MatchString(fmt.Sprintf("%s", outputs["size"].Value)) {
			return fmt.Errorf("size value isn't an integer")
		}

		return nil
	}
}

func testAccCheckVcdCatalogMediaExists(mediaName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		catalogMediaRs, ok := s.RootModule().Resources[mediaName]
		if !ok {
			return fmt.Errorf("not found: %s", mediaName)
		}

		if catalogMediaRs.Primary.ID == "" {
			return fmt.Errorf("catalog media ID is not set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.Name, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%s)", testConfig.VCD.Catalog.Name, err)
		}

		foundMedia, err := catalog.GetMediaByName(catalogMediaRs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("catalog media %s does not exist (%#v)", catalogMediaRs.Primary.ID, foundMedia.Media)
		}

		return nil
	}
}

func testAccCheckCatalogMediaDestroy(mediaName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_catalog_media" && rs.Primary.Attributes["name"] != mediaName {
				continue
			}

			adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
			if err != nil {
				return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
			}

			catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.Name, false)
			if err != nil {
				return fmt.Errorf("catalog query %s ended with error: %#v", rs.Primary.ID, err)
			}

			_, err = catalog.GetMediaByName(mediaName, false)
			if err == nil {
				return fmt.Errorf("catalog media %s still exists", mediaName)
			}
		}

		return nil
	}
}

const testAccCheckVcdCatalogMediaBasic = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = {{.UploadProgress}}

  metadata = {
    mediaItem_metadata = "mediaItem Metadata"
    mediaItem_metadata2 = "mediaItem Metadata2"
  }
}

resource "vcd_catalog_media"  "text_file" {
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.CatalogMediaName}}-2"
  description          = "{{.Description}} 2"
  media_path           = "{{.MediaPath2}}"
  upload_any_file      = true
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = {{.UploadProgress}}
}

output "creation_date" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.creation_date
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "is_iso" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.is_iso
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "owner_name" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.owner_name
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "is_published" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.is_published
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "size" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.size
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "status" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.status
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}
output "storage_profile_name" {
  value = vcd_catalog_media.{{.CatalogMediaName}}.storage_profile_name
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}`

const testAccCheckVcdCatalogMediaUpdate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    mediaItem_metadata  = "mediaItem Metadata v2"
    mediaItem_metadata2 = "mediaItem Metadata2 v2"
    mediaItem_metadata3 = "mediaItem Metadata3"
  }
}

resource "vcd_catalog_media"  "text_file" {
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.CatalogMediaName}}-2-update"
  description          = "{{.Description}} 2 update"
  media_path           = "{{.MediaPath2}}"
  upload_any_file      = true
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = {{.UploadProgress}}
}
`

// TestAccVcdCatalogMediaMetadata tests metadata CRUD on catalog media
func TestAccVcdCatalogMediaMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdCatalogMediaMetadata, "vcd_catalog_media.test-catalog-media",
		testAccCheckVcdCatalogMediaMetadataDatasource, "data.vcd_catalog_media.test-catalog-media-ds",
		StringMap{
			"Catalog":   testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"MediaPath": testConfig.Media.MediaPath,
		}, true)
}

const testAccCheckVcdCatalogMediaMetadata = `
resource "vcd_catalog_media" "test-catalog-media" {
  org        = "{{.Org}}"
  name       = "{{.Name}}"
  catalog    = "{{.Catalog}}"
  media_path = "{{.MediaPath}}"
  {{.Metadata}}
}
`

const testAccCheckVcdCatalogMediaMetadataDatasource = `
data "vcd_catalog_media" "test-catalog-media-ds" {
  org        = vcd_catalog_media.test-catalog-media.org
  catalog_id = vcd_catalog_media.test-catalog-media.catalog_id
  name       = vcd_catalog_media.test-catalog-media.name
}
`

func TestAccVcdCatalogMediaMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", testConfig.VCD.Catalog.NsxtBackedCatalogName, err)
		}
		media, err := catalog.GetMediaById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Media '%s': %s", id, err)
		}
		return media, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogMediaMetadata, "vcd_catalog_media.test-catalog-media",
		testAccCheckVcdCatalogMediaMetadataDatasource, "data.vcd_catalog_media.test-catalog-media-ds",
		getObjectById, StringMap{
			"Catalog":   testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"MediaPath": testConfig.Media.MediaPath,
		})
}
