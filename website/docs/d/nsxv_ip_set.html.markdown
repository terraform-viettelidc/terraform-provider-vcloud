---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxv_ip_set"
sidebar_current: "docs-vcd-datasource-ipset"
description: |-
  Provides an IP set data source.
---

# vcd\_nsxv\_ip\_set

Provides a Viettel IDC Cloud IP set data source. An IP set is a group of IP addresses that you can add
  as the source or destination in a firewall rule or in DHCP relay configuration.

Supported in provider *v2.6+*

## Example Usage

```hcl
data "vcloud_nsxv_ip_set" "ip-set-DS" {
  org = "my-org"
  vdc = "my-org-vdc"

  name = "not-managed"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) IP set name for identifying the exact IP set

## Attribute Reference

All the attributes defined in [`vcloud_nsxv_ip_set`](/providers/vmware/vcd/latest/docs/resources/nsxv_ip_set) resource are available.
