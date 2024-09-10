---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_ip_space_uplink"
sidebar_current: "docs-vcd-resource-ip-space-uplink"
description: |-
  Provides a resource to manage IP Space Uplinks in External Networks (Provider Gateways).
---

# vcloud\_ip\_space\_uplink

Provides a resource to manage IP Space Uplinks in External Networks (Provider Gateways).

~> Only `System Administrator` can create this resource.

## Example Usage (Adding IP Space Uplink to Provider Gateway)

```hcl
data "vcloud_nsxt_manager" "main" {
  name = "nsxManager1"
}

data "vcloud_nsxt_tier0_router" "router" {
  name            = "tier0Router"
  nsxt_manager_id = data.vcloud_nsxt_manager.main.id
}

resource "vcloud_ip_space" "space1" {
  name = "ip-space-1"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}

resource "vcloud_external_network_v2" "provider-gateway" {
  name = "ProviderGateway1"

  nsxt_network {
    nsxt_manager_id      = data.vcloud_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcloud_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}

resource "vcloud_ip_space_uplink" "u1" {
  name                = "uplink"
  description         = "uplink number one"
  external_network_id = vcloud_external_network_v2.provider-gateway.id
  ip_space_id         = vcloud_ip_space.space1.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A tenant facing name for IP Space Uplink
* `description` - (Optional) An optional description for IP Space Uplink
* `external_network_id` - (Required) External Network ID For IP Space Uplink configuration
* `ip_space_id` - (Required) IP Space ID configuration

## Attribute Reference

The following attributes are exported on this resource:

* `ip_space_type` - Backing IP Space type
* `status` - Status of IP Space Uplink


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing IP Space Uplink configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_ip_space_uplink.imported external-network-name.ip-space-uplink-name
```

The above would import the `ip-space-uplink-name` IP Space Uplink that is set for
`external-network-name`
