---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edgegateway_dhcp_forwarding"
sidebar_current: "docs-vcloud-data-source-nsxt-edge-dhcp-forwarding"
description: |-
  Provides a data source to manage NSX-T Edge Gateway DHCP forwarding configuration.
---

# vcloud\_nsxt\_edgegateway\_dhcp\_forwarding

Supported in provider *v3.10+* and Vcloud 10.3.1+ with NSX-T.

Provides a data source to read NSX-T Edge Gateway DHCP forwarding configuration.

## Example Usage

```hcl
data "vcloud_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcloud_nsxt_edgegateway" "testing-in-vdc" {
  org      = "datacloud"
  owner_id = data.vcloud_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

data "vcloud_nsxt_edgegateway_dhcp_forwarding" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing-in-vdc.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Org in which the NSX-T Edge Gateway is located, required
  if not set in the provider section.
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_edgegateway_dhcp_forwarding`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_edgegateway_dhcp_forwarding)
resource are available.
