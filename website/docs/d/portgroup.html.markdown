---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_portgroup"
sidebar_current: "docs-vcloud-data-source-portgroup"
description: |-
  Provides a data source for available vCenter Port Groups.
---

# vcloud\_portgroup

Provides a data source for available vCenter Port Groups.

Supported in provider *v3.0+*


## Example Usage for vSwitch Port Group

```hcl
data "vcloud_portgroup" "first-pg-vswitch" {
  name = "pg-name"
  type = "NETWORK"
}
```

## Example Usage for Distributed Port Group

```hcl
data "vcloud_portgroup" "first-pg-dvswitch" {
  name = "pg-name"
  type = "DV_PORTGROUP"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Organization VDC name
* `type` - (Required) `NETWORK` for vSwitch port group or `DV_PORTGROUP` for distributed port group.

## Attribute reference

Only ID is set to be able and reference in other resources or data sources.
