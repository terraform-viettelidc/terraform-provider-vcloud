---
layout: "vcd"
page_title: "Viettel IDC Cloud: security_tag"
sidebar_current: "docs-vcd-resource-security-tag"
description: |-
Provides a Viettel IDC Cloud Security Tag resource. This can be
used to assign security tag to VMs.
---

# vcd\_security\_tag

Provides a Viettel IDC Cloud Security Tag resource. This can be
used to assign security tag to VMs.

Supported in provider *v3.7+* and requires VCD 10.3.0+

~> **Note:** Only one of `vcloud_security_tag` resource or [`security_tags` attribute from `vcloud_vapp_vm`](/providers/vmware/vcd/latest/docs/resources/vapp_vm)
should be used. Using both would cause a behavioral conflict.

-> **Note:** This resource requires either system or org administrator privileges.

## Example Usage

```hcl
resource "vcloud_security_tag" "my_tag" {
  name   = "test-tag"
  vm_ids = [vcloud_vm.my-vm-one.id, vcloud_vm.my-vm-two.id]
}
```
## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) The name of the security tag.
* `vm_ids` - (Required) List of VM IDs that the security tag is going to be applied to.

-> The ID of `vcloud_security_tag` is set to its name since VCD behind the scenes doesn't create an ID.

# Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing security tag can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.security-tag-name
An example is below:

```
terraform import vcloud_security_tag.my-tag my-org.my-security-tag-name
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR


[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the security tag as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the security tag stored properties.
