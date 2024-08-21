---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxv_firewall_rule"
sidebar_current: "docs-vcloud-data-source-nsxv-firewall-rule"
description: |-
  Provides a Viettel IDC Cloud firewall rule data source for advanced edge gateways (NSX-V). This can
  be used to read existing rules by ID and use its attributes in other resources.
---

# vcloud\_nsxv\_firewall\_rule

Provides a Viettel IDC Cloud firewall rule data source for advanced edge gateways (NSX-V). This can be
used to read existing rules by ID and use its attributes in other resources.

~> **Note:** This data source requires advanced edge gateway.

## Example Usage

```hcl
data "vcloud_nsxv_firewall_rule" "my-rule" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  rule_id = "133048" # real firewall rule ID, not the UI number
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the DNAT rule.
* `rule_id` - (Required) ID of firewall rule (not UI number). See more information about firewall
rule ID in `vcloud_nsxv_firewall_rule` [import section](/providers/vmware/vcloud/latest/docs/resources/nsxv_firewall_rule#listing-real-firewall-rule-ids).

## Attribute Reference

All the attributes defined in [`vcloud_nsxv_firewall_rule`](/providers/vmware/vcloud/latest/docs/resources/nsxv_firewall_rule)
resource are available.
