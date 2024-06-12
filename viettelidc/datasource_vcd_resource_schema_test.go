//go:build ALL || functional

package viettelidc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceResourceSchema(t *testing.T) {
	preTestChecks(t)

	for name := range globalResourceMap {
		t.Run(name, func(t *testing.T) { runResourceSchemaTest(name, t) })
	}
	postTestChecks(t)
}

func runResourceSchemaTest(name string, t *testing.T) {

	var data = StringMap{
		"ResName":  name,
		"ResType":  name,
		"FuncName": fmt.Sprintf("ResourceSchema-%s", name),
	}
	configText := templateFill(testAccCheckVcdDatasourceStructure, data)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vcd_resource_schema."+name, "name", name),
				),
			},
		},
	})
}

const testAccCheckVcdDatasourceStructure = `
data "vcd_resource_schema" "{{.ResName}}" {
  name          = "{{.ResName}}"
  resource_type = "{{.ResType}}"
}

output "resources" {
  value = data.vcd_resource_schema.{{.ResName}}
}
`
