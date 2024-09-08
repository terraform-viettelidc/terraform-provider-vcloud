---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_rde_behavior_invocation"
sidebar_current: "docs-vcd-data-source-rde-behavior-invocation"
description: |-
   Provides the capability of invoking an existing Runtime Defined Entity Behavior in Viettel IDC Cloud.
---

# vcd\_rde\_behavior\_invocation

~> This feature is **experimental** and may change in future

Provides the capability of invoking an existing [RDE Interface Behavior](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_interface_behavior)
or [RDE Type Behavior](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_type_behavior) in Viettel IDC Cloud.

Supported in provider *v3.11+*

## Example Usage

```hcl
resource "vcloud_rde_interface" "interface" {
  nss     = "example"
  version = "1.2.3"
  vendor  = "vcloud"
  name    = "MyInterface"
}

resource "vcloud_rde_interface_behavior" "behavior" {
  rde_interface_id = vcloud_rde_interface.interface.id
  name             = "MyInterfaceBehavior"
  description      = "A cool behavior example"
  execution = {
    "id" : "Noop"
    "type" : "noop"
  }
}

resource "vcloud_rde_type" "type" {
  nss           = "example"
  version       = "1.1.0"
  vendor        = "vcloud"
  name          = "MyType"
  description   = "A cool type example"
  interface_ids = [vcloud_rde_interface.interface.id]
  schema        = file("../my-rde-type.json")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behaviors to wait for them to be created first.
  depends_on = [vcloud_rde_interface_behavior.behavior]
}

resource "vcloud_rde" "rde" {
  org          = "System"
  rde_type_id  = vcloud_rde_type.type.id
  name         = "MyRde"
  resolve      = true
  input_entity = file("../my-rde.json")
}

# Required Access Levels to invoke Behaviors
resource "vcloud_rde_type_behavior_acl" "interface_acl" {
  rde_type_id      = vcloud_rde_type.type.id
  behavior_id      = vcloud_rde_interface_behavior.behavior.id
  access_level_ids = ["urn:vcloud:accessLevel:FullControl"]
}

data "vcloud_rde_behavior_invocation" "invoke" {
  rde_id            = vcloud_rde.rde.id
  behavior_id       = vcloud_rde_interface_behavior.behavior.id
  invoke_on_refresh = true
}

output "rde_behavior_invocation_output" {
  value = data.vcloud_rde_behavior_invocation.invoke.result
}
```

## Argument Reference

The following arguments are supported:

* `rde_id` - (Required) The ID of the [RDE](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde) which Behavior will be invoked
* `behavior_id` - (Required) The ID of the [RDE Interface Behavior](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_interface_behavior) or
  the [RDE Interface Behavior](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_type_behavior) to invoke
* `arguments` - (Optional) A map with the arguments of the invocation
* `metadata` - (Optional) A map with the metadata of the invocation
* `invoke_on_refresh` - (Optional) Defaults to `true`, indicates whether the Behavior should be invoked on every refresh.
  E.g. to suppress the invocation on the refresh step of `terraform destroy` operations, or to stop invoking a given Behavior temporarily.

## Attribute Reference

The following attributes are supported:

* `result` - The invocation result in plain text
