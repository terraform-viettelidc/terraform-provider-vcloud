---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_cloud"
sidebar_current: "docs-vcloud-resource-nsxt-alb-cloud"
description: |-
  Provides a resource to manage ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct that
  consists of an NSX-T Manager and an NSX-T Data Center transport zone.
---

# vcloud\_nsxt\_alb\_cloud

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a resource to manage ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct that
consists of an NSX-T Manager and an NSX-T Data Center transport zone.

~> Only `System Administrator` can create this resource.

~> Vcloud 10.3.0 has a caching bug which prevents listing importable clouds immediately (retrieved using
[`vcloud_nsxt_alb_importable_cloud`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_importable_cloud)) after ALB
Controller is created. This data should be available 15 minutes after the Controller is created.


## Example Usage (Adding ALB Cloud)

```hcl
data "vcloud_nsxt_alb_controller" "main" {
  name = "aviController1"
}

data "vcloud_nsxt_alb_importable_cloud" "cld" {
  name          = "NSXT Importable Cloud"
  controller_id = vcloud_nsxt_alb_controller.first.id
}

resource "vcloud_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "ALB Cloud"

  controller_id       = data.vcloud_nsxt_alb_controller.main.id
  importable_cloud_id = data.vcloud_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcloud_nsxt_alb_importable_cloud.cld.network_pool_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for ALB Cloud
* `description` - (Optional) An optional description ALB Cloud
* `controller_id` - (Required) ALB Controller ID
* `importable_cloud_id` - (Required) Importable Cloud ID. Can be looked up using `vcloud_nsxt_alb_importable_cloud` data
  source
* `network_pool_id` - (Required) Network pool ID for ALB Cloud. Can be looked up using `vcloud_nsxt_alb_importable_cloud` data
  source


## Attribute Reference

The following attributes are exported on this resource:

* `health_status` - HealthStatus contains status of the Load Balancer Cloud. Possible values are:
  * UP - The cloud is healthy and ready to enable Load Balancer for an Edge Gateway
  * DOWN - The cloud is in a failure state. Enabling Load balancer on an Edge Gateway may not be possible
  * RUNNING - The cloud is currently processing. An example is if it's enabling a Load Balancer for an Edge Gateway
  * UNAVAILABLE - The cloud is unavailable
  * UNKNOWN - The cloud state is unknown
* `health_message` - DetailedHealthMessage contains detailed message on the health of the Cloud
* `network_pool_name` - Network Pool Name used by the Cloud


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing ALB Cloud configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_alb_cloud.imported my-alb-cloud-name
```

The above would import the `my-alb-cloud-name` ALB cloud settings that are defined at provider level.