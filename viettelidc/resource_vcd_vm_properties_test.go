//go:build (standaloneVm || vm || ALL || functional) && !skipStandaloneVm

package viettelidc

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneVmProperties(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VmName":      standaloneVmName,
		"Tags":        "vm standaloneVm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVm_properties, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccCheckVcdVm_propertiesUpdate, params)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccCheckVcdVm_propertiesRemove, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.another.subkey`, "another-value"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckNoResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.hostname`),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.another.subkey`, "new-value"),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.guest.third.subkey`, "third-value"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr("vcd_vm."+standaloneVmName, `guest_properties.%`, "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVm_properties = `
resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}
`

const testAccCheckVcdVm_propertiesUpdate = `
resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  guest_properties = {
	"guest.another.subkey" = "new-value"
	"guest.third.subkey"   = "third-value"
  }
}
`

const testAccCheckVcdVm_propertiesRemove = `
resource "vcd_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1
}
`
