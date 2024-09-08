---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_lb_app_profile"
sidebar_current: "docs-vcd-data-source-lb-app-profile"
description: |-
  Provides an NSX edge gateway load balancer application profile data source.
---

# vcd\_lb\_app\_profile

Provides a Viettel IDC Cloud Edge Gateway Load Balancer Application Profile data source. An
application profile defines the behavior of the load balancer for a particular type of network
traffic. After configuring a profile, you associate it with a virtual server. The virtual server
then processes traffic according to the values specified in the profile.

~> **Note:** See additional support notes in [application profile resource page]
(/providers/terraform-viettelidc/vcloud/latest/docs/resources/lb_app_profile).

Supported in provider *v2.4+*

## Example Usage

```hcl
data "vcloud_lb_app_profile" "my-profile" {
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
* `name` - (Required) Application profile name for identifying the exact application profile

## Attribute Reference

All the attributes defined in `vcloud_lb_app_profile` resource are available.
