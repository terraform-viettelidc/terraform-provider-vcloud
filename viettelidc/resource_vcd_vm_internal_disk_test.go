//go:build vapp || vm || ALL || functional

package viettelidc

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVmInternalDisk(t *testing.T) {
	preTestChecks(t)

	// In general VM internal disks works with Org users, but since we need to create VDC with disabled fast provisioning value, we have to be sys admins
	skipIfNotSysAdmin(t)

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	internalDiskSize := 20000
	storageProfile := testConfig.VCD.ProviderVdc.StorageProfile
	diskResourceName := "disk1"
	diskSize := "13333"
	biggerDiskSize := "14333"
	busType := "sata"
	busNumber := "1"
	unitNumber := "0"
	allowReboot := true

	vappName := "TestInternalDiskVapp"
	vmName := "TestInternalDiskVm"
	vdcName := "ForInternalDiskTest"
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"FuncName":           "TestVappVmDS",
		"Tags":               "vm",
		"DiskResourceName":   diskResourceName,
		"Size":               diskSize,
		"SizeBigger":         biggerDiskSize,
		"BusType":            busType,
		"BusNumber":          busNumber,
		"UnitNumber":         unitNumber,
		"StorageProfileName": storageProfile,
		"AllowReboot":        allowReboot,

		"VdcName":                   vdcName,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		// because vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",

		"Catalog":      testSuiteCatalogName,
		"CatalogItem":  testSuiteCatalogOVAItem,
		"VappName":     vappName,
		"VmName":       vmName,
		"ComputerName": vmName + "Unique",

		"InternalDiskSize": internalDiskSize,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-IdeCreate"
	configTextIde := templateFill(sourceTestVmInternalDiskIde, params)
	params["FuncName"] = t.Name() + "-CreateALl"
	configText := templateFill(sourceTestVmInternalDisk, params)
	params["FuncName"] = t.Name() + "-Update1"
	configText_update1 := templateFill(sourceTestVmInternalDisk_Update1, params)
	params["FuncName"] = t.Name() + "-Update2"
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText+configText_update1)

	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configTextIde,
				// expected to fail for allow_vm_reboot=false and bus_type = "ide" (VM needs to be power off to add IDE disk)
				ExpectError: regexp.MustCompile(`.*The attempted operation cannot be performed in the current state \(Powered on\).*`),
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].size_in_mb", "20000"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].iops"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].bus_type", "paravirtual"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].bus_number"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].unit_number"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.TestInternalDiskVm", "internal_disk[0].storage_profile", "*"),
				),
			},
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vapp_vm.TestInternalDiskVm", "description", "description-text"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", diskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "iops", "0"),
				),
			},
			{
				Config: configText_update1,
				Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "size_in_mb", biggerDiskSize),
					resource.TestCheckResourceAttr("vcd_vapp_vm.TestInternalDiskVm", "description", "description-text"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_type", busType),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "bus_number", busNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "unit_number", unitNumber),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "allow_vm_reboot", "false"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_type", "ide"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "bus_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "unit_number", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "storage_profile", storageProfile),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName, "thin_provisioned", "true"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "size_in_mb", biggerDiskSize),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "iops", "0"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk."+diskResourceName+"_ide", "allow_vm_reboot", "true"),
				),
			},
			{
				ResourceName:      "vcd_vm_internal_disk." + diskResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVmObject(testConfig.VCD.Org, vdcName, vappName, vmName, "3000"),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"org", "vdc", "allow_vm_reboot", "thin_provisioned",
					"consolidate_disks_on_create"},
			},
		},
	})
	postTestChecks(t)
}

// we need VDC with disabled fast provisioning to edit disks
const sourceTestVmInternalDiskOrgVdcAndVM = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}" 

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

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
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 102400
    default  = true
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = false
  delete_force             = true
  delete_recursive         = true
}

resource "vcd_vapp" "{{.VappName}}" {
  org              = "{{.Org}}"
  vdc              =  vcd_org_vdc.{{.VdcName}}.name
  name = "{{.VappName}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org              = "{{.Org}}"
  vdc              =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name     = vcd_vapp.{{.VappName}}.name
  name          = "{{.VmName}}"
  description   = "description-text"
  computer_name = "{{.ComputerName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  cpu_cores     = 1

  override_template_disk {
    bus_type         = "paravirtual"
    size_in_mb       = "{{.InternalDiskSize}}"
    bus_number       = 0
    unit_number      = 0
    iops             = 0
    storage_profile  = "{{.StorageProfileName}}"
  }

  disk {
    name        = vcd_independent_disk.IndependentDisk1.name
    bus_number  = 3
    unit_number = 0
  }
}

resource "vcd_independent_disk" "IndependentDisk1" {
  org             = "{{.Org}}"
  vdc             = vcd_org_vdc.{{.VdcName}}.name
  name            = "TestAccVcdVmInternalDiskTest"
  size_in_mb      = "5"
  bus_type        = "SCSI"
  bus_sub_type    = "lsilogicsas"
  storage_profile = "{{.StorageProfileName}}"
}
`

const sourceTestVmInternalDiskIde = sourceTestVmInternalDiskOrgVdcAndVM + `
# skip-binary-test: expected to fail for allow_vm_reboot=false and bus_type = "ide"
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "ide"
  size_in_mb      = "{{.Size}}"
  bus_number      = "0"
  unit_number     = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk = sourceTestVmInternalDiskOrgVdcAndVM + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "ide"
  size_in_mb      = "{{.Size}}"
  bus_number      = "0"
  unit_number     = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true" 
}

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "{{.BusType}}"
  size_in_mb      = "{{.Size}}"
  bus_number      = "{{.BusNumber}}"
  unit_number     = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}
`

const sourceTestVmInternalDisk_Update1 = sourceTestVmInternalDiskOrgVdcAndVM + `
resource "vcd_vm_internal_disk" "{{.DiskResourceName}}" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "{{.BusType}}"
  size_in_mb      = "{{.SizeBigger}}"
  bus_number      = "{{.BusNumber}}"
  unit_number     = "{{.UnitNumber}}"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "false"
}

resource "vcd_vm_internal_disk" "{{.DiskResourceName}}_ide" {
  org             = "{{.Org}}"
  vdc             =  vcd_org_vdc.{{.VdcName}}.name
  vapp_name       = vcd_vapp.{{.VappName}}.name
  vm_name         = vcd_vapp_vm.{{.VmName}}.name
  bus_type        = "ide"
  size_in_mb      = "{{.SizeBigger}}"
  bus_number      = "0"
  unit_number     = "0"
  storage_profile = "{{.StorageProfileName}}"
  allow_vm_reboot = "true"
}
`

// TestAccVcdVmInternalDiskNvme explicitly tests NVMe disk support.
// It was introduced in VCD 10.2.1 and cannot be tested in earlier versions
func TestAccVcdVmInternalDiskNvme(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.VCD.Vdc,
		"FuncName": t.Name(),
		"Tags":     "vm",
		"BusType":  "nvme",
		"VmName":   t.Name() + "-vm",
	}
	testParamsNotEmpty(t, params)

	configTextNvme := templateFill(sourceTestVmInternalDiskOrgVdcAndVMNvme, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVcdStandaloneVmDestroy(params["VmName"].(string), params["Org"].(string), params["Vdc"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: configTextNvme,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vm.nvme", "id"),
					resource.TestCheckResourceAttrSet("vcd_vm_internal_disk.nvme", "id"),
					resource.TestCheckResourceAttr("vcd_vm_internal_disk.nvme", "bus_type", "nvme"),
				),
			},
		},
	})
	postTestChecks(t)
}

const sourceTestVmInternalDiskOrgVdcAndVMNvme = `
resource "vcd_vm" "nvme" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  power_on  = false
  name      = "{{.VmName}}"
  memory    = 512
  cpus      = 2
  cpu_cores = 1

  os_type          = "windows9Server64Guest"
  hardware_version = "vmx-18"
  computer_name    = "compName"
}

resource "vcd_vm_internal_disk" "nvme" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name       = vcd_vm.nvme.vapp_name
  vm_name         = vcd_vm.nvme.name
  bus_type        = "nvme"
  size_in_mb      = "100"
  bus_number      = "1"
  unit_number     = "0"
  allow_vm_reboot = "false"

  depends_on = [vcd_vm.nvme]
}
`

// TestAccVcdVmInternalDiskResourceNotFound checks that internal disk resource does not return error
// when parent VM is deleted, but removes it from state instead.
func TestAccVcdVmInternalDiskResourceNotFound(t *testing.T) {
	preTestChecks(t)

	// This test invokes go-vcloud-director SDK directly therefore it should not run binary tests
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,
		"TestName": t.Name(),
		"Tags":     "vm",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(sourceTestVmInternalDiskResourceNotFound, params)
	cachedvAppName := &testCachedFieldValue{}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{

			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_vm_internal_disk.destroy-test", "id"),
					cachedvAppName.cacheTestResourceFieldValue("vcd_vm.disk-vm", "vapp_name"),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()
					org, err := vcdClient.GetAdminOrgByName(params["Org"].(string))
					if err != nil {
						t.Errorf("error: could not find Org: %s", err)
					}
					vdc, err := org.GetVDCByName(params["Vdc"].(string), false)
					if err != nil {
						t.Errorf("error: could not find VDC: %s", err)
					}

					vapp, err := vdc.GetVAppByName(cachedvAppName.fieldValue, false)
					if err != nil {
						t.Errorf("could not find vApp %s: %s", cachedvAppName.fieldValue, err)
					}

					task, err := vapp.Delete()
					if err != nil {
						t.Errorf("error triggering vApp delete: %s", err)
					}

					err = task.WaitTaskCompletion()
					if err != nil {
						t.Errorf("vApp deletion task error: %s", err)
					}
				},
				// Expecting to get a non-empty plan (but not an error) because resource was removed
				// using SDK in PreConfig
				Config:             configText,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const sourceTestVmInternalDiskResourceNotFound = `
resource "vcd_vm" "disk-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "{{.TestName}}"
  computer_name = "emptyVM"
  power_on      = false
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

resource "vcd_vm_internal_disk" "destroy-test" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name       = vcd_vm.disk-vm.vapp_name
  vm_name         = vcd_vm.disk-vm.name
  bus_type        = "sata"
  size_in_mb      = "100"
  bus_number      = "1"
  unit_number     = "0"
}
`
