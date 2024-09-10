---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_distributed_firewall"
sidebar_current: "docs-vcd-data-source-nsxt-distributed-firewall"
description: |-
  The Distributed Firewall data source reads all defined rules for a particular VDC Group.
---

# vcloud\_nsxt\_distributed\_firewall

The Distributed Firewall data source reads all defined rules for a particular VDC Group.

-> There is a different data source
[`vcloud_nsxt_distributed_firewall_rule`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_distributed_firewall_rule)
resource are available that can fetch a single firewall rule by name.

## Example Usage

```hcl
data "vcloud_vdc_group" "g1" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "my-vdc-group"
}

data "vcloud_nsxt_distributed_firewall" "t1" {
  org          = "my-org" # Optional, can be inherited from Provider configuration
  vdc_group_id = data.vcloud_vdc_group.g1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization in which Distributed Firewall is located. Optional if
  defined at provider level.
* `vdc_group_id` - (Required) The ID of a VDC Group

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_distributed_firewall`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_distributed_firewall)
resource are available.
