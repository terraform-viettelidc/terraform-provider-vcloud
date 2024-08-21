---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edgegateway_dhcp_forwarding"
sidebar_current: "docs-vcloud-resource-nsxt-edge-dhcp-forwarding"
description: |-
  Provides a resource to manage NSX-T Edge Gateway DHCP forwarding configuration.
---

# vcloud\_nsxt\_edgegateway\_dhcp\_forwarding

Supported in provider *v3.10+* and Vcloud 10.3.1+ with NSX-T.

Provides a resource to manage NSX-T Edge Gateway DHCP forwarding configuration.

## Example Usage

```hcl
data "vcloud_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcloud_nsxt_edgegateway" "testing-in-vdc" {
  org      = "datacloud"
  owner_id = data.vcloud_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

resource "vcloud_nsxt_edgegateway_dhcp_forwarding" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing-in-vdc.id

  enabled = true

  dhcp_servers = [
    "192.168.0.13",
    "fe80::aaaa",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Org in which the NSX-T Edge Gateway is located, required
  if not set in the provider section.
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID.
* `enabled` - (Required) DHCP Forwarding status.
* `dhcp_servers` - (Required) IP addresses of DHCP servers. Maximum 8 can be specified.

~> Modification of the `dhcp_servers` field will not be changed in Vcloud when `enabled = false` because Vcloud API ignores DHCP server changes when DHCP forwarding is disabled.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)


An existing Edge Gateway DHCP forwarding configuration can be 
[imported][docs-import] into a resource via supplying the 
full dot separated path. For example: 

```hcl
resource "vcloud_nsxt_edgegateway_dhcp_forwarding" "imported" {
  org             = "my-org"
  edge_gateway_id = vcloud_nsxt_edgegateway.nsxt-edge.id

  enabled = true
  dhcp_servers = [
    "192.168.0.2",
  ]
}
```

You can import such configuration into terraform state using this command
```
terraform import vcloud_nsxt_edgegateway_dhcp_forwarding.imported my-org.nsxt-vdc.nsxt-edge
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR

The above would import the `nsxt-edge` Edge Gateway DHCP forwarding configuration for this particular
Edge Gateway.

[docs-import]: https://www.terraform.io/docs/import/
