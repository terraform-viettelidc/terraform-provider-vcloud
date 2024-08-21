---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_ui_plugin"
sidebar_current: "docs-vcd-resource-ui-plugin"
description: |-
  Provides a Viettel IDC Cloud UI Plugin resource. This can be used to manage UI Plugins.
---

# vcd\_ui\_plugin

Provides a Viettel IDC Cloud UI Plugin resource. This can be used to manage UI Plugins in VCD, for example to add a new
plugin from a local ZIP file, to publish/unpublish a UI Plugin to different Organizations, etc.

-> Managing UI Plugins requires System Administrator privileges.

Supported in provider *v3.10+* and requires VCD 10.3+

## Example Usage with specific Organizations to publish

```hcl
locals {
  my_plugin_orgs = [
    "myOrg1",
    "myOrg2"
  ]
}

data "vcloud_org" "my_plugin_orgs" {
  count = length(local.my_plugin_orgs)
  name  = local.my_plugin_orgs[count.index]
}

resource "vcloud_ui_plugin" "my_plugin" {
  plugin_path = "./container-ui-plugin-4.0.zip"
  enabled     = true
  tenant_ids  = data.vcloud_org.my_plugin_orgs[*].id
}
```

## Example Usage publishing to all Organizations available

```hcl
data "vcloud_resource_list" "all_orgs" {
  name          = "all_orgs"
  resource_type = "vcloud_org"
  list_mode     = "id"
}

resource "vcloud_ui_plugin" "my_plugin" {
  plugin_path = "./container-ui-plugin-4.0.zip"
  enabled     = true
  tenant_ids  = data.vcloud_resource_list.all_orgs.list
}
```

## Argument Reference

The following arguments are supported:

* `plugin_path` - (Required) Path to a .zip file that contains the bundled UI Plugin
* `enabled` - (Required) Whether the UI Plugin will be enabled (`true`) or not (`false`)
* `tenant_ids` - (Optional) The identifiers of the [Organizations](/providers/vmware/vcd/latest/docs/data-sources/org)
  that will be able to use the UI Plugin if enabled. If not set, it doesn't publish to any Organization.
* `provider_scoped` - (Optional) **Can only be set on updates**, the initial value is taken from the JSON manifest.
  Changes the scope of the UI Plugin for System providers. It should be set to `true` when the UI Plugin is published to the System organization, to prevent
  unwanted updates-in-place.
* `tenant_scoped` - (Optional) **Can only be set on updates**, the initial value is taken from the JSON manifest.
  Changes the scope of the UI Plugin for Organization users. It should be set to `true` when the UI Plugin is published to any organization, to prevent
  unwanted updates-in-place.

## Attribute Reference

* `vendor` - The vendor of the UI Plugin
* `name` - The name of the UI Plugin
* `version` - The version of the UI Plugin
* `license` - The license of the UI Plugin
* `link` - The website or custom URL of the UI Plugin
* `description` - The description of the UI Plugin
* `status` - The status of the UI Plugin (for example, `ready`, `unavailable`, etc)

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing UI Plugin can be [imported][docs-import] into this resource via supplying its vendor, name and version, which
unequivocally identifies it.
For example, using this structure, representing an existing UI Plugin that was **not** created using Terraform:

```hcl
resource "vcloud_ui_plugin" "my_existing_plugin" {
  # `plugin_path` is not needed as it was already created
  enabled = true
}
```

For example, you can import the "Customize Portal" UI Plugin into Terraform state using this command

```
terraform import vcloud_ui_plugin.my_plugin VMware."Customize Portal".3.1.4
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the UI Plugin as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the UI Plugin's stored properties.
