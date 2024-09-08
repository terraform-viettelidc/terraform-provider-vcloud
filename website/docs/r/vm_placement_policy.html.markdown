---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vm_placement_policy"
sidebar_current: "docs-vcd-resource-vm-placement-policy"
description: |-
  Provides a Viettel IDC Cloud VM Placement Policy resource. This can be
  used to create, modify, and delete VM Placement Policies.
---

# vcd\_vm\_placement\_policy

Provides a Viettel IDC Cloud VM Placement Policy resource. This can be used to create, modify, and delete VM Placement Policy.

Supported in provider *v3.8+* and requires VCLOUD 10.2+

-> **Note:** This resource requires system administrator privileges.

## Example Usage

```hcl
data "vcloud_provider_vdc" "pvdc" {
  name = "my-pvdc"
}

data "vcloud_vm_group" "vm-group" {
  name            = "vmware-vm-group"
  provider_vdc_id = data.vcloud_provider_vdc.pvdc.id
}

resource "vcloud_vm_placement_policy" "test-placement-pol" {
  name            = "my-placement-pol"
  description     = "My awesome VM Placement Policy"
  provider_vdc_id = data.vcloud_provider_vdc.pvdc.id
  vm_group_ids    = [data.vcloud_vm_group.vm-group.id]
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of VM Placement Policy.
* `provider_vdc_id` - (Required) The ID of the Provider VDC to which this VM Placement Policy belongs.
* `description` - (Optional) description of VM Placement Policy.
* `vm_group_ids` - (Optional) IDs of the collection of VMs with similar host requirements. **Note:** Either `vm_group_ids` or `logical_vm_group_ids` must be set.
* `logical_vm_group_ids` - (Optional) IDs of one or more Logical VM Groups to define this VM Placement policy. There is an AND relationship among all the entries set in this attribute. **Note:** Either `vm_group_ids` or `logical_vm_group_ids` must be set.

# Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing an VM Placement Policy can be [imported][docs-import] into this resource
via supplying the full dot separated path to VM Placement Policy. An example is
below:

```
terraform import vcloud_vm_placement_policy.my-policy policy_name_in_vcd
```
or using IDs:
```
terraform import vcloud_vm_placement_policy.my-policy urn:vcloud:vdcComputePolicy:39579936-6211-40b5-adf6-2a74d4413e9e
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCLOUD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the VM Placement Policy as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the VM Placement Policy stored properties.

### Listing VM Placement Policies

If you want to list IDs there is a special command **`terraform import vcloud_vm_placement_policy.imported list@`**. 
The output for this command should look similar to the one below:

```
terraform import vcloud_vm_placement_policy.imported list@
vcloud_vm_placement_policy.import: Importing from ID "list@"...
Retrieving all VM Placement Policies
No	ID									Name	
--	--									----	
1	urn:vcloud:vdcComputePolicy:100dc35a-572b-4876-a762-c734d67c56ef	tf_policy_3
2	urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b	tf_policy_2

```

Now to import VM Placement Policy with ID urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b one could supply this command:

```shell
$ terraform import vcloud_vm_placement_policy.imported urn:vcloud:vdcComputePolicy:446d623e-1eec-4c8c-8a14-2f7e6086546b
```
