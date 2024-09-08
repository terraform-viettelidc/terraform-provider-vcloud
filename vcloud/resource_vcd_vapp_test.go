//go:build vapp || ALL || functional

package vcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVApp_Basic(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp
	var vappName = "TestAccVcdVAppVapp"
	var vappDescription = "A long description containing some text."
	var vappUpdateDescription = "A shorter description."

	secondsInDay := 60 * 60 * 24
	runtimeLease := secondsInDay * 30
	storageLease := secondsInDay * 3

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.Nsxt.Vdc,
		"NetworkName":     "TestAccVcdVAppNet",
		"NetworkName2":    "TestAccVcdVAppNet2",
		"NetworkName3":    "TestAccVcdVAppNet3",
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"VappName":        vappName,
		"VappDescription": vappDescription,
		"FuncName":        "TestAccVcdVApp_Basic",
		"RuntimeLease":    runtimeLease,
		"StorageLease":    storageLease,
		"Tags":            "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVApp_basic, params)

	params["FuncName"] = "TestAccCheckVcdVApp_update"
	params["VappDescription"] = vappUpdateDescription
	configTextUpdate := templateFill(testAccCheckVcdVApp_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp."+vappName, &vapp),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "name", vappName),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "description", vappDescription),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "status", "1"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "metadata.vapp_metadata", "vApp Metadata."),
					resource.TestMatchResourceAttr("vcd_vapp."+vappName, "href",
						getUuidRegex("", "")),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "metadata.vapp_metadata", "vApp Metadata."),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `guest_properties.guest.another.subkey`, "another-value"),

					// For an empty lease section, check that the lease is inherited from the organization
					resource.TestCheckResourceAttrPair("data.vcd_org."+testConfig.VCD.Org, "vapp_lease.0.maximum_runtime_lease_in_sec",
						"vcd_vapp."+vappName, "lease.0.runtime_lease_in_sec"),
					resource.TestCheckResourceAttrPair("data.vcd_org."+testConfig.VCD.Org, "vapp_lease.0.maximum_storage_lease_in_sec",
						"vcd_vapp."+vappName, "lease.0.storage_lease_in_sec"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp."+vappName, &vapp),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "name", vappName),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "description", vappUpdateDescription),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "power_on", "true"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "status", "4"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, "metadata.vapp_metadata", "vApp Metadata updated"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `guest_properties.guest.another.subkey`, "new-value"),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `guest_properties.guest.third.subkey`, "third-value"),

					// Check that the updated lease corresponds with the new parameters
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `lease.0.runtime_lease_in_sec`, fmt.Sprintf("%d", runtimeLease)),
					resource.TestCheckResourceAttr("vcd_vapp."+vappName, `lease.0.storage_lease_in_sec`, fmt.Sprintf("%d", storageLease)),
				),
			},
			{
				ResourceName:      "vcd_vapp." + vappName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(vappName),
				// These fields can't be retrieved from user data
				ImportStateVerifyIgnore: []string{"power_on"},
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdVAppExists(n string, vapp *govcd.VApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		newVapp, err := vdc.GetVAppByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return err
		}

		*vapp = *newVapp

		return nil
	}
}

func testAccCheckVcdVAppDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vapp" {
			continue
		}

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetVAppByNameOrId(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

func init() {
	testingTags["vapp"] = "resource_vcd_vapp_test.go"
}

const testAccCheckVcdVApp_basic = `

data "vcd_org" "{{.Org}}" {
	name = "{{.Org}}"
}

resource "vcd_vapp" "{{.VappName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VappName}}"
  description   = "{{.VappDescription}}"

  metadata = {
    vapp_metadata = "vApp Metadata."
  }

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}

# needed to check power on on update in next step
resource "vcd_vapp_vm" "test_vm1" {
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "test_vm1"
  memory        = 512
  cpus          = 1
  cpu_cores     = 1 

  os_type                        = "rhel4Guest"
  hardware_version               = "vmx-14"
  computer_name                  = "compNameUp"
}
`

const testAccCheckVcdVApp_update = `
# skip-binary-test: only for updates
resource "vcd_vapp" "{{.VappName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VappName}}"
  description   = "{{.VappDescription}}"

  metadata = {
    vapp_metadata = "vApp Metadata updated"
  }

  guest_properties = {
	"guest.another.subkey" = "new-value"
	"guest.third.subkey"   = "third-value"
  }

  lease {
    runtime_lease_in_sec = {{.RuntimeLease}}
    storage_lease_in_sec = {{.StorageLease}}
  }

  power_on = true
}

# vApp power on won't work if vApp doesn't have VM
resource "vcd_vapp_vm" "test_vm1" {
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "test_vm1"
  memory        = 512
  cpus          = 1
  cpu_cores     = 1 

  os_type                        = "rhel4Guest"
  hardware_version               = "vmx-14"
  computer_name                  = "compNameUp"
}
`

// TestAccVcdVAppMetadata tests metadata CRUD on vApps
func TestAccVcdVAppMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdVAppMetadata, "vcd_vapp.test-vapp",
		testAccCheckVcdVAppMetadataDatasource, "data.vcd_vapp.test-vapp-ds",
		nil, true)
}

const testAccCheckVcdVAppMetadata = `
resource "vcd_vapp" "test-vapp" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.Name}}"
  {{.Metadata}}
}
`

const testAccCheckVcdVAppMetadataDatasource = `
data "vcd_vapp" "test-vapp-ds" {
  org  = vcd_vapp.test-vapp.org
  vdc  = vcd_vapp.test-vapp.vdc
  name = vcd_vapp.test-vapp.name
}
`

func TestAccVcdVAppMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		vApp, err := vdc.GetVAppById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp '%s': %s", id, err)
		}
		return vApp, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdVAppMetadata, "vcd_vapp.test-vapp",
		testAccCheckVcdVAppMetadataDatasource, "data.vcd_vapp.test-vapp-ds",
		getObjectById, nil)
}
