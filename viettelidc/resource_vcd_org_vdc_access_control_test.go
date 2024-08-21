//go:build vdc || ALL || functional

package viettelidc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdOrgVdcAccessControl(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	userName1 := strings.ToLower(t.Name())
	userName2 := strings.ToLower(t.Name()) + "2"
	accessControlName := "test-access-control"

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.Nsxt.Vdc,
		"AccessControlName":  accessControlName,
		"AccessControlName2": accessControlName + "2",
		"UserName":           userName1,
		"UserName2":          userName2,
		"Password":           "CHANGE-ME",
		"RoleName":           govcd.OrgUserRoleOrganizationAdministrator,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText := templateFill(testAccCheckVcdAccessControlStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccCheckVcdAccessControlStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVDCControlAccessDestroy(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithEverybody(testConfig.Nsxt.Vdc),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					assertVdcAccessControlIsSharedWithSpecificUser(userName1, testConfig.Nsxt.Vdc),
					assertVdcAccessControlIsSharedWithSpecificUser(userName2, testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      fmt.Sprintf("vcd_org_vdc_access_control.%s", accessControlName),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.Nsxt.Vdc,
			},
		},
	})
}

const testAccCheckVcdAccessControlStep1 = `
resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
`

const testAccCheckVcdAccessControlStep2 = `
resource "vcd_org_user" "{{.UserName}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName}}"
  password       = "{{.Password}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}

resource "vcd_org_user" "{{.UserName2}}" {
  org            = "{{.Org}}"
  name           = "{{.UserName2}}"
  password       = "{{.Password}}"
  role           = "{{.RoleName}}"
  take_ownership = true
}

resource "vcd_org_vdc_access_control" "{{.AccessControlName}}" {
  org                   = "{{.Org}}"
  vdc                   = "{{.Vdc}}"
  shared_with_everyone  = false
  shared_with {
    user_id             = vcd_org_user.{{.UserName}}.id
    access_level        = "ReadOnly"
  }
  shared_with {
    user_id             = vcd_org_user.{{.UserName2}}.id
    access_level        = "ReadOnly"
  }
}
`

func assertVdcAccessControlIsSharedWithEverybody(vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if !controlAccessParams.IsSharedToEveryone {
			return fmt.Errorf("this VDC was expected to be shared with everyone but it is not")
		}

		return nil
	}
}

func assertVdcAccessControlIsSharedWithSpecificUser(userName string, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf("error retrieving Org %s - %s", testConfig.VCD.Org, err)
		}

		controlAccessParams, err := vdc.GetControlAccess(true)
		if err != nil {
			return fmt.Errorf("error retrieving VDC controll access parameters - %s", err)
		}

		if controlAccessParams.AccessSettings == nil {
			return fmt.Errorf("there are not users configured for sharing in this VDC and they were expected to be")
		}

		for _, accessControlEntry := range controlAccessParams.AccessSettings.AccessSetting {
			if accessControlEntry.Subject.Name == userName {
				return nil
			}
		}

		return fmt.Errorf("userName %s wasn't found in VDC %s and it was expected to be", userName, vdc.Vdc.Name)
	}
}
