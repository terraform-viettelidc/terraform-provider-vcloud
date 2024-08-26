---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_virtual_service"
sidebar_current: "docs-vcloud-datasource-nsxt-alb-virtual-service"
description: |-
  Provides a data source to read ALB Virtual services for particular NSX-T Edge Gateway. A virtual service
  advertises an IP address and ports to the external world and listens for client traffic. When a virtual service receives
  traffic, it directs it to members in ALB Pool.
---

# vcloud\_nsxt\_alb\_virtual\_service

Supported in provider *v3.5+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a data source to read ALB Virtual services for particular NSX-T Edge Gateway. A virtual service
advertises an IP address and ports to the external world and listens for client traffic. When a virtual service receives
traffic, it directs it to members in ALB Pool.

## Example Usage

```hcl
data "vcloud_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcloud_nsxt_alb_virtual_service" "test" {
  org = "dainius"

  edge_gateway_id = vcloud_nsxt_edgegateway.existing.id
  name            = "virutal-service-name"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_edgegateway) data source
* `name` - (Required) The name of ALB Virtual Service

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_virtual_service`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_virtual_service) resource are
available.
