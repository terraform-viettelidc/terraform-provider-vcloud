---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_rde_type_behavior"
sidebar_current: "docs-vcloud-data-source-rde-type-behavior"
description: |-
  Provides the capability of reading RDE Type Behaviors in Viettel IDC Cloud, which override an existing RDE Interface
  Behavior.
---

# vcloud\_rde\_type\_behavior

Provides the capability of reading RDE Type Behaviors in Viettel IDC Cloud, which override an existing [RDE Interface
Behavior](/providers/vmware/vcloud/latest/docs/data-sources/rde_interface_behavior).

Supported in provider *v3.10+*. Requires System Administrator privileges.

## Example Usage

```hcl
data "vcloud_rde_interface" "my_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

data "vcloud_rde_interface_behavior" "my_interface_behavior" {
  interface_id = data.vcloud_rde_interface.my_interface.id
  name         = "createKubeConfig"
}

data "vcloud_rde_type" "my_type" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.2.0"
}

data "vcloud_rde_type_behavior" "my_behavior" {
  rde_type_id               = data.vcloud_rde_type.my_type.id
  rde_interface_behavior_id = data.vcloud_rde_interface_behavior.my_interface_behavior.id
}

output "execution_id" {
  value = data.vcloud_rde_type_behavior.my_behavior.execution.id
}
```

## Argument Reference

The following arguments are supported:

* `rde_type_id` - (Required) The ID of the [RDE Type](/providers/vmware/vcloud/latest/docs/data-sources/rde_type) that owns the Behavior override
* `rde_interface_behavior_id` - (Required) The ID of the original [RDE Interface Behavior](/providers/vmware/vcloud/latest/docs/data-sources/rde_interface_behavior)

## Attribute Reference

All the supported attributes are defined in the
[RDE Type Behavior resource](/providers/vmware/vcloud/latest/docs/resources/rde_type_behavior#argument-reference).
