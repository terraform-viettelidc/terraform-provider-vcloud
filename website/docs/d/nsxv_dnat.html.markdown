---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_nsxv_dnat"
sidebar_current: "docs-vcd-data-source-nsxv-dnat"
description: |-
  Provides a VMware Cloud Director DNAT data source for advanced edge gateways (NSX-V). This can be used to read
  existing rule by ID and use its attributes in other resources.
---

# vcd\_nsxv\_dnat

Provides a VMware Cloud Director DNAT data source for advanced edge gateways (NSX-V). This can be used to
read existing rule by ID and use its attributes in other resources.

~> **Note:** This data source requires advanced edge gateway.

## Example Usage

```hcl
data "vcloud_nsxv_dnat" "my-rule" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  rule_id = "197864"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT rule.
* `rule_id` - (Required) ID of DNAT rule as shown in the UI.

## Attribute Reference

All the attributes defined in `vcloud_nsxv_dnat` resource are available.
