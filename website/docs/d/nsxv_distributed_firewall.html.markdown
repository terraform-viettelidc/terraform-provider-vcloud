---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxv_distributed_firewall"
sidebar_current: "docs-vcd-data-source-nsxv-distributed-firewall"
description: |-
  The NSX-V Distributed Firewall data source reads all defined rules for a particular VDC
---

# vcd\_nsxv\_distributed\_firewall

The NSX-V Distributed Firewall data source reads all defined rules for a particular VDC.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcloud_org_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcloud_nsxv_distributed_firewall" "dfw1" {
  vdc_id = data.vcloud_org_vdc.my-vdc.id
}
```

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to manage the Distributed Firewall in. Can be looked up using a `vcloud_org_vdc` data source

## Attributes reference

All the arguments and attributes defined for the `vcloud_nsxv_distributed_firewall` resource are reported as attributes for this data source.