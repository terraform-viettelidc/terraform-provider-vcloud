---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_controller"
sidebar_current: "docs-vcloud-datasource-nsxt-alb-controller"
description: |-
  Provides a data source to read ALB Controller for Providers. It helps to integrate Viettel IDC Cloud with
  Avi Load Balancer deployment. Controller instances are registered with Viettel IDC Cloud instance.
  Controller instances serve as a central control plane for the load-balancing services provided by Avi Load
  Balancer.
---

# vcloud\_nsxt\_alb\_controller

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a data source to read ALB Controller for Providers. It helps to integrate Viettel IDC Cloud with
Avi Load Balancer deployment. Controller instances are registered with Viettel IDC Cloud instance.
Controller instances serve as a central control plane for the load-balancing services provided by Avi Load
Balancer.

~> Only `System Administrator` can use this data source.

~> Vcloud 10.3.0 has a caching bug which prevents listing importable clouds immediately (retrieved using
[`vcloud_nsxt_alb_importable_cloud`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_importable_cloud)) after ALB
Controller is created. This data should be available 15 minutes after the Controller is created.

## Example Usage

```hcl
data "vcloud_nsxt_alb_controller" "first" {
  name = "avi controller"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Unique name of existing ALB Controller.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_controller`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_controller) resource are available.
