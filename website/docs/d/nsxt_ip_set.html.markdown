---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_ip_set"
sidebar_current: "docs-vcd-datasource-nsxt-ip-set"
description: |-
  Provides a data source to read NSX-T IP Set. IP Sets are groups of objects to which the firewall rules apply. 
  Combining multiple objects into IP Sets helps reduce the total number of firewall rules to be created.
---

# vcd\_nsxt\_ip\_set

Supported in provider *v3.3+* and VCLOUD 10.1+ with NSX-T backed VDCs.

Provides a data source to read NSX-T IP Set. IP Sets are groups of objects to which the firewall rules apply. Combining
multiple objects into IP Sets helps reduce the total number of firewall rules to be created.

## Example Usage

```hcl
data "vcloud_nsxt_edgegateway" "main" {
  org  = "my-org" # Optional
  name = "main-edge"
}

data "vcloud_nsxt_ip_set" "my-set-1" {
  org = "my-org" # Optional

  edge_gateway_id = data.vcloud_nsxt_edgegateway.main.id

  name = "frontend-servers"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level. **Deprecated**
in favor of `edge_gateway_id` field.
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using
* `name` - (Required)  - Unique name of existing IP Set.

## Attribute Reference
* `owner_id` - Parent VDC or VDC Group ID.

All the arguments and attributes defined in
[`vcloud_nsxt_ip_set`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_ip_set) resource are available.
