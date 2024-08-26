---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_rde_interface_behavior"
sidebar_current: "docs-vcloud-data-source-rde-interface-behavior"
description: |-
   Provides the capability of fetching an existing RDE Interface Behavior from Viettel IDC Cloud.
---

# vcloud\_rde\_interface\_behavior

Provides the capability of fetching an existing RDE Interface Behavior from Viettel IDC Cloud.

Supported in provider *v3.10+*. Requires System Administrator privileges.

## Example Usage

```hcl
data "vcloud_rde_interface" "my_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

data "vcloud_rde_interface_behavior" "my_behavior" {
  rde_interface_id = data.vcloud_rde_interface.my_interface.id
  name             = "createKubeConfig"
}

output "execution_id" {
  value = data.vcloud_rde_interface_behavior.my_behavior.execution.id
}

output "execution_type" {
  value = data.vcloud_rde_interface_behavior.my_behavior.execution.type
}
```

## Argument Reference

The following arguments are supported:

* `rde_interface_id` - (Required) The ID of the RDE Interface that owns the Behavior to fetch
* `name` - (Required) The name of the RDE Interface Behavior to fetch

## Attribute Reference

All the supported attributes are defined in the
[RDE Interface Behavior resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_interface_behavior#argument-reference).
