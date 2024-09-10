---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_edgegateway_service_engine_group"
sidebar_current: "docs-vcd-datasource-nsxt-alb-edgegateway-service-engine-group"
description: |-
  Provides a datasource to read ALB Service Engine Group assignment to NSX-T Edge Gateway.
---

# vcloud\_nsxt\_alb\_edgegateway\_service\_engine\_group

Supported in provider *v3.5+* and VCLOUD 10.2+ with NSX-T and ALB.

Provides a datasource to read ALB Service Engine Group assignment to NSX-T Edge Gateway.

## Example Usage (Referencing Service Engine Group by ID)

```hcl
data "vcloud_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcloud_nsxt_alb_service_engine_group" "first" {
  name = "first-se"
}

data "vcloud_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id         = data.vcloud_nsxt_edgegateway.existing.id
  service_engine_group_id = data.vcloud_nsxt_alb_service_engine_group.first.id
}
```

## Example Usage (Referencing Service Engine Group by Name)

```hcl
data "vcloud_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcloud_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id           = data.vcloud_nsxt_edgegateway.existing.id
  service_engine_group_name = "known-service-engine-group-name"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Optional) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_edgegateway) data source
* `service_engine_group_id` - (Required) An ID of NSX-T Service Engine Group. Can be looked up using
  [vcloud_nsxt_alb_service_engine_group](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_service_engine_group) data
  source. **Note** Either `service_engine_group_name` or `service_engine_group_id` require it.
* `service_engine_group_name` - (Optional) A Name of NSX-T Service Engine Group. **Note** Either
  `service_engine_group_name` or `service_engine_group_id` require it.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_edgegateway_service_engine_group`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)
resource are available.
