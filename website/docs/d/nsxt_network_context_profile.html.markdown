---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_network_context_profile"
sidebar_current: "docs-vcd-data-source-nsxt-network-context-profile"
description: |-
  Provides a data source for NSX-T Network Context Profile lookup to later be used in Distributed
  Firewall.
---

# vcloud\_nsxt\_network\_context\_profile

Provides a data source for NSX-T Network Context Profile lookup to later be used in Distributed
Firewall.

## Example Usage (SYSTEM scope network context profile lookup in a VDC Group)

```hcl
data "vcloud_vdc_group" "existing" {
  org  = "my-org" # Optional, can be inherited from Provider configuration
  name = "main-vdc-group"
}

data "vcloud_nsxt_network_context_profile" "cp1" {
  context_id = data.vcloud_vdc_group.existing.id
  name       = "CTRXICA"
  scope      = "SYSTEM"
}
```

## Example Usage (SYSTEM profile lookup in an NSX-T Manager)
```hcl
data "vcloud_nsxt_manager" "main" {
  name = "first-nsxt-manager"
}

data "vcloud_nsxt_network_context_profile" "p" {
  context_id = data.vcloud_nsxt_manager.main.id
  name       = "CTRXICA"
  scope      = "SYSTEM"
}
``` 

## Argument Reference

The following arguments are supported:

* `context_id` - (Required) Context ID specifies the context for Network Context Profile look up.
  This ID can be one of `VDC Group ID` (data source `vcloud_vdc_group`), `Org VDC ID` (data source
  `vcloud_org_vdc`), or `NSX-T Manager ID` (data source `vcloud_nsxt_manager`)
* `scope` - (Optional) Can be one of `SYSTEM`, `TENANT`, `PROVIDER`. (default `SYSTEM`)
* `name` - (Required) Name of Network Context Profile

## Attribute Reference

`id` - can be used in `vcloud_nsxt_distributed_firewall` resource field `network_context_profile_ids`
