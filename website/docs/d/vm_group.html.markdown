---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vm_group"
sidebar_current: "docs-vcd-data-source-vm-group"
description: |-
  Provides a Viettel IDC Cloud VM Group data source. This can be used to fetch vSphere VM Groups and create VM Placement Policies with them.
---

# vcd\_vm\_group

Provides a Viettel IDC Cloud VM Group data source. This can be used to fetch vSphere VM Groups and create VM Placement Policies with them.

Supported in provider *v3.8+*

## Example Usage

```hcl
data "vcloud_provider_vdc" "my-vdc" {
  name = "my-pvdc"
}

data "vcloud_vm_group" "vm-group" {
  name            = "vmware-license-group"
  provider_vdc_id = data.vcloud_provider_vdc.my-vdc.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of VM Group to fetch from vSphere.
* `provider_vdc_id` - (Required) The ID of [Provider VDC](/providers/vmware/vcd/latest/docs/data-sources/provider_vdc) to which the VM Group belongs.

## Attributes reference

* `cluster_name` - Name of the vSphere cluster associated to this VM Group.
* `named_vm_group_id` - ID of the named VM Group. Used to create Logical VM Groups.
* `vcenter_id` - ID of the vCenter server.
* `cluster_moref` - Managed object reference of the vSphere cluster associated to this VM Group.
