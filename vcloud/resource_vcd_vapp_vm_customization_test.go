//go:build vapp || vm || ALL || functional

package vcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"testing"
)

// TestAccVcdVAppVmUpdateCustomization tests that setting attribute customizaton.force to `true`
// during update triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdVAppVmUpdateCustomization(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"VappPowerOn": "false",
		"Tags":        "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configTextVM := templateFill(testAccCheckVcdVAppVmUpdateCustomization, params)

	params["FuncName"] = t.Name() + "-step2"
	params["Customization"] = "true"
	params["VappPowerOn"] = "true"
	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	params["FuncName"] = t.Name() + "-step3"
	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	params["Customization"] = "false"
	params["VappPowerOn"] = "false"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppVmCreateCustomizationPowerOff, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create without customization flag
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVMCustomization("vcd_vapp_vm.test-vm", false),
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "1"),
				),
			},
			// Step 1 - Update - change network configuration and force customization
			{
				Config: configTextVMUpdateStep1,
				// The plan should never be empty because force works as a flag and every update triggers "update"
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVMCustomization("vcd_vapp_vm.test-vm", true),
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.force", "false"),
				),
			},
			{
				Config: configTextVMUpdateStep2,
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdVAppVmCreateCustomization tests that setting attribute customizaton.force to `true`
// during create triggers VM customization and waits until it is completed.
// It is important to wait until the operation is completed to test what VM was properly handled before triggering
// power on and force customization. (VM must be un-deployed for customization to work, otherwise it would stay in
// "GC_PENDING" state for long time)
func TestAccVcdVAppVmCreateCustomization(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"Catalog":       testSuiteCatalogName,
		"CatalogItem":   testSuiteCatalogOVAItem,
		"VAppName":      netVappName,
		"VMName":        netVmName1,
		"VappPowerOn":   "true",
		"Tags":          "vapp vm",
		"Customization": "true",
	}
	testParamsNotEmpty(t, params)

	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	params["FuncName"] = t.Name() + "-step2"
	params["Customization"] = false
	params["VappPowerOn"] = false
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create new VM and force customization initially
			{
				Config: configTextVMUpdateStep1,
				// The plan should never be empty because force works as a flag and every update triggers "update"
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "1"),
					// Always store 'customization.0.force=false' in statefile so that a diff is always triggered
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.force", "false"),
				),
			},
			{ // Power off vApp
				Config: configTextVMUpdateStep2,
			},
		},
	})
	postTestChecks(t)
}

// testAccCheckVcdVMCustomization functions acts as a check and a function which waits until
// the VM exits its original "GC_PENDING" state after provisioning. This is needed in order to
// be able to check that setting customization.force flag to `true` actually has impact on VM
// settings.
func testAccCheckVcdVMCustomization(node string, customizationPending bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.Attributes["vapp_name"] == "" {
			return fmt.Errorf("no vApp name specified: %+#v", rs)
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("no VM name specified: %+#v", rs)
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(rs.Primary.Attributes["vapp_name"], false)
		if err != nil {
			return err
		}

		vm, err := vapp.GetVMByName(rs.Primary.Attributes["name"], false)

		if err != nil {
			return err
		}

		// When force customization was not explicitly triggered - wait until the VM exits from its original GC_PENDING
		// state after provisioning. This takes some time until the VM boots starts guest tools and reports success.
		if !customizationPending {
			// Not using maxRetryTimeout for timeout here because it would force for maxRetryTimeout to be quite long
			// time by default as it takes some time (around 150s during testing) for Photon OS to boot
			// first time and get rid of "GC_PENDING" state
			err = vm.BlockWhileGuestCustomizationStatus("GC_PENDING", minIfLess(300, conn.Client.MaxRetryTimeout))
			if err != nil {
				return err
			}
		}
		customizationStatus, err := vm.GetGuestCustomizationStatus()
		if err != nil {
			return fmt.Errorf("unable to get VM customization status: %s", err)
		}
		// At the stage where "GC_PENDING" should not be set. The state should be something else or this
		// is an error
		if !customizationPending && customizationStatus == "GC_PENDING" {
			return fmt.Errorf("customizationStatus should not be in pending state for vm %s", vm.VM.Name)
		}

		// Customization status of "GC_PENDING" is expected now and it is an error if something else is set
		if customizationPending && customizationStatus != "GC_PENDING" {
			return fmt.Errorf("customizationStatus should be 'GC_PENDING'instead of '%s' for vm %s",
				customizationStatus, vm.VM.Name)
		}

		if customizationPending && customizationStatus == "GC_PENDING" {
			err = vm.BlockWhileGuestCustomizationStatus("GC_PENDING", minIfLess(300, conn.Client.MaxRetryTimeout))
			if err != nil {
				return fmt.Errorf("timed out waiting for VM %s to leave 'GC_PENDING' state: %s", vm.VM.Name, err)
			}
		}

		return nil
	}
}

const testAccCheckVcdVAppVmCustomizationShared = `
resource "vcd_vapp" "test-vapp" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = {{.VappPowerOn}}

  name       = "{{.VAppName}}"
}

resource "vcd_vapp_network" "vappNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name       = "vapp-net"
  vapp_name  = vcd_vapp.test-vapp.name
  gateway    = "192.168.2.1"
  netmask    = "255.255.255.0"
  dns1       = "192.168.2.1"
  dns2       = "192.168.2.2"
  dns_suffix = "mybiz.biz"

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}
`

const testAccCheckVcdVAppVmUpdateCustomization = testAccCheckVcdVAppVmCustomizationShared + `
# skip-binary-test: vApp network removal from powered on vApp fails
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

}
`

