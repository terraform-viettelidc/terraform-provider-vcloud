---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_edgegateway_service_engine_group"
sidebar_current: "docs-vcloud-resource-nsxt-alb-edge-service-engine-group"
description: |-
  Provides a resource to manage ALB Service Engine Group assignment to Edge Gateway.
---

# vcloud\_nsxt\_alb\_edgegateway\_service\_engine\_group

Supported in provider *v3.5+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a resource to manage ALB Service Engine Group assignment to NSX-T Edge Gateway.

~> Only `System Administrator` can create this resource.

## Example Usage (Enabling ALB on NSX-T Edge Gateway)

```hcl
data "vcloud_nsxt_edgegateway" "existing" {
  org = "my-org"
  vdc = "nsxt-vdc"

  name = "nsxt-gw"
}

data "vcloud_nsxt_alb_service_engine_group" "first" {
  name = "first-se"
}

resource "vcloud_nsxt_alb_edgegateway_service_engine_group" "first" {
  edge_gateway_id         = data.vcloud_nsxt_edgegateway.existing.id
  service_engine_group_id = data.vcloud_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 100
  reserved_virtual_services = 30
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the edge gateway belongs. Optional if defined at provider level.
* `edge_gateway_id` - (Required) An ID of NSX-T Edge Gateway. Can be looked up using
  [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_edgegateway) data source.
* `service_engine_group_id` - (Required) An ID of NSX-T Service Engine Group. Can be looked up using
  [vcloud_nsxt_alb_service_engine_group](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_service_engine_group) data
  source.
* `max_virtual_services` - (Optional) Maximum amount of Virtual Services to run on this Service Engine Group. **Only for
  Shared Service Engine Groups**.
* `reserved_virtual_services` - (Optional) Number of reserved Virtual Services for this Edge Gateway. **Only for Shared
  Service Engine Groups.**

## Attribute reference

* `deployed_virtual_services` -  Number of deployed Virtual Services on this Service Engine Group.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway ALB Service Engine Group assignment configuration can be [imported][docs-import] into
this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_alb_settings.imported my-org.my-org-vdc-org-vdc-group-name.my-nsxt-edge-gateway-name.service-engine-group-name
```

The above would import the NSX-T Edge Gateway ALB Service Engine Group assignment configuration for
Service Engine Group Name `service-engine-group-name` on  Edge Gateway named
`my-nsxt-edge-gateway-name` in Org `my-org` and VDC or VDC Group `my-org-vdc-org-vdc-group-name`.