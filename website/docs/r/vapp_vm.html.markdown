---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_vapp_vm"
sidebar_current: "docs-vcloud-resource-vapp-vm"
description: |-
  Provides a Viettel IDC Cloud VM resource. This can be used to create, modify, and delete VMs within a vApp.
---

# vcloud\_vapp\_vm

Provides a Viettel IDC Cloud VM resource. This can be used to create,
modify, and delete VMs within a vApp.

## Example Usage

```hcl
# System administrator rights are required to connect external network
resource "vcloud_network_direct" "direct-external" {
  name             = "net"
  external_network = "my-ext-net"
}

resource "vcloud_vapp" "web" {
  name = "web"
}

resource "vcloud_vapp_org_network" "routed-net" {
  vapp_name        = vcloud_vapp.web.name
  org_network_name = "my-vdc-int-net"
}

resource "vcloud_vapp_org_network" "direct-net" {
  vapp_name        = vcloud_vapp.web.name
  org_network_name = vcloud_network_direct.direct-external.name
}

resource "vcloud_vapp_network" "vapp-net" {
  name               = "my-vapp-net"
  vapp_name          = vcloud_vapp.web.name
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"
  dns1               = "192.168.2.1"
  dns2               = "192.168.2.2"
  dns_suffix         = "mybiz.biz"
  guest_vlan_allowed = true

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}

data "vcloud_catalog" "my-catalog" {
  org  = "test"
  name = "my-catalog"
}

data "vcloud_catalog_vapp_template" "photon-os" {
  org        = "test"
  catalog_id = data.vcloud_catalog.my-catalog.id
  name       = "photon-os"
}

resource "vcloud_vapp_vm" "web1" {
  vapp_name        = vcloud_vapp.web.name
  name             = "web1"
  vapp_template_id = data.vcloud_catalog_vapp_template.photon-os.id
  memory           = 1024
  cpus             = 2
  cpu_cores        = 1

  metadata_entry {
    key   = "role"
    value = "web"
  }

  metadata_entry {
    key   = "env"
    value = "staging"
  }

  metadata_entry {
    key   = "version"
    value = "v1"
  }

  metadata_entry {
    key   = "my_key"
    value = "my value"
  }

  guest_properties = {
    "guest.hostname"   = "my-host"
    "another.var.name" = "var-value"
  }

  network {
    type               = "org"
    name               = vcloud_vapp_org_network.direct-net.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }
}

resource "vcloud_independent_disk" "disk1" {
  name         = "logDisk"
  size         = "512"
  bus_type     = "SCSI"
  bus_sub_type = "VirtualSCSI"
}

resource "vcloud_vapp_vm" "web2" {
  vapp_name        = vcloud_vapp.web.name
  name             = "web2"
  vapp_template_id = data.vcloud_catalog_vapp_template.photon-os.id
  memory           = 1024
  cpus             = 1

  metadata_entry {
    key   = "role"
    value = "web"
  }

  metadata_entry {
    key   = "env"
    value = "staging"
  }

  metadata_entry {
    key   = "env"
    value = "staging"
  }

  metadata_entry {
    key   = "version"
    value = "v1"
  }

  metadata_entry {
    key   = "my_key"
    value = "my value"
  }

  guest_properties = {
    "guest.hostname" = "my-hostname"
    "guest.other"    = "another-setting"
  }

  network {
    type               = "org"
    name               = vcloud_vapp_org_network.routed-net.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }

  network {
    type               = "vapp"
    name               = vcloud_vapp_network.vapp-net.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "none"
    ip_allocation_mode = "NONE"
    connected          = false
  }

  disk {
    name        = vcloud_independent_disk.disk1.name
    bus_number  = 1
    unit_number = 0
  }

}
```

