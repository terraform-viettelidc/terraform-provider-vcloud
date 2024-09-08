---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_pool"
sidebar_current: "docs-vcd-datasource-nsxt-alb-pool"
description: |-
  Provides a data source to read ALB Pools for particular NSX-T Edge Gateway. Pools maintain the list of servers
  assigned to them and perform health monitoring, load balancing, persistence. A pool may only be used or referenced by
  only one virtual service at a time.
---

# vcd\_nsxt\_alb\_pool

Supported in provider *v3.5+* and VCLOUD 10.2+ with NSX-T and ALB.

Provides a data source to read ALB Pools for particular NSX-T Edge Gateway. Pools maintain the list of servers
assigned to them and perform health monitoring, load balancing, persistence. A pool may only be used or referenced by
only one virtual service at a time.

## Example Usage

```hcl
data "vcloud_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcloud_nsxt_alb_pool" "test" {
  org = "my-org"

  edge_gateway_id = vcloud_nsxt_alb_settings.existing.edge_gateway_id
  name            = "existing-alb-pool-1"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_edgegateway) data source
* `name` - (Required) Name of existing ALB Pool.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_pool`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_pool) resource are available.
