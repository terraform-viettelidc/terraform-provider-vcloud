//go:build (standaloneVm || vm || ALL || functional) && !skipStandaloneVm

package viettelidc

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneVm_HardwareVirtualization(t *testing.T) {
	preTestChecks(t)
	var standaloneVmName = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())

	var params = StringMap{
		"Org":                          testConfig.VCD.Org,
		"Vdc":                          testConfig.VCD.Vdc,
		"EdgeGateway":                  testConfig.Networking.EdgeGateway,
		"NetworkName":                  "TestAccVcdVmNetHwVirt",
		"Catalog":                      testSuiteCatalogName,
		"CatalogItem":                  testSuiteCatalogOVAItem,
		"VmName":                       standaloneVmName,
		"ExposeHardwareVirtualization": "false",
		"Tags":                         "standaloneVm vm",
	}
	testParamsNotEmpty(t, params)

	configTextStep0 := templateFill(testAccCheckVcdVm_hardwareVirtualization, params)

	params["ExposeHardwareVirtualization"] = "true"
	params["FuncName"] = t.Name() + "-step1"
	configTextStep1 := templateFill(testAccCheckVcdVm_hardwareVirtualization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep0)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroy(standaloneVmName, "", ""),
		Steps: []resource.TestStep{
			{
				Config: configTextStep0,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "expose_hardware_virtualization", "false"),
				),
			},
			{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdStandaloneVmExists(standaloneVmName, "vcd_vm."+standaloneVmName, "", ""),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "name", standaloneVmName),
					resource.TestCheckResourceAttr(
						"vcd_vm."+standaloneVmName, "expose_hardware_virtualization", "true"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVm_hardwareVirtualization = `

resource "vcd_vm" "{{.VmName}}" {
  org                            = "{{.Org}}"
  vdc                            = "{{.Vdc}}"
  name                           = "{{.VmName}}"
  catalog_name                   = "{{.Catalog}}"
  template_name                  = "{{.CatalogItem}}"
  memory                         = 384
  cpus                           = 2
  cpu_cores                      = 1
  power_on                       = false
  expose_hardware_virtualization = {{.ExposeHardwareVirtualization}}
}
`