## Example Usage (Override Template Disk)
This example shows how to [change VM template's disk properties](#override-template-disk) when the VM is created.

```hcl
data "vcloud_catalog" "boxes" {
  org  = "test"
  name = "Boxes"
}

data "vcloud_catalog_vapp_template" "lampstack" {
  org        = "test"
  catalog_id = data.vcloud_catalog.boxes.id
  name       = "lampstack-1.10.1-ubuntu-10.04"
}

resource "vcloud_vapp_vm" "internalDiskOverride" {
  vapp_name        = vcloud_vapp.web.name
  name             = "internalDiskOverride"
  vapp_template_id = data.vcloud_catalog_vapp_template.lampstack.id
  memory           = 2048
  cpus             = 2
  cpu_cores        = 1

  # Fast provisioned VDCs require disks to be consolidated
  # if their size is to be changed
  # consolidate_disks_on_create = true 

  override_template_disk {
    bus_type        = "paravirtual"
    size_in_mb      = "22384"
    bus_number      = 0
    unit_number     = 0
    iops            = 0
    storage_profile = "*"
  }
}

```

## Example Usage (Wait for IP addresses on DHCP NIC)
This example shows how to use [`network_dhcp_wait_seconds`](#network_dhcp_wait_seconds) with DHCP.

```hcl
data "vcloud_catalog" "cat-dserplis" {
  org  = "test"
  name = "cat-dserplis"
}

data "vcloud_catalog_vapp_template" "photon-rev2" {
  org        = "test"
  catalog_id = data.vcloud_catalog.cat-dserplis.id
  name       = "photon-rev2"
}

resource "vcloud_vapp_vm" "TestAccVcdVAppVmDhcpWaitVM" {
  vapp_name        = vcloud_vapp.TestAccVcdVAppVmDhcpWait.name
  name             = "brr"
  computer_name    = "dhcp-vm"
  vapp_template_id = data.vcloud_catalog_vapp_template.photon-rev2.id
  memory           = 512
  cpus             = 2
  cpu_cores        = 1

  network_dhcp_wait_seconds = 300 # 5 minutes
  network {
    type               = "org"
    name               = vcloud_network_routed.net.name
    ip_allocation_mode = "DHCP"
    is_primary         = true
  }
}

resource "vcloud_nsxv_ip_set" "test-ipset" {
  name         = "ipset-with-dhcp-ip"
  ip_addresses = [vcloud_vapp_vm.TestAccVcdVAppVmDhcpWaitVM.network.0.ip]
}
```

## Example Usage (Empty VM)
This example shows how to create an empty VM.

```hcl
data "vcloud_catalog" "my-catalog" {
  org  = "test"
  name = "my-catalog"
}

data "vcloud_catalog_media" "myMedia" {
  org     = "test"
  catalog = data.vcloud_catalog.my-catalog.name
  name    = "myMedia"
}

resource "vcloud_vapp_vm" "emptyVM" {
  vapp_name     = vcloud_vapp.web.name
  name          = "VmWithoutTemplate"
  computer_name = "emptyVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  boot_image_id    = data.vcloud_catalog_media.myMedia.id
}

```

## Example Usage (vApp template with multi VMs)
This example shows how to create a VM from a vApp template with multiple VMs by specifying which VM to use.

```hcl
data "vcloud_catalog" "cat-where-is-template" {
  org  = "test"
  name = "cat-where-is-template"
}

data "vcloud_catalog_vapp_template" "vappWithMultiVm" {
  org        = "test"
  catalog_id = data.vcloud_catalog.cat-where-is-template.id
  name       = "vappWithMultiVm"
}

resource "vcloud_vapp_vm" "secondVM" {
  vapp_name           = vcloud_vapp.web.name
  name                = "secondVM"
  computer_name       = "db-vm"
  vapp_template_id    = data.vcloud_catalog_vapp_template.vappWithMultiVm.id
  vm_name_in_template = "secondVM" # Specifies which VM to use from the template
  memory              = 512
  cpus                = 2
  cpu_cores           = 1
}

```

## Example Usage (VM with sizing policy)
This example shows how to create a VM using VM sizing policy.

```hcl
data "vcloud_catalog" "cat-where-is-template" {
  org  = "test"
  name = "cat-where-is-template"
}

data "vcloud_catalog_vapp_template" "vappWithMultiVm" {
  org        = "test"
  catalog_id = data.vcloud_catalog.cat-where-is-template.id
  name       = "vappWithMultiVm"
}

data "vcloud_vm_sizing_policy" "minSize" {
  name = "minimum size"
}

resource "vcloud_vapp_vm" "secondVM" {
  vapp_name        = vcloud_vapp.web.name
  name             = "secondVM"
  computer_name    = "db-vm"
  vapp_template_id = data.vcloud_catalog_vapp_template.vappWithMultiVm.id
  sizing_policy_id = data.vcloud_vm_sizing_policy.minSize.id # Specifies which sizing policy to use
}

```

## Example Usage (VM with sizing policy and VM placement policy)
This example shows how to create a VM using a VM sizing policy and a VM placement policy.

```hcl
data "vcloud_catalog" "cat-where-is-template" {
  org  = "test"
  name = "cat-where-is-template"
}

data "vcloud_catalog_vapp_template" "vappWithMultiVm" {
  org        = "test"
  catalog_id = data.vcloud_catalog.cat-where-is-template.id
  name       = "vappWithMultiVm"
}

data "vcloud_vm_sizing_policy" "minSize" {
  name = "minimum size"
}

data "vcloud_provider_vdc" "myPvdc" {
  name = "nsxt-Pvdc"
}

data "vcloud_vm_placement_policy" "placementPolicy" {
  name            = "vmware"
  provider_vdc_id = data.vcloud_provider_vdc.myPvdc.id
}

resource "vcloud_vapp_vm" "secondVM" {
  vapp_name           = vcloud_vapp.web.name
  name                = "secondVM"
  computer_name       = "db-vm"
  vapp_template_id    = data.vcloud_catalog_vapp_template.vappWithMultiVm.id
  sizing_policy_id    = data.vcloud_vm_sizing_policy.minSize.id # Specifies which sizing policy to use
  placement_policy_id = data.vcloud_vm_placement_policy.placementPolicy.id
}
```

## Example Usage (VM with sizing policy and VM placement policy)
This example shows how to create a VM using a VM sizing policy and a VM placement policy.

```hcl
data "vcloud_vm_sizing_policy" "minSize" {
  name = "minimum size"
}

data "vcloud_provider_vdc" "myPvdc" {
  name = "nsxt-Pvdc"
}

data "vcloud_vm_placement_policy" "placementPolicy" {
  name            = "vmware"
  provider_vdc_id = data.vcloud_provider_vdc.myPvdc.id
}

resource "vcloud_vapp_vm" "secondVM" {
  vapp_name           = vcloud_vapp.web.name
  name                = "secondVM"
  computer_name       = "db-vm"
  catalog_name        = "cat-where-is-template"
  template_name       = "vappWithMultiVm"
  sizing_policy_id    = data.vcloud_vm_sizing_policy.minSize.id # Specifies which sizing policy to use
  placement_policy_id = data.vcloud_vm_placement_policy.placementPolicy.id
}
```

## Example Usage (using advanced compute settings)
This example shows how to create an empty VM with advanced compute settings.

```hcl
data "vcloud_catalog" "my-catalog" {
  org  = "test"
  name = "my-catalog"
}

data "vcloud_catalog_media" "myMedia" {
  org     = "test"
  catalog = data.vcloud_catalog.my-catalog.name
  name    = "myMedia"
}

resource "vcloud_vapp_vm" "advancedVM" {
  vapp_name     = vcloud_vapp.web.name
  name          = "advancedVM"
  computer_name = "advancedVM"
  memory        = 2048
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  boot_image_id    = data.vcloud_catalog_media.myMedia.id

  memory_priority    = "CUSTOM"
  memory_shares      = "480"
  memory_reservation = "8"
  memory_limit       = "48"

  cpu_priority    = "CUSTOM"
  cpu_shares      = "512"
  cpu_reservation = "200"
  cpu_limit       = "1000"
}
```

## Example Usage (VM copy)
This example shows how to create a copy of an existing VM

```hcl
data "vcloud_vapp_vm" "existing" {
  vapp_name = data.vcloud_vapp.web.name
  name      = "web1"
}

data "vcloud_vapp_org_network" "net1" {
  vapp_name        = web1
  org_network_name = "my-vapp-org-network"
}

resource "vcloud_vapp_vm" "vm-copy" {
  org = "org"
  vdc = "vdc"

  copy_from_vm_id = data.vcloud_vapp_vm.existing.id # source VM ID
  vapp_name       = data.vcloud_vapp_vm.existing.vapp_name
  name            = "VM Copy"
  power_on        = false

  network {
    type               = "org"
    name               = data.vcloud_vapp_org_network.net1.org_network_name
    adapter_type       = "VMXNET3"
    ip_allocation_mode = "POOL"
  }

  prevent_update_power_off = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `vapp_name` - (Required) The vApp this VM belongs to.
* `name` - (Required) A name for the VM, unique within the vApp 
* `computer_name` - (Optional; *v2.5+*) Computer name to assign to this virtual machine.
* `vapp_template_id` - (Optional; *v3.8+*) The URN of the vApp Template to use. You can fetch it using a [`vcloud_catalog_vapp_template`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/catalog_vapp_template) data source.
* `vm_name_in_template` - (Optional; *v2.9+*) The name of the VM in vApp Template to use. For cases when vApp template has more than one VM.
* `copy_from_vm_id` - (Optional; *v3.12+*) The ID of *an existing VM* to make a copy of it (it
  cannot be a vApp template). The source VM *must be in the same Org* (but can be in different VDC).
  *Note:* `sizing_policy_id` must be specified when creating a standalone VM (using `vcloud_vm`
  resource) and using different source/destination VDCs.
* `memory` - (Optional) The amount of RAM (in MB) to allocate to the VM. If `memory_hot_add_enabled` is true, then memory will be increased without VM power off
* `memory_reservation` - The amount of RAM (in MB) reservation on the underlying virtualization infrastructure
* `memory_priority` - Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload
* `memory_shares` - Custom priority for the resource in MB. This is a read-only, unless the `memory_priority` is "CUSTOM"
* `memory_limit` - The limit (in MB) for how much of memory can be consumed on the underlying virtualization infrastructure. `-1` value for unlimited.
* `cpus` - (Optional) The number of virtual CPUs to allocate to the VM. Socket count is a result of: virtual logical processors/cores per socket. If `cpu_hot_add_enabled` is true, then cpus will be increased without VM power off.
* `cpu_cores` - (Optional; *v2.1+*) The number of cores per socket.
* `cpu_reservation` - The amount of MHz reservation on the underlying virtualization infrastructure.
* `cpu_priority` - Pre-determined relative priorities according to which the non-reserved portion of this resource is made available to the virtualized workload
* `cpu_shares` - Custom priority for the resource in MHz. This is a read-only, unless the `cpu_priority` is "CUSTOM"
* `cpu_limit` - The limit (in MHz) for how much of CPU can be consumed on the underlying virtualization infrastructure. `-1` value for unlimited. 
* `metadata` - (Deprecated; *v2.2+*) Use `metadata_entry` instead. Key value map of metadata to assign to this VM
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `storage_profile` (Optional; *v2.6+*) Storage profile to override the default one
* `power_on` - (Optional) A boolean value stating if this VM should be powered on. Default is `true`
* `accept_all_eulas` - (Optional; *v2.0+*) Automatically accept EULA if OVA has it. Default is `true`
* `disk` - (Optional; *v2.1+*) Independent disk attachment configuration. See [Disk](#disk) below for details.
* `expose_hardware_virtualization` - (Optional; *v2.2+*) Boolean for exposing full CPU virtualization to the
guest operating system so that applications that require hardware virtualization can run on virtual machines without binary
translation or paravirtualization. Useful for hypervisor nesting provided underlying hardware supports it. Default is `false`.
* `network` - (Optional; *v2.2+*) A block to define network interface. Multiple can be used. See [Network](#network-block) and 
example for usage details.
* `customization` - (Optional; *v2.5+*) A block to define for guest customization options. See [Customization](#customization-block)
* `guest_properties` - (Optional; *v2.5+*) Key value map of guest properties
* `description`  - (Optional; *v2.9+*) The VM description. Note: for VM from Template `description` is read only. Currently, this field has
  the description of the OVA used to create the VM.
* `override_template_disk` - (Optional; *v2.7+*) Allows to update internal disk in template before first VM boot. Disk is matched by `bus_type`, `bus_number` and `unit_number`. See [Override template Disk](#override-template-disk) below for details.
* `consolidate_disks_on_create` - (Optional; *3.12+*) Performs disk consolidation during creation.
  The main use case is when one wants to grow template disk size using `override_template_disk` in
  fast provisioned VDCs. **Note:** Consolidating disks requires right `vApp: VM Migrate, Force
  Undeploy, Relocate, Consolidate`. This operation _may take long time_ depending on disk size and
  storage performance.
* `network_dhcp_wait_seconds` - (Optional; *v2.7+*) Optional number of seconds to try and wait for DHCP IP (only valid
  for adapters in `network` block with `ip_allocation_mode=DHCP`). It constantly checks if IP is present so the time given
  is a maximum. VM must be powered on and _at least one_ of the following _must be true_:
* VM has Guest Tools. It waits for IP address to be reported by Guest Tools. This is a slower option, but
  does not require for the VM to use Edge Gateways DHCP service.
* VM DHCP interface is connected to routed Org network and is using Edge Gateways DHCP service (not
  relayed). It works by querying DHCP leases on Edge Gateway. In general it is quicker than waiting
  until Guest Tools report IP addresses, but is more constrained. However this is the only option if Guest
  Tools are not present on the VM.
* `os_type` - (Optional; *v2.9+*) Operating System type. Possible values can be found in [Os Types](#os-types). Required when creating empty VM.
* `hardware_version` - (Optional; *v2.9+*) Virtual Hardware Version (e.g.`vmx-14`, `vmx-13`, `vmx-12`, etc.). Required when creating empty VM.
* `firmware` - (Optional; v3.11+, Vcloud 10.4.1+) Specify boot firmware of the VM. Can be `efi` or `bios`. If unset, defaults to `bios`. Changing the value requires the VM to power off.
* `boot_options` - (Optional; v3.11+) A block to define boot options of the VM. See [Boot Options](#boot-options)
* `boot_image_id` - (Optional; *v3.8+*) Media URN to mount as boot image. You can fetch it using a [`vcloud_catalog_media`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/catalog_media) data source.
  Image is mounted only during VM creation. On update if value is changed to empty it will eject the mounted media. If you want to mount an image later, please use [vcloud_inserted_media](/providers/terraform-viettelidc/vcloud/latest/docs/resources/inserted_media). 
* `cpu_hot_add_enabled` - (Optional; *v3.0+*) True if the virtual machine supports addition of virtual CPUs while powered on. Default is `false`.
* `memory_hot_add_enabled` - (Optional; *v3.0+*) True if the virtual machine supports addition of memory while powered on. Default is `false`.
* `prevent_update_power_off` - (Optional; *v3.0+*) True if the update of resource should fail when virtual machine power off needed. Default is `false`.
* `sizing_policy_id` (Optional; *v3.0+*, *vCD 10.0+*) VM sizing policy ID. To be used, it needs to be assigned to [Org VDC](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org_vdc)
  using `vcloud_org_vdc.vm_sizing_policy_ids` (and `vcloud_org_vdc.default_compute_policy_id` to make it default).
  In this case, if the sizing policy is not set, it will pick the VDC default on creation. It must be set explicitly
  if one wants to update it to another policy (the VM requires at least one Compute Policy), and needs to be set to `""` to be removed.
* `placement_policy_id` (Optional; *v3.8+*) VM placement policy or [vGPU policy][vgpu-policy] (*3.11+*) ID. To be used, it needs to be assigned to [Org VDC](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org_vdc)
  In this case, if the placement policy is not set, it will pick the VDC default on creation. It must be set explicitly
  if one wants to update it to another policy (the VM requires at least one Compute Policy), and needs to be set to `""` to be removed.
* `security_tags` - (Optional; *v3.9+*) Set of security tags to be managed by the `vcloud_vapp_vm` resource.
  To remove `security_tags` you must set `security_tags = []` and do not remove the attribute. Removing the attribute will cause the tags to remain unchanged and just stop being managed by this resource.
  This is to be consistent with existing security tags that were created by the `vcloud_security_tags` resource.
* `set_extra_config` - (Optional; *v3.13+*) Set of extra configuration key/values to be added or modified. See [Extra Configuration](#extra-configuration)

~> **Note:** Only one of `security_tags` attribute or [`vcloud_security_tag`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/security_tag) resource
  should be used. Using both would cause a behavioral conflict.

* `catalog_name` - (Deprecated; *v2.9+*) Use a [`vcloud_catalog`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/catalog) data source along with `vapp_template_id` or `boot_image_id` instead. The catalog name in which to find the given vApp Template or media for `boot_image`.
* `template_name` - (Deprecated; *v2.9+*) Use `vapp_template_id` instead. The name of the vApp Template to use
* `boot_image` - (Deprecated; *v2.9+*) Use `boot_image_id` instead. Media name to mount as boot image. Image is mounted only during VM creation. On update if value is changed to empty it will eject the mounted media. If you want to mount an image later, please use [vcloud_inserted_media](/providers/terraform-viettelidc/vcloud/latest/docs/resources/inserted_media).

## Attribute reference

* `vm_type` - (*3.2+*) Type of the VM (either `vcloud_vapp_vm` or `vcloud_vm`).
* `status` - (*v3.8+*) The vApp status as a numeric code.
* `status_text` - (*v3.8+*) The vApp status as text.
* `inherited_metadata` - (*v3.11+*; *Vcloud 10.5.1+*) A map that contains read-only metadata that is automatically added by Vcloud (10.5.1+) and provides
  details on the origin of the VM (e.g. `vm.origin.id`, `vm.origin.name`, `vm.origin.type`).
* `extra_config` - (*v3.13.+*) The VM extra configuration. See [Extra Configuration](#extra-configuration) for more detail.

<a id="disk"></a>
## Disk

* `name` - (Required) Independent disk name
* `bus_number` - (Required) Bus number on which to place the disk controller
* `unit_number` - (Required) Unit number (slot) on the bus specified by BusNumber.

<a id="network-block"></a>
## Network

* `type` (Required) Network type, one of: `none`, `vapp` or `org`. `none` creates a NIC with no network attached. `vapp` requires `name` of existing vApp network (created with `vcloud_vapp_network`). `org` requires attached vApp Org network `name` (attached with `vcloud_vapp_org_network`).
* `name` (Optional) Name of the network this VM should connect to. Always required except for `type` `NONE`. 
* `is_primary` (Optional) Set to true if network interface should be primary. First network card in the list will be primary by default.
* `mac` - (Computed) Mac address of network interface.
* `adapter_type` - (Optional, Computed) Adapter type (names are case insensitive). Some known adapter types - `VMXNET3`,
    `E1000`, `E1000E`, `SRIOVETHERNETCARD`, `VMXNET2`, `PCNet32`.

    **Note:** Adapter type change for existing NIC will return an error during `apply` operation because vCD does not
    support changing adapter type for existing resource.

    **Note:** Adapter with type `SRIOVETHERNETCARD` **must** be connected to a **direct** vApp
    network connected to a direct VDC network. Unless such an SR-IOV-capable external network is
    available in your VDC, you cannot connect an SR-IOV device.

* `ip_allocation_mode` (Required) IP address allocation mode. One of `POOL`, `DHCP`, `MANUAL`, `NONE`:  

  * `POOL` - Static IP address is allocated automatically from defined static pool in network.
  
  * `DHCP` - IP address is obtained from a DHCP service. Field `ip` is not guaranteed to be populated. Because of this it may appear
  after multiple `terraform refresh` operations.  **Note.**
    [`network_dhcp_wait_seconds`](#network_dhcp_wait_seconds) parameter can help to ensure IP is
    reported on first run.

  
  * `MANUAL` - IP address is assigned manually in the `ip` field. Must be valid IP address from static pool.
  
  * `NONE` - No IP address will be set because VM will have a NIC without network.

* `ip` (Optional, Computed) Settings depend on `ip_allocation_mode`. Field requirements for each `ip_allocation_mode` are listed below:

  * `ip_allocation_mode=POOL` - **`ip`** value must be omitted or empty string "". Empty string may be useful when doing HCL
  variable interpolation. Field `ip` will be populated with an assigned IP from static pool after run.
  
  * `ip_allocation_mode=DHCP` - **`ip`** value must be omitted or empty string "". Field `ip` is not
    guaranteed to be populated after run due to the VM lacking VMware tools or not working properly
    with DHCP. Because of this `ip` may also appear after multiple `terraform refresh` operations
    when is reported back to vCD. **Note.**
    [`network_dhcp_wait_seconds`](#network_dhcp_wait_seconds) parameter can help to ensure IP is
    reported on first run.

  * `ip_allocation_mode=MANUAL` - **`ip`** value must be valid IP address from a subnet defined in `static pool` for network.

  * `ip_allocation_mode=NONE` - **`ip`** field can be omitted or set to an empty string "". Empty string may be useful when doing HCL variable interpolation.
  
  * `connected` - (Optional; *v3.0+*) It defines if NIC is connected or not. Network with `ip_allocation_mode=NONE` can't be connected by default, please use `connected=false` in such case.   

<a id="override-template-disk"></a>
## Override template disk
Allows to update internal disk in template before first VM boot. Disk is matched by `bus_type`, `bus_number` and `unit_number`.
Changes are ignored on update. This part isn't reread on refresh. To manage internal disk later please use [`vcloud_vm_internal_disk`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/vm_internal_disk) resource.
 
~> **Note:** Managing disks in VM with fast provisioned VDC require
[`consolidate_disks_on_create`](#consolidate_disks_on_create) option.

* `bus_type` - (Required) The type of disk controller. Possible values: `ide`, `parallel`( LSI Logic Parallel SCSI),
  `sas`(LSI Logic SAS (SCSI)), `paravirtual`(Paravirtual (SCSI)), `sata`, `nvme`. **Note** `nvme` requires *v3.5.0+* and
  Vcloud *10.2.1+*
* `size_in_mb` - (Required) The size of the disk in MB. 
* `bus_number` - (Required) The number of the SCSI or IDE controller itself.
* `unit_number` - (Required) The device number on the SCSI or IDE controller of the disk.
* `iops` - (Optional) Specifies the IOPS for the disk. Default is 0.
* `storage_profile` - (Optional) Storage profile which overrides the VM default one.

<a id="boot-options"></a>
## Boot options

Allows to specify the boot options of a VM.

* `efi_secure_boot` - (Optional, Vcloud 10.4.1+) Enable EFI Secure Boot on subsequent boots, requires `firmware` to be set to `efi`.
* `enter_bios_setup_on_next_boot` - (Optional) Enter BIOS setup on the next boot of the VM. After a VM is booted, the value is set back to false in Vcloud, because of that, 
  Terraform will return an inconsistent plan and try to set this field back to `true`. **NOTE:** If there are any [cold changes](#hot-and-cold-update) on update that cause the VM to power-cycle with this field set to `true`,
  the VM will boot straight into BIOS. For reducing side effects, one should set this field to `true` and `power_on` to `false`, then switch `power_on` to `true`.
* `boot_delay` - (Optional) Delay between the power-on and boot of the VM in milliseconds.
* `boot_retry_enabled` - (Optional, Vcloud 10.4.1+) If set to `true`, will attempt to reboot the VM after a failed boot.
* `boot_retry_delay` - (Optional, Vcloud 10.4.1+) Delay before the VM is rebooted after a failed boot. Has no effect if `boot_retry_enabled` is set to `false`

<a id="customization-block"></a>
## Customization

When you customize your guest OS you can set up a virtual machine with the operating system that you want.

Viettel IDC Cloud can customize the network settings of the guest operating system of a virtual machine created from a
vApp template. When you customize your guest operating system, you can create and deploy multiple unique virtual
machines based on the same vApp template without machine name or network conflicts.

When you configure a vApp template with the prerequisites for guest customization and add a virtual machine to a vApp
based on that template, Viettel IDC Cloud creates a package with guest customization tools. When you deploy and power on
the virtual machine for the first time, Viettel IDC Cloud copies the package, runs the tools, and deletes the package from
the virtual machine.

~> **Note:** The settings below work so that all values are inherited from template and only the specified fields are
overridden with exception being `force` field which works like a flag.

* `force` (Optional) **Warning.** `true` value will cause the VM to reboot on every `apply` operation.
This field works as a flag and triggers force customization when `true` during an update 
(`terraform apply`) every time. It never complains about a change in statefile. Can be used when guest customization
is needed after VM configuration (e.g. NIC change, customization options change, etc.) and then set back to `false`.
**Note.** It will not have effect when `power_on` field is set to `false`. See [example workflow below](#example-forced-customization-workflow).
* `enabled` (Optional; *v2.7+*) `true` will enable guest customization which may occur on first boot or if the `force` flag is used.
This option should be selected for **Power on and Force re-customization to work**.
* `change_sid` (Optional; *v2.7+*) Allows to change SID (security identifier). Only applicable for Windows operating systems.
* `allow_local_admin_password` (Optional; *v2.7+*) Allow local administrator password.
* `must_change_password_on_first_login` (Optional; *v2.7+*) Require Administrator to change password on first login.
* `auto_generate_password` (Optional; *v2.7+*) Auto generate password. **Note:**
  `allow_local_admin_password` must be enabled, otherwise next plan will be inconsistent and report
  `auto_generate_password=false`
* `admin_password` (Optional; *v2.7+*) Manually specify Administrator password.
* `number_of_auto_logons` (Optional; *v2.7+*) Number of times to log on automatically. `0` means disabled.
* `join_domain` (Optional; *v2.7+*) Enable this VM to join a domain.
* `join_org_domain` (Optional; *v2.7+*) Set to `true` to use organization's domain.
* `join_domain_name` (Optional; *v2.7+*) Set the domain name to override organization's domain name.
* `join_domain_user` (Optional; *v2.7+*) User to be used for domain join.
* `join_domain_password` (Optional; *v2.7+*) Password to be used for domain join.
* `join_domain_account_ou` (Optional; *v2.7+*) Organizational unit to be used for domain join.
* `initscript` (Optional; *v2.7+*) Provide initscript to be executed when customization is applied.

## Example of a Forced Customization Workflow

Step 1 - Setup VM:

```hcl
data "vcloud_catalog" "boxes" {
  org  = "test"
  name = "Boxes"
}

data "vcloud_catalog_vapp_template" "windows" {
  org        = "test"
  catalog_id = data.vcloud_catalog.boxes.id
  name       = "windows"
}

resource "vcloud_vapp_vm" "web2" {
  vapp_name        = vcloud_vapp.web.name
  name             = "web2"
  vapp_template_id = data.vcloud_catalog_vapp_template.windows.id
  memory           = 2048
  cpus             = 1

  network {
    type               = "org"
    name               = "net"
    ip                 = "10.10.104.162"
    ip_allocation_mode = "MANUAL"
  }
}
```

Step 2 - Override some VM customization options and force customization (VM will be rebooted during `terraform apply`):

```hcl
resource "vcloud_vapp_vm" "web2" {
  # ...

  network {
    type               = "org"
    name               = "net"
    ip_allocation_mode = "DHCP"
  }

  customization {
    force                      = true
    change_sid                 = true
    allow_local_admin_password = true
    auto_generate_password     = false
    admin_password             = "my-secure-password"
    # Other customization options to override the ones from template
  }
}
```

Step 3 - Once customization is done, set the force customization flag to false (or remove it) to prevent forcing
customization on every `terraform apply` command:

```hcl
resource "vcloud_vapp_vm" "web2" {
  # ...

  network {
    type               = "org"
    name               = "net"
    ip_allocation_mode = "DHCP"
  }

  customization {
    force                      = false
    change_sid                 = true
    allow_local_admin_password = true
    auto_generate_password     = false
    admin_password             = "my-secure-password"
    # Other customization options to override the ones from template
  }
}
```
<a id="os-types"></a>
## Os Types
* **Linux**
  - `other5xLinux64Guest` - Other 5.x or later Linux (64-bit)
  - `other5xLinuxGuest` - Other 5.x or later Linux (32-bit)
  - `other4xLinux64Guest` - Other 4.x Linux (64-bit)
  - `other4xLinuxGuest` - Other 4.x Linux (32-bit)
  - `other3xLinux64Guest` - Other 3.x Linux (64-bit)
  - `other3xLinuxGuest` - Other 3.x Linux (32-bit)
  - `other26xLinux64Guest` - Other 2.6.x Linux (64-bit)
  - `other26xLinuxGuest` - Other 2.6.x Linux (32-bit)
  - `other24xLinux64Guest` - Other 2.4.x Linux (64-bit)
  - `other24xLinuxGuest` - Other 2.4.x Linux (32-bit)
  - `otherLinux64Guest` - Other Linux (64-bit)
  - `otherLinuxGuest` - Other Linux (32-bit)
  - `sles16_64Guest` - SUSE Linux Enterprise 16 (64-bit)
  - `sles15_64Guest` - SUSE Linux Enterprise 15 (64-bit)
  - `sles12_64Guest` - SUSE Linux Enterprise 12 (64-bit)
  - `sles11_64Guest` - SUSE Linux Enterprise 11 (64-bit)
  - `sles11Guest` - SUSE Linux Enterprise 11 (32-bit)
  - `sles10_64Guest` - SUSE Linux Enterprise 10 (64-bit)
  - `sles10Guest` - SUSE Linux Enterprise 10 (32-bit)
  - `sles64Guest` - SUSE Linux Enterprise 8/9 (64-bit)
  - `slesGuest` - SUSE Linux Enterprise 8/9 (32-bit)
  - `rhel9_64Guest` - Red Hat Enterprise Linux 9 (64-bit)
  - `rhel8_64Guest` - Red Hat Enterprise Linux 8 (64-bit)
  - `rhel7_64Guest` - Red Hat Enterprise Linux 7 (64-bit)
  - `rhel6_64Guest` - Red Hat Enterprise Linux 6 (64-bit)
  - `rhel6Guest` - Red Hat Enterprise Linux 6 (32-bit)
  - `rhel5_64Guest` - Red Hat Enterprise Linux 5 (64-bit)
  - `rhel5Guest` - Red Hat Enterprise Linux 5 (32-bit)
  - `rhel4_64Guest` - Red Hat Enterprise Linux 4 (64-bit)
  - `rhel4Guest` - Red Hat Enterprise Linux 4 (32-bit)
  - `rhel3_64Guest` - Red Hat Enterprise Linux 3 (64-bit)
  - `rhel3Guest` - Red Hat Enterprise Linux 3 (32-bit)
  - `rhel2Guest` - Red Hat Enterprise Linux 2.1
  - `oracleLinux9_64Guest` - Oracle Linux 9 (64-bit)
  - `oracleLinux8_64Guest` - Oracle Linux 8 (64-bit)
  - `oracleLinux7_64Guest` - Oracle Linux 7 (64-bit)
  - `oracleLinux6_64Guest` - Oracle Linux 6 (64-bit)
  - `oracleLinux6Guest` - Oracle Linux 6 (32-bit)
  - `oracleLinux64Guest` - Oracle Linux 4/5 (64-bit)
  - `oracleLinuxGuest` - Oracle Linux 4/5 (32-bit)
  - `centos9_64Guest` - CentOS 9 (64-bit)
  - `centos8_64Guest` - CentOS 8 (64-bit)
  - `centos7_64Guest` - CentOS 7 (64-bit)
  - `centos6_64Guest` - CentOS 6 (64-bit)
  - `centos6Guest` - CentOS 6 (32-bit)
  - `centos64Guest` - CentOS 4/5 (64-bit)
  - `centosGuest` - CentOS 4/5 (32-bit)
  - `asianux9_64Guest` - Asianux 9 (64-bit)
  - `asianux8_64Guest` - Asianux 8 (64-bit)
  - `asianux7_64Guest` - Asianux 7 (64-bit)
  - `asianux4_64Guest` - Asianux 4 (64-bit)
  - `asianux4Guest` - Asianux 4 (32-bit)
  - `asianux3_64Guest` - Asianux 3 (64-bit)
  - `asianux3Guest` - Asianux 3 (32-bit)
  - `amazonlinux3_64Guest` - Amazon Linux 3 (64-bit)
  - `amazonlinux2_64Guest` - Amazon Linux 2 (64-bit)
  - `debian11_64Guest` - Debian GNU/Linux 11 (64-bit)
  - `debian11Guest` - Debian GNU/Linux 11 (32-bit)
  - `debian10_64Guest` - Debian GNU/Linux 10 (64-bit)
  - `debian10Guest` - Debian GNU/Linux 10 (32-bit)
  - `debian9_64Guest` - Debian GNU/Linux 9 (64-bit)
  - `debian9Guest` - Debian GNU/Linux 9 (32-bit)
  - `debian8_64Guest` - Debian GNU/Linux 8 (64-bit)
  - `debian8Guest` - Debian GNU/Linux 8 (32-bit)
  - `debian7_64Guest` - Debian GNU/Linux 7 (64-bit)
  - `debian7Guest` - Debian GNU/Linux 7 (32-bit)
  - `debian6_64Guest` - Debian GNU/Linux 6 (64-bit)
  - `debian6Guest` - Debian GNU/Linux 6 (32-bit)
  - `debian5_64Guest` - Debian GNU/Linux 5 (64-bit)
  - `debian5Guest` - Debian GNU/Linux 5 (32-bit)
  - `debian4_64Guest` - Debian GNU/Linux 4 (64-bit)
  - `debian4Guest` - Debian GNU/Linux 4 (32-bit)
  - `crxPod1Guest` - VMware CRX Pod 1 (64-bit)
  - `vmwarePhoton64Guest` - VMware Photon OS (64-bit)
  - `opensuse64Guest` - SUSE openSUSE (64-bit)
  - `opensuseGuest` - SUSE openSUSE (32-bit)
  - `fedora64Guest` - Red Hat Fedora (64-bit)
  - `fedoraGuest` - Red Hat Fedora (32-bit)
  - `ubuntu64Guest` - Ubuntu Linux (64-bit)
  - `ubuntuGuest` - Ubuntu Linux (32-bit)
  - `coreos64Guest` - CoreOS Linux (64-bit)
  - `oesGuest` - Novell Open Enterprise Server
* **Windows**
  - `windows2019srvNext_64Guest` - Microsoft Windows Server 2022 (64-bit)
  - `windows2019srv_64Guest` - Microsoft Windows Server 2019 (64-bit)
  - `windows9Server64Guest` - Microsoft Windows Server 2016 (64-bit)
  - `windows9_64Guest` - Microsoft Windows 10 (64-bit)
  - `windows9Guest` - Microsoft Windows 10 (32-bit)
  - `windows8Server64Guest` - Microsoft Windows Server 2012 (64-bit)
  - `windows8_64Guest` - Microsoft Windows 8.x (64-bit)
  - `windows8Guest` - Microsoft Windows 8.x (32-bit)
  - `win98Guest` - Microsoft Windows 98
  - `win95Guest` - Microsoft Windows 95
  - `win31Guest` - Microsoft Windows 3.1
  - `dosGuest` - Microsoft MS-DOS
  - `winXPPro64Guest` - Microsoft Windows XP Professional (64-bit)
  - `winXPProGuest` - Microsoft Windows XP Professional (32-bit)
  - `winVista64Guest` - Microsoft Windows Vista (64-bit)
  - `winVistaGuest` - Microsoft Windows Vista (32-bit)
  - `windows7Server64Guest` - Microsoft Windows Server 2008 R2 (64-bit)
  - `winLonghorn64Guest` - Microsoft Windows Server 2008 (64-bit)
  - `winLonghornGuest` - Microsoft Windows Server 2008 (32-bit)
  - `winNetWebGuest` - Microsoft Windows Server 2003 Web Edition (32-bit)
  - `winNetStandard64Guest` - Microsoft Windows Server 2003 Standard (64-bit)
  - `winNetStandardGuest` - Microsoft Windows Server 2003 Standard (32-bit)
  - `winNetDatacenter64Guest` - Microsoft Windows Server 2003 Datacenter (64-bit)
  - `winNetDatacenterGuest` - Microsoft Windows Server 2003 Datacenter (32-bit)
  - `winNetEnterprise64Guest` - Microsoft Windows Server 2003 (64-bit)
  - `winNetEnterpriseGuest` - Microsoft Windows Server 2003 (32-bit)
  - `winNTGuest` - Microsoft Windows NT
  - `windows7_64Guest` - Microsoft Windows 7 (64-bit)
  - `windows7Guest` - Microsoft Windows 7 (32-bit)
  - `win2000ServGuest` - Microsoft Windows 2000 Server
  - `win2000ProGuest` - Microsoft Windows 2000 Professional
  - `win2000AdvServGuest` - Microsoft Windows 2000
  - `winNetBusinessGuest` - Microsoft Small Business Server 2003
* **Other**:
  - `freebsd13_64Guest` - FreeBSD 13 or later versions (64-bit)
  - `freebsd13Guest` - FreeBSD 13 or later versions (32-bit)
  - `freebsd12_64Guest` - FreeBSD 12 (64-bit)
  - `freebsd12Guest` - FreeBSD 12 (32-bit)
  - `freebsd11_64Guest` - FreeBSD 11 (64-bit)
  - `freebsd11Guest` - FreeBSD 11 (32-bit)
  - `freebsd64Guest` - FreeBSD Pre-11 versions (64-bit)
  - `freebsdGuest` - FreeBSD Pre-11 versions (32-bit)
  - `darwin21_64Guest` - Apple macOS 12 (64-bit)
  - `darwin20_64Guest` - Apple macOS 11 (64-bit)
  - `darwin19_64Guest` - Apple macOS 10.15 (64-bit)
  - `darwin18_64Guest` - Apple macOS 10.14 (64-bit)
  - `darwin17_64Guest` - Apple macOS 10.13 (64-bit)
  - `darwin16_64Guest` - Apple macOS 10.12 (64-bit)
  - `darwin15_64Guest` - Apple Mac OS X 10.11 (64-bit)
  - `darwin14_64Guest` - Apple Mac OS X 10.10 (64-bit)
  - `darwin13_64Guest` - Apple Mac OS X 10.9 (64-bit)
  - `darwin12_64Guest` - Apple Mac OS X 10.8 (64-bit)
  - `darwin11_64Guest` - Apple Mac OS X 10.7 (64-bit)
  - `darwin11Guest` - Apple Mac OS X 10.7 (32-bit)
  - `darwin10_64Guest` - Apple Mac OS X 10.6 (64-bit)
  - `darwin10Guest` - Apple Mac OS X 10.6 (32-bit)
  - `darwin64Guest` - Apple Mac OS X 10.5 (64-bit)
  - `darwinGuest` - Apple Mac OS X 10.5 (32-bit)
  - `vmkernel65Guest` - VMware ESXi 6.x
  - `vmkernel6Guest` - VMware ESXi 6.0
  - `vmkernel5Guest` - VMware ESXi 5.x
  - `vmkernelGuest` - VMware ESX 4.x
  - `eComStation2Guest` - Serenity Systems eComStation 2
  - `eComStationGuest` - Serenity Systems eComStation 1
  - `unixWare7Guest` - SCO UnixWare 7
  - `openServer6Guest` - SCO OpenServer 6
  - `openServer5Guest` - SCO OpenServer 5
  - `solaris11_64Guest` - Oracle Solaris 11 (64-bit)
  - `solaris10_64Guest` - Oracle Solaris 10 (64-bit)
  - `solaris10Guest` - Oracle Solaris 10 (32-bit)
  - `solaris9Guest` - Sun Microsystems Solaris 9
  - `solaris8Guest` - Sun Microsystems Solaris 8
  - `os2Guest` - IBM OS/2
  - `otherGuest64` - Other (64-bit)
  - `otherGuest` - Other (32-bit)
  - `netware6Guest` - Novell NetWare 6.x
  - `netware5Guest` - Novell NetWare 5.1

## Attribute Reference

The following additional attributes are exported:

* `internal_disk` - (*v2.7+*) A block providing internal disk of VM details. See [Internal Disk](#internalDisk) below for details.
* `disk.size_in_mb` - (*v2.7+*) Independent disk size in MB.

<a id="internalDisk"></a>
## Internal disk

* `disk_id` - (*v2.7+*) Specifies a unique identifier for this disk in the scope of the corresponding VM.
* `bus_type` - (*v2.7+*) The type of disk controller. Possible values: `ide`, `parallel`( LSI Logic Parallel SCSI), `sas`(LSI Logic SAS (SCSI)), `paravirtual`(Paravirtual (SCSI)), `sata`, `nvme`.
* `size_in_mb` - (*v2.7+*) The size of the disk in MB. 
* `bus_number` - (*v2.7+*) The number of the SCSI or IDE controller itself.
* `unit_number` - (*v2.7+*) The device number on the SCSI or IDE controller of the disk.
* `thin_provisioned` - (*v2.7+*) Specifies whether the disk storage is pre-allocated or allocated on demand.
* `iops` - (*v2.7+*) Specifies the IOPS for the disk. Default is 0.
* `storage_profile` - (*v2.7+*) Storage profile which overrides the VM default one.
* `vapp_id` - (*v3.12+*) Parent vApp ID.

## Hot and Cold update

These fields can be updated only when VM is **powered off** (provider automatically restarts the VM):

`cpu_cores`, `power_on`, `disk`, `expose_hardware_virtualization`, `boot_image`, `hardware_version`, `os_type`,
`description`, `cpu_hot_add_enabled`, `memory_hot_add_enabled`, `network`, `firmware`, `boot_options.efi_secure_boot`

These fields can be updated when VM is **powered on**:

`memory`, `cpus`, `network`, `metadata`, `guest_properties`, `sizing_policy_id`, `placement_policy_id`, `boot_options (except efi_secure_boot)` 

Notes about **removing** `network`:

* Guest OS must support hot NIC removal for NICs to be removed using network definition. If Guest OS doesn't support it - `power_on=false` can be used to power off the VM before removing NICs.
* Vcloud 10.1 has a bug and all NIC removals will be performed in cold manner.

## Extra Configuration

We can add, modify, and remove VM extra configuration items using the property `set_extra_config`, which consists on one or
more blocks with the following fields:

* `key` - (Required) Is the unique identifier of this item. It must not contain any spaces.
* `value` - (Optional) A value for this item. You can use it to insert new items or modify existing ones. Setting
  the value to an empty string will remove the item.

~> We should only insert or modify the fields we want to handle, without touching the ones already present in the VM.
   Vcloud uses the extra-config items for its own purposes, and modifying them could lead to destabilising side effects.

Notes:

1. To remove an item, we need to use the same `key` with which it was inserted, and set its `value` to `""`.
1. The property `set_extra_config` is only used as input. The state of the existing or modified configuration is reported
   in the read-only property `extra_config`, which has the following fields:
   * `key` - The key that identifies this entry
   * `value` - The value for this entry
   * `required` - Whether the entry is required

## Example of extra configuration

```hcl
resource "vcloud_vapp_vm" "web2" {
  # ...

  network {
    type               = "org"
    name               = "net"
    ip_allocation_mode = "DHCP"
  }

  set_extra_config {
    key   = "some-new-key"
    value = "some-new-value"
  }

  set_extra_config {
    key   = "some-other-new--key"
    value = "some-other-new-value"
  }
}
```

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcloud_vapp_vm" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = true # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

To remove all metadata one needs to specify an empty `metadata_entry`, like:

```
metadata_entry {}
```

The same applies also for deprecated `metadata` attribute:

```
metadata = {}
```

## Importing

Supported in provider *v2.6+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing VM can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.vapp-name.vm-name
For example, using this structure, representing a VM that was **not** created using Terraform:

```hcl
resource "vcloud_vapp_vm" "tf-vm" {
  name      = "my-vm"
  org       = "my-org"
  vdc       = "my-vdc"
  vapp_name = "my-vapp"
}
```

You can import such vapp into terraform state using this command

```
terraform import vcloud_vapp_vm.tf-vm my-org.my-vdc.my-vapp.my-vm
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR

After importing, the data for this VM will be in the state file (`terraform.tfstate`). If you want to use this
resource for further operations, you will need to integrate it with data from the state file, and with some data that
is used to create the VM, such as `catalog_name`, `template_name`.

[docs-import]:https://www.terraform.io/docs/import/
[vgpu-policy]:/providers/terraform-viettelidc/vcloud/latest/docs/resources/vm_vgpu_policy
