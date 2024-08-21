---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_lb_app_rule"
sidebar_current: "docs-vcd-data-source-lb-app-rule"
description: |-
  Provides an NSX edge gateway load balancer application rule data source.
---

# vcd\_lb\_app\_rule

Provides a Viettel IDC Cloud Edge Gateway Load Balancer Application Rule data source. An application
rule allows to directly manipulate and manage IP application traffic with load balancer.

~> **Note:** See additional support notes in [application rule resource page]
(/providers/vmware/vcd/latest/docs/resources/lb_app_rule).

Supported in provider *v2.4+*

## Example Usage

```hcl
data "vcloud_lb_app_rule" "my-rule" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  name = "not-managed"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `edge_gateway` - (Required) The name of the edge gateway on which the service monitor is defined
* `name` - (Required) Application rule name for identifying the exact application rule

## Attribute Reference

All the attributes defined in `vcloud_lb_app_rule` resource are available.
