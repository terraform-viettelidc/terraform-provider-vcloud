---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vm_vgpu_policy"
sidebar_current: "docs-vcd-data-source-vm-vgpu-policy"
description: |-
  Provides a data source to read vGPU policies in Viettel IDC Cloud.
---

# vcd\_vm\_vgpu\_policy

Experimental in provider *3.11*.

Provides a data source to read vGPU policies in Viettel IDC Cloud.

## Example Usage

```hcl
data "vcloud_vm_vgpu_policy" "tf-policy-name" {
  name = "my-rule"
}

output "policyId" {
  value = data.vcloud_vm_vgpu_policy.tf-policy-name.id
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the VM vGPU policy.

All arguments defined in [`vcloud_vm_vgpu_policy`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/vm_vgpu_policy#argument-reference) are supported.

