---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_ui_plugin"
sidebar_current: "docs-vcloud-datasource-ui-plugin"
description: |-
  Provides a Viettel IDC Cloud UI Plugin data source. This can be used to fetch and read an existing UI Plugin.
---

# vcloud\_ui\_plugin

Provides a Viettel IDC Cloud UI Plugin data source. This can be used to fetch and read an existing UI Plugin.

-> Reading UI Plugins requires System Administrator privileges.

Supported in provider *v3.10+* and requires Vcloud 10.3+

## Example Usage

```hcl
data "vcloud_ui_plugin" "existing_ui_plugin" {
  vendor  = "VMware"
  name    = "Customize Portal"
  version = "3.1.4"
}

output "license" {
  value = data.vcloud_ui_plugin.existing_ui_plugin.license
}

output "tenants" {
  value = data.vcloud_ui_plugin.existing_ui_plugin.tenant_ids
}
```

## Argument Reference

The following arguments are supported:

* `vendor` - (Required) The vendor of the UI Plugin
* `name` - (Required) The name of the UI Plugin
* `version` - (Required) The version of the UI Plugin

## Attribute Reference

All **attributes** defined in [`vcloud_ui_plugin`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/ui_plugin#attribute-reference) are supported.

Also, the arguments `enabled`, `tenant_ids`, `provider_scoped` and `tenant_scoped` that are defined in the same resource
are available as read-only attributes. 