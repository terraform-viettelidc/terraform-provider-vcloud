---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_alb_service_engine_group"
sidebar_current: "docs-vcloud-datasource-nsxt-alb-service-engine-group"
description: |-
  Provides a data source to read ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
  defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
  group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
  between different service engine groups.
---

# vcloud\_nsxt\_alb\_service\_engine\_group

Supported in provider *v3.4+* and Vcloud 10.2+ with NSX-T and ALB.

Provides a data source to read ALB Service Engine Groups. A Service Engine Group is an isolation domain that also
defines shared service engine properties, such as size, network access, and failover. Resources in a service engine
group can be used for different virtual services, depending on your tenant needs. These resources cannot be shared
between different service engine groups.

~> Only `System Administrator` can use this data source.

## Example Usage

```hcl
data "vcloud_nsxt_alb_service_engine_group" "demo" {
  name            = "configured-service-engine-group"
  sync_on_refresh = false
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required)  - Name of Service Engine Group.
* `sync_on_refresh` - (Optional) - A special argument that is not passed to Vcloud, but alters behaviour of this resource so
  that it performs a Sync operation on every Terraform refresh. *Note* this may impact refresh performance, but should
  ensure up-to-date information is read. Default is **false**.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_alb_service_engine_group`](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_service_engine_group)
resource are available.
