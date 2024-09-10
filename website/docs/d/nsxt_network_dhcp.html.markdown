---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_network_dhcp"
sidebar_current: "docs-vcd-datasource-nsxt-network-dhcp"
description: |-
  Provides a data source to read DHCP pools for NSX-T Org VDC networks.
---

# vcloud\_nsxt\_network\_dhcp

Provides a data source to read DHCP pools for NSX-T Org VDC networks.

## Example Usage 1

```hcl

data "vcloud_network_routed_v2" "parent" {
  org = "my-org"
  vdc = "my-vdc"

  name = "my-parent-network"
}

data "vcloud_nsxt_network_dhcp" "pools" {
  org = "my-org"

  org_network_id = data.vcloud_network_routed_v2.parent.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `org_network_id` - (Required) ID of parent Org VDC Routed network

## Attribute Reference

All the attributes defined in [`vcloud_nsxt_network_dhcp`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_dhcp)
resource are available.
