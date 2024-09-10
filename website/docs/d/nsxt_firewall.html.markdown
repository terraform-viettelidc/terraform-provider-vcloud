---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_firewall"
sidebar_current: "docs-vcd-resource-nsxt-firewall"
description: |-
  Provides a data source to read NSX-T Firewall configuration of an Edge Gateway. Firewalls allow 
  user to control the incoming and outgoing network traffic to and from an NSX-T Data Center 
  Edge Gateway.
---

# vcloud\_nsxt\_firewall

Supported in provider *v3.3+* and VCLOUD 10.1+ with NSX-T backed Edge Gateways.

Provides a data source to read NSX-T Firewall configuration of an Edge Gateway. Firewalls allow 
user to control the incoming and outgoing network traffic to and from an NSX-T Data Center 
Edge Gateway.


## Example Usage (Read a list of firewall rules on Edge Gateway)
```hcl
data "vcloud_nsxt_firewall" "testing" {
  org = "my-org"

  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organizations.
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using
  `vcloud_nsxt_edgegateway` data source

## Attribute reference

All properties defined in [vcloud_nsxt_firewall](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_firewall)
resource are available.
