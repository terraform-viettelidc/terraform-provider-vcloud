---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vcenter"
sidebar_current: "docs-vcd-data-source-vcenter"
description: |-
  Provides a data source for vCenter server attached to VCD.
---

# vcloud\_vcenter

Provides a data source for vCenter server attached to VCD.

Supported in provider *v3.0+*


## Example Usage

```hcl
data "vcloud_vcenter" "vc" {
  name = "vcenter-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) vCenter name

## Attribute reference

* `connection_status` -  vCenter connection status (e.g. `CONNECTED`). 
* `is_enabled` -  Boolean value if vCenter is enabled.
* `status` -  vCenter status (e.g. `READY`).
* `vcenter_host` -  Hostname of configured vCenter.
* `vcenter_version` -  vCenter version (e.g. `6.7.0`)
