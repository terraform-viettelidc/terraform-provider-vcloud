---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_vm"
sidebar_current: "docs-vcloud-datasource-vm"
description: |-
  Provides a Viettel IDC Cloud VM data source. This can be used to access standalone VMs.
---

# vcloud\_vm

Provides a Viettel IDC Cloud standalone VM data source. This can be used to access standalone VMs.

Supported in provider *v3.2+*

## Example Usage

```hcl
data "vcloud_vm" "web1" {
  name = "web1"
}
```

## General notes

* Although from the UI standpoint a standalone VM appears to exist without a vApp, in reality there is a hidden vApp that
is generated automatically when the VM is created, and removed when the VM is terminated. The field `vapp_name` is populated
  with the hidden vApp name, and readable in Terraform state.
  
* The VM name is unique **within the vApp**, which means that it is possible to create multiple standalone VMs with the same name.
This fact has consequences when defining a data source, where we identify the VM by name. If there are duplicates, we get
  an error message containing the list of VMs that share the same name. To retrieve a specific VM in such scenario, we need
  to provide the VM ID in the `name` field
  
For example, given this input
```hcl
data "vcloud_vm" "test_vm" {
  name = "TestVm"
}
```

If multiple VMs have the same name, we get an error like:

```
$ terraform apply -auto-approve
data.vcloud_vm.test_vm: Refreshing state...

Error: [VM read] error retrieving VM TestVm by name: more than one VM found with name TestVm
ID                                                 Guest OS                       Network
-------------------------------------------------- ------------------------------ --------------------
urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003 Other 3.x Linux (64-bit)       (net-datacloud-r - 192.168.2.2)
urn:vcloud:vm:41d5d5a7-040e-49cb-a516-5a604211a395 Debian GNU/Linux 10 (32-bit)   (net-datacloud-i - DHCP)

[ENF] entity not found
```

We can achieve the goal by providing the ID instead of the name

```hcl
data "vcloud_vm" "test_vm" {
  name = "urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003"
}
```

```
 terraform apply -auto-approve
data.vcloud_vm.test_vm: Refreshing state...

Apply complete! Resources: 0 added, 0 changed, 0 destroyed.
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A name or ID for the standalone VM in VDC

## Attributes reference

This data source provides all attributes available for [`vcloud_vapp_vm`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/vapp_vm).

