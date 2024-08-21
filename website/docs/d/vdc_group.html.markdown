---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vdc_group"
sidebar_current: "docs-vcd-data-source-vdc-group"
description: |-
  Provides a data source to read VDC groups.
---

# vcd\_vdc\_group
Supported in provider *v3.5+* and VCD 10.2+.

Provides a data source to read NSX-T VDC group and reference in other resources.

~> Only `System Administrator` and `Org Users` with right `View VDC Group` can access VDC groups using this data source.

## Example Usage

```hcl
data "vcloud_vdc_group" "startVdc" {
  org  = "myOrg"
  name = "myVDC"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional)  - Name of VDC group
* `id` - (Optional)  - ID of VDC group

Either `name` or `id` must be used. If both are missing, an error arises.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_vdc_group`](/providers/vmware/vcd/latest/docs/resources/vdc_group) resource are available.