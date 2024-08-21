---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_security_group"
sidebar_current: "docs-vcloud-data-source-nsxt-security-group"
description: |-
  Provides a data source to access NSX-T Security Group configuration. Security Groups are groups of
  data center group networks to which distributed firewall rules apply. Grouping networks helps you 
  to reduce the total number of distributed firewall rules to be created. 
---

# vcloud\_nsxt\_security\_group

Supported in provider *v3.3+* and Vcloud 10.1+ with NSX-T backed VDCs.

Provides a data source to access NSX-T Security Group configuration. Security Groups are groups of
data center group networks to which distributed firewall rules apply. Grouping networks helps you to
reduce the total number of distributed firewall rules to be created.

## Example Usage 1

```hcl

data "vcloud_nsxt_edgegateway" "main" {
  org  = "my-org" # Optional
  name = "main-edge"
}

data "vcloud_nsxt_security_group" "group1" {
  org = "my-org" # Optional

  edge_gateway_id = data.vcloud_nsxt_edgegateway.main.id

  name = "test-security-group-changed"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Deprecated; Optional) The name of VDC to use, optional if defined at provider level. **Deprecated**
  in favor of `edge_gateway_id` field.
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using
* `name` - (Required)  - Unique name of existing Security Group.

## Attribute Reference
* `owner_id` - Parent VDC or VDC Group ID.
 
All the arguments and attributes defined in
[`vcloud_nsxt_security_group`](/providers/vmware/vcloud/latest/docs/resources/nsxt_security_group) resource are available.
