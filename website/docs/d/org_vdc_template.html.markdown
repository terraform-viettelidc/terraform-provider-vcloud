---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_org_vdc_template"
sidebar_current: "docs-vcd-data-source-org-vdc-template"
description: |-
  Provides a data source to read Organization VDC Templates from Viettel IDC Cloud.
---

# vcd\_org\_vdc\_template

Provides a data source to read Organization VDC Templates from Viettel IDC Cloud.
Can be used by System Administrators or tenants, only if the template is published in that tenant.

Supported in provider *v3.13+*

-> VDC Templates that do not use NSX-T can be read, but may be missing some details.

## Example Usage

```hcl
data "vcloud_org_vdc_template" "template" {
  name = "myTemplate"
}
```

-> Note that when using the data source as a tenant, `name` is the VDC Template name as seen by tenants

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the existing Organization VDC Template to read

## Attribute Reference

All the arguments from [the `vcloud_org_vdc_template` resource](/providers/vmware/vcd/latest/docs/resources/org_vdc_template) are available as read-only.
If you use this data source as a tenant, the attribute `readable_by_org_ids` must not be used (it will be always empty).
