---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_settings"
sidebar_current: "docs-vcloud-data-source-nsxt-edge-dhcpv6"
description: |-
  Provides a data source to read DHCPv6 configuration for NSX-T Edge Gateways.
---

# vcloud\_nsxt\_edgegateway\_dhcpv6

Provides a data source to read DHCPv6 configuration for NSX-T Edge Gateways.

## Example Usage

```hcl
data "vcloud_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing-in-vdc.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcloud_nsxt_edgegateway](/providers/vmware/vcloud/latest/docs/data-sources/nsxt_edgegateway) data source

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_edgegateway_dhcpv6`](/providers/vmware/vcloud/latest/docs/resources/nsxt_edgegateway_dhcpv6)
resource are available.
