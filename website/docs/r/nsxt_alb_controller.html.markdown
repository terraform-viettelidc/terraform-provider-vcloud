---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_controller"
sidebar_current: "docs-vcloud-resource-nsxt-alb-controller"
description: |-
  Provides a resource to manage ALB Controller for Providers. It helps to integrate Viettel IDC Cloud with Avi Load Balancer deployment. Controller instances are registered with Viettel IDC Cloud instance. Controller
  instances serve as a central control plane for the load-balancing services provided by Avi Load Balancer.
---

# vcloud\_nsxt\_alb\_controller

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a resource to manage ALB Controller for Providers. It helps to integrate Viettel IDC Cloud with Avi 
Load Balancer deployment. Controller instances are registered with Viettel IDC Cloud instance. Controller
instances serve as a central control plane for the load-balancing services provided by Avi Load Balancer.

~> Only `System Administrator` can create this resource.

~> Vcloud 10.3.0 has a caching bug which prevents listing importable clouds immediately (retrieved using
[`vcloud_nsxt_alb_importable_cloud`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_importable_cloud)) after ALB
Controller is created. This data should be available 15 minutes after the Controller is created.

## Example Usage (Adding ALB Controller to provider)

```hcl
resource "vcloud_nsxt_alb_controller" "first" {
  name         = "aviController1"
  description  = "first alb controller"
  url          = "https://my.controller"
  username     = "admin"
  password     = "CHANGE-ME"
  license_type = "ENTERPRISE"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for ALB Controller
* `description` - (Optional) An optional description ALB Controller
* `url` - (Required) The URL of ALB Controller
* `username` - (Required) The username for ALB Controller
* `password` - (Required) The password for ALB Controller. Password will not be refreshed.
* `license_type` - (Optional) License type of ALB Controller (`ENTERPRISE` or `BASIC`)

~> The attribute `license_type` must not be used in Vcloud 10.4+, it is replaced by [nsxt_alb_service_engine_group](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_service_engine_group) and [nsxt_alb_settings](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_settings) attribute `supported_feature_set`.

## Attribute Reference

The following attributes are exported on this resource:

* `version` - ALB Controller version (e.g. 20.1.3)


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing ALB Controller configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_alb_controller.imported my-controller-name
```

The above would import the `my-controller-name` ALB controller settings that are defined at provider level.