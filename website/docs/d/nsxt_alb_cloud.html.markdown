---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_cloud"
sidebar_current: "docs-vcloud-datasource-nsxt-alb-cloud"
description: |-
  Provides a data source to read ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct
  that consists of an NSX-T Manager and an NSX-T Data Center transport zone.
---

# vcloud\_nsxt\_alb\_cloud

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a data source to manage ALB Clouds for Providers. An NSX-T Cloud is a service provider-level construct that
consists of an NSX-T Manager and an NSX-T Data Center transport zone.

~> Only `System Administrator` can use this data source.

~> Vcloud 10.3.0 has a caching bug which prevents listing importable clouds immediately (retrieved using
[`vcloud_nsxt_alb_importable_cloud`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_alb_importable_cloud)) after ALB
Controller is created. This data should be available 15 minutes after the Controller is created.

## Example Usage

```hcl
data "vcloud_nsxt_alb_cloud" "first" {
  name = "cloud-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Name of ALB Cloud

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_cloud`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_cloud) resource are available.