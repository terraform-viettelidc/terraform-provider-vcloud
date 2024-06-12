package viettelidc

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func vcdVmDS(vmType typeOfVm) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"vapp_name": {
			Type:        schema.TypeString,
			Required:    vmType == vappVmType,
			Optional:    vmType == standaloneVmType,
			Computed:    vmType == standaloneVmType,
			Description: "The vApp this VM belongs to - Required, unless it is a standalone VM",
		},
		"vapp_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of parent vApp",
		},
		"vm_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("Type of VM: either '%s' or '%s'", vappVmType, standaloneVmType),
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A name for the VM, unique within the vApp",
		},
		"org": {
			Type:     schema.TypeString,
			Optional: true,
			Description: "The name of organization to use, optional if defined at provider " +
				"level. Useful when connected as sysadmin working across different organizations",
		},
		"vdc": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of VDC to use, optional if defined at provider level",
		},
		"computer_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Computer name assigned to this virtual machine",
		},
		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The VM description",
			// Currently, this field has the description of the OVA used to create the VM
		},
		"memory": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The amount of RAM (in MB) to allocate to the VM",
		},
		"memory_reservation": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The amount of RAM (in MB) reservation on the underlying virtualization infrastructure",
		},
		"memory_priority": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload",
		},
		"memory_shares": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Custom priority for the resource",
		},
		"memory_limit": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The limit for how much of memory can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited",
		},
		"cpus": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of virtual CPUs to allocate to the VM",
		},
		"cpu_cores": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of cores per socket",
		},
		"cpu_reservation": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The amount of MHz reservation on the underlying virtualization infrastructure",
		},
		"cpu_priority": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload",
		},
		"cpu_shares": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Custom priority for the resource",
		},
		"cpu_limit": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The limit for how much of CPU can be consumed on the underlying virtualization infrastructure. This is only valid when the resource allocation is not unlimited",
		},
		"metadata": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Key value map of metadata to assign to this VM",
			Deprecated:  "Use metadata_entry instead",
		},
		"metadata_entry": metadataEntryDatasourceSchema("VM"),
		"href": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "VM Hyper Reference",
		},
		"storage_profile": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Storage profile used with the VM",
		},
		"os_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Operating System type.",
		},
		"firmware": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Firmware of the VM",
		},
		"hardware_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Virtual Hardware Version.",
		},
		"network_dhcp_wait_seconds": {
			Optional:     true,
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Description: "Optional number of seconds to try and wait for DHCP IP (valid for " +
				"'network' block only)",
		},
		"network": {
			Computed:    true,
			Type:        schema.TypeList,
			Description: "",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "Network type",
					},
					"ip_allocation_mode": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "IP address allocation mode.",
					},
					"name": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "Name of the network this VM should connect to.",
					},
					"ip": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "IP of the VM. Settings depend on `ip_allocation_mode`",
					},
					"is_primary": {
						Computed:    true,
						Type:        schema.TypeBool,
						Description: "Set to true if network interface should be primary. First network card in the list will be primary by default",
					},
					"mac": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "Mac address of network interface",
					},
					"adapter_type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Network card adapter type. (e.g. 'E1000', 'E1000E', 'SRIOVETHERNETCARD', 'VMXNET3', 'PCNet32')",
					},
					"connected": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "It defines if NIC is connected or not.",
					},
				},
			},
		},
		"disk": {
			Type: schema.TypeSet,
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Independent disk name",
				},
				"bus_number": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Bus number on which to place the disk controller",
				},
				"unit_number": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Unit number (slot) on the bus specified by BusNumber",
				},
				"size_in_mb": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The size of the disk in MB.",
				},
			}},
			Computed: true,
			Set:      resourceVcdVmIndependentDiskHash,
		},
		"internal_disk": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A block will show internal disk details",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"disk_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The disk ID.",
				},
				"bus_type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The type of disk controller. Possible values: ide, parallel( LSI Logic Parallel SCSI), sas(LSI Logic SAS (SCSI)), paravirtual(Paravirtual (SCSI)), sata, nvme",
				},
				"size_in_mb": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The size of the disk in MB.",
				},
				"bus_number": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The number of the SCSI or IDE controller itself.",
				},
				"unit_number": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The device number on the SCSI or IDE controller of the disk.",
				},
				"thin_provisioned": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Specifies whether the disk storage is pre-allocated or allocated on demand.",
				},
				"iops": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Specifies the IOPS for the disk. Default is 0.",
				},
				"storage_profile": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Storage profile to override the VM default one",
				},
			}},
		},
		"boot_options": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A block defining the boot options of a VM",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"boot_delay": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Number of milliseconds to wait between powering-on and booting the VM",
				},
				"boot_retry_delay": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Delay in milliseconds before a boot retry. Only works if 'boot_retry_enabled' is set to true.",
				},
				"boot_retry_enabled": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "If set to true, a VM that fails to boot will try again after the 'boot_retry_delay' time period has expired",
				},
				"efi_secure_boot": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "If set to true, enables EFI Secure Boot for the VM. Can only be changed when the VM is powered off.",
				},
				"enter_bios_setup_on_next_boot": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "If set to true, the VM will enter BIOS setup on boot.",
				},
			},
			},
		},
		"expose_hardware_virtualization": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Expose hardware-assisted CPU virtualization to guest OS.",
		},
		"guest_properties": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Key/value settings for guest properties",
		},

		"customization": {
			Computed:    true,
			Type:        schema.TypeList,
			Description: "Guest customization block",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"force": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "'true' value will cause the VM to reboot on every 'apply' operation",
					},
					"enabled": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "'true' value will enable guest customization. It may occur on first boot or when 'force' is used",
					},
					"change_sid": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "'true' value will change SID. Applicable only for Windows VMs",
					},
					"allow_local_admin_password": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Allow local administrator password",
					},
					"must_change_password_on_first_login": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Require Administrator to change password on first login",
					},
					"auto_generate_password": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Auto generate password",
					},
					"admin_password": {
						Type:        schema.TypeString,
						Computed:    true,
						Sensitive:   true,
						Description: "Manually specify admin password",
					},
					"number_of_auto_logons": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of times to log on automatically",
					},
					"join_domain": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Enable this VM to join a domain",
					},
					"join_org_domain": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Use organization's domain for joining",
					},
					"join_domain_name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Custom domain name for join",
					},
					"join_domain_user": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Username for custom domain name join",
					},
					"join_domain_password": {
						Type:        schema.TypeString,
						Computed:    true,
						Sensitive:   true,
						Description: "Password for custom domain name join",
					},
					"join_domain_account_ou": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Account organizational unit for domain name join",
					},
					"initscript": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Script to run on initial boot or with customization.force=true set",
					},
				},
			},
		},
		"extra_config": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A block to retrieve extra configuration key-value pairs",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The key of the extra configuration item",
					},
					"value": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The value of the extra configuration item",
					},
					"required": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the extra configuration item is required",
					},
				},
			},
		},
		"cpu_hot_add_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True if the virtual machine supports addition of virtual CPUs while powered on.",
		},
		"memory_hot_add_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "True if the virtual machine supports addition of memory while powered on.",
		},
		"sizing_policy_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "VM sizing policy ID.",
		},
		"placement_policy_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "VM placement policy ID.",
		},
		"security_tags": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Security tags assigned to this VM",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"status": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Shows the status code of the VM",
		},
		"status_text": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Shows the status of the VM",
		},
		"inherited_metadata": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "A map that contains metadata that is automatically added by VCD (10.5.1+) and provides details on the origin of the VM",
		},
	}
}

func datasourceVcdVAppVm() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVAppVmRead,
		Schema:      vcdVmDS(vappVmType),
	}
}

func datasourceVcdVAppVmRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmRead(d, meta, "datasource")
}
