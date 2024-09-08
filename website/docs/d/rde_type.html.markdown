---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_rde_type"
sidebar_current: "docs-vcd-data-source-rde-type"
description: |-
   Provides the capability of fetching an existing Runtime Defined Entity Type from Viettel IDC Cloud.
---

# vcd\_rde\_type

Provides the capability of fetching an existing Runtime Defined Entity Type from Viettel IDC Cloud.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcloud_rde_type" "my_rde_type" {
  vendor  = "bigcorp"
  nss     = "tech"
  version = "1.2.3"
}

output "type-name" {
  value = data.vcloud_rde_type.my_rde_type.name
}

output "type-id" {
  value = data.vcloud_rde_type.my_rde_type.id
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the Runtime Defined Entity Type.
* `nss` - (Required) A unique namespace associated with the Runtime Defined Entity Type.
* `version` - (Required) The version of the Runtime Defined Entity Type. Must follow [semantic versioning](https://semver.org/) syntax.

## Attribute Reference

All the supported attributes are defined in the
[Runtime Defined Entity Type resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_type#argument-reference).
