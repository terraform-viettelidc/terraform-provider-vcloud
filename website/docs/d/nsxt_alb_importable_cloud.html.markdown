---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_importable_cloud"
sidebar_current: "docs-vcloud-datasource-nsxt-alb-importable-cloud"
description: |-
  Provides a data source to reference existing ALB Importable Clouds. An NSX-T Importable Cloud is a reference to a
  Cloud configured in ALB Controller.
---

# vcloud\_nsxt\_alb\_importable\_cloud

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a data source to reference existing ALB Importable Clouds. An NSX-T Importable Cloud is a reference to a
Cloud configured in ALB Controller.

~> Only `System Administrator` can use this data source.

~> Vcloud 10.3.0 has a caching bug which prevents listing importable clouds immediately after [ALB
Controller](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_controller) is created. This data should be
available 15 minutes after the Controller is created.

## Example Usage

```hcl
data "vcloud_nsxt_alb_controller" "first" {
  name = "alb-controller"
}

data "vcloud_nsxt_alb_importable_cloud" "cld" {
  name          = "NSXT Importable Cloud"
  controller_id = data.vcloud_nsxt_alb_controller.first.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Name of ALB Importable Cloud
* `controller_id` - (Required)  - ALB Controller ID

## Attribute Reference

* `already_imported` - boolean value which displays if the ALB Importable Cloud is already consumed
* `network_pool_id` - backing network pool ID 
* `network_pool_name` - backing network pool ID
* `transport_zone_name` - backing transport zone name
