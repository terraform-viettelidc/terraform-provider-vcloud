---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_manager"
sidebar_current: "docs-vcd-data-source-nsxt-manager"
description: |-
  Provides a data source for available NSX-T manager.
---

# vcloud\_nsxt\_manager

Provides a data source for NSX-T manager.

Supported in provider *v3.0+*

~> **Note:** This resource uses new Viettel IDC Cloud
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least VCLOUD *10.1.1+* and NSX-T *3.0+*.

## Example Usage 

```hcl
data "vcloud_nsxt_manager" "main" {
  name = "nsxt-manager-one"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) NSX-T manager name

## Attribute reference

* `id` - ID of the manager
* `href` - Full URL of the manager