const testAccCheckVcdVAppVmCreateCustomization = testAccCheckVcdVAppVmCustomizationShared + `
{{.SkipTest}}
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
    force = {{.Customization}}
  }

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappNet.name
    ip_allocation_mode = "POOL"
  }
}
`

const testAccCheckVcdVAppVmCreateCustomizationPowerOff = testAccCheckVcdVAppVmCustomizationShared + `
{{.SkipTest}}
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on = false

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappNet.name
    ip_allocation_mode = "POOL"
  }
}
`

// TestAccVcdVAppVmCreateCustomizationFalse checks if VM is booted up successfully when  customization.force=true.
// This test covers a previous bug.
func TestAccVcdVAppVmCreateCustomizationFalse(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName string = t.Name()
		netVmName1  string = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"Catalog":       testSuiteCatalogName,
		"CatalogItem":   testSuiteCatalogOVAItem,
		"VAppName":      netVappName,
		"VMName":        netVmName1,
		"VappPowerOn":   "true",
		"Tags":          "vapp vm",
		"Customization": "false",
		"SkipTest":      " ",
	}
	testParamsNotEmpty(t, params)

	params["SkipTest"] = "# skip-binary-test: vApp network removal from powered on vApp fails"
	configTextVM := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	params["SkipTest"] = "# skip-binary-test: customization.force=true must always request for update"
	params["VappPowerOn"] = false
	params["FuncName"] = t.Name() + "-step2"
	configTextVMUpdateStep2 := templateFill(testAccCheckVcdVAppVmCreateCustomization, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create new VM and set set customization.force=false
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
				),
			},
			{
				Config: configTextVMUpdateStep2,
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdVAppVmCustomizationSettings tests out possible customization options
func TestAccVcdVAppVmCustomizationSettings(t *testing.T) {
	preTestChecks(t)
	var (
		vapp        govcd.VApp
		vm          govcd.VM
		netVappName = t.Name()
		netVmName1  = t.Name() + "VM"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VappPowerOn": "true",
		"VMName":      netVmName1,
		"Tags":        "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configTextVM := templateFill(testAccCheckVcdVAppVmUpdateCustomizationSettings, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMStep1 := templateFill(testAccCheckVcdVAppVmUpdateCustomizationSettingsStep1, params)

	params["FuncName"] = t.Name() + "-step2"
	params["VappPowerOn"] = "false"
	configTextVMStep2 := templateFill(testAccCheckVcdVAppVmUpdateCustomizationSettingsStep2, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(netVappName),
		Steps: []resource.TestStep{
			// Step 1
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.change_sid", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.allow_local_admin_password", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.auto_generate_password", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.must_change_password_on_first_login", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm", "customization.0.number_of_auto_logons", "4"),
				),
			},
			// Step 2 - join org domain (does not fail because enabled=false even though OS is not windows)
			{
				// Taint:  []string{"vcd_vapp_vm.test-vm"},
				// Taint does not work in SDK 2.1.0 therefore every test step has resource address changed to force
				// recreation of the VM
				Config: configTextVMStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm-step2", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "customization.0.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "customization.0.admin_password", "some password"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "customization.0.join_domain", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step2", "customization.0.join_org_domain", "true"),
				),
			},
			// Step 3 - join org domain enabled
			{
				// Taint:  []string{"vcd_vapp_vm.test-vm"},
				// Taint does not work in SDK 2.1.0 therefore every test step has resource address changed to force
				// recreation of the VM
				Config: configTextVMStep2,
				// Our testing suite does not have Windows OS to actually try domain join so the point of this test is
				// to prove that values are actually set and try to be applied on vCD.
				ExpectError: regexp.MustCompile(`Join Domain is not supported for OS type .*`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExists(netVappName, netVmName1, "vcd_vapp_vm.test-vm-step3", &vapp, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "name", netVmName1),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "network.#", "0"),

					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.#", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.join_domain", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.join_domain_name", "UnrealDomain"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.join_domain_user", "NoUser"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.join_domain_password", "NoPass"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.test-vm-step3", "customization.0.join_domain_account_ou", "ou=IT,dc=some,dc=com"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppVmUpdateCustomizationSettings = testAccCheckVcdVAppVmCustomizationShared + `
# skip-binary-test: vApp network removal from powered on vApp fails
resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled                             = true
	change_sid                          = true
	allow_local_admin_password          = true
	must_change_password_on_first_login = true
	auto_generate_password              = true
	number_of_auto_logons               = 4
  }
}
`

const testAccCheckVcdVAppVmUpdateCustomizationSettingsStep1 = testAccCheckVcdVAppVmCustomizationShared + `
# skip-binary-test: it will fail on purpose
resource "vcd_vapp_vm" "test-vm-step2" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled                    = false
	allow_local_admin_password = true
	admin_password             = "some password"
	auto_generate_password     = false
	join_domain                = true
	join_org_domain            = true
  }
}
`

const testAccCheckVcdVAppVmUpdateCustomizationSettingsStep2 = testAccCheckVcdVAppVmCustomizationShared + `
# skip-binary-test: it will fail on purpose
resource "vcd_vapp_vm" "test-vm-step3" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.test-vapp.name
  name          = "{{.VMName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 512
  cpus          = 2
  cpu_cores     = 1

  customization {
	enabled                = true
	join_domain            = true
	join_domain_name       = "UnrealDomain"
	join_domain_user       = "NoUser"
	join_domain_password   = "NoPass"
	join_domain_account_ou = "ou=IT,dc=some,dc=com"
  }
}
`
