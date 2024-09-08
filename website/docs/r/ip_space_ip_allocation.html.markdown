---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_ip_space_ip_allocation"
sidebar_current: "docs-vcd-resource-ip-space-ip-allocation"
description: |-
  Provides a resource to manage IP Allocations within IP Spaces. It supports both - Floating IPs 
  (IPs from IP Ranges) and IP Prefix (subnet) allocations with manual and automatic reservations.
---

# vcd\_ip\_space\_ip\_allocation

Provides a resource to manage IP Allocations within IP Spaces. It supports both - Floating IPs (IPs
from IP Ranges) and IP Prefix (subnet) allocations with manual and automatic reservations.


## Example Usage (Floating IP Usage for NAT rule)

```hcl
resource "vcloud_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcloud_org.org1.id
  ip_space_id = vcloud_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}

resource "vcloud_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "v42"
  edge_gateway_id = vcloud_nsxt_edgegateway.ip-space.id

  name      = "IP Space integration"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = vcloud_ip_space_ip_allocation.public-floating-ip.ip_address
  internal_address = "77.77.77.1"
  logging          = true
}
```

## Example Usage (Manual Floating IP reservation)

```hcl
resource "vcloud_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcloud_org.org1.id
  ip_space_id = vcloud_ip_space.space1.id
  type        = "FLOATING_IP"
  usage_state = "USED_MANUAL"
  description = "manually used floating IP"

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}
```

## Example Usage (IP Prefix)

```hcl
resource "vcloud_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcloud_org.org1.id
  ip_space_id   = vcloud_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 29

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}

resource "vcloud_network_routed_v2" "using-public-prefix" {
  org             = "v42"
  name            = "ip-space-backed-external-network"
  edge_gateway_id = vcloud_nsxt_edgegateway.ip-space.id
  gateway         = cidrhost(vcloud_ip_space_ip_allocation.public-ip-prefix.ip_address, 1)
  prefix_length   = split("/", vcloud_ip_space_ip_allocation.public-ip-prefix.ip_address)[1]

  static_ip_pool {
    start_address = cidrhost(vcloud_ip_space_ip_allocation.public-ip-prefix.ip_address, 2)
    end_address   = cidrhost(vcloud_ip_space_ip_allocation.public-ip-prefix.ip_address, 4)
  }
}
```

## Example Usage (Manual IP Prefix)

```hcl
resource "vcloud_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id        = data.vcloud_org.org1.id
  ip_space_id   = vcloud_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 30
  usage_state   = "USED_MANUAL"
  description   = "manually used IP Prefix"

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}
```

## Example Usage (Specifying explicit value on VCLOUD 10.4.2+)

```hcl
resource "vcloud_ip_space_ip_allocation" "public-floating-ip-2" {
  org_id      = data.vcloud_org.org1.id
  ip_space_id = vcloud_ip_space.space1.id
  type        = "FLOATING_IP"
  value       = "11.11.11.102"

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}

resource "vcloud_ip_space_ip_allocation" "public-ip-prefix" {
  org_id      = data.vcloud_org.org1.id
  ip_space_id = vcloud_ip_space.space1.id
  type        = "IP_PREFIX"
  value       = "10.10.10.96/29"

  depends_on = [vcloud_nsxt_edgegateway.ip-space]
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) Org ID in which the IP is allocated
* `ip_space_id` - (Required) IP Space ID to use for IP Allocations
* `type` - (Required) One of `FLOATING_IP`, `IP_PREFIX`
  * `FLOATING_IP` - allocates single IP from defined ranges in IP Space
  * `IP_PREFIX` - allocates subnets. **Note** field `prefix_length` is required to allocate IP
    Prefix
* `prefix_length` (Optional) Required when `type=IP_PREFIX`
* `value` - (Optional; VCLOUD *10.4.2+*) An option to request a specific IP or subnet from IP Space.
  **Note:** This field does not support IP ranges because it would cause multiple allocations
  created in one resource. Please use multiple resource instances to allocate IP ranges.
* `usage_state` - (Optional) (Optional) Only used with manual reservations. Value `USED_MANUAL`
  enables manual IP reservation. Value `UNUSED` is set to release manual allocation of IP.
* `description` - (Optional) Can only be set when `usage_state=USED_MANUAL`

~> IP Allocation resources can be created only if there is a NSX-T Edge Gateway
(`vcloud_nsxt_edgegateway`) that is backed by the Provider Gateway (`vcloud_external_network_v2`) with IP
Space Uplinks (`vcloud_ip_space_uplink`). Attempting to allocate IP Addresses before having an
Edge Gateway withing VDC will return errors of type `This operation is denied`.

## Attribute Reference

* `ip_address` - contains either single IP when `type=FLOATING_IP` (e.g. `192.168.1.100`) or subnet
  in CIDR format when `type=IP_PREFIX` (e.g. `192.168.1.100/30`). **Note:** Terraform built-in
  function [cidrhost](https://developer.hashicorp.com/terraform/language/functions/cidrhost) is a
  convenient method to getting IPs within returned CIDR
* `allocation_date` - allocation date in formated as `2023-06-07T09:57:58.721Z` (ISO 8601)
* `usage_state` - `USED` or `UNUSED` is populated by system unless set to `USED_MANUAL` or `UNUSED`
* `used_by_id` - contains entity ID that is using the IP if `usage_state=USED`
* `ip` - convenience field. For `type=IP_PREFIX` it will contain only the IP from CIDR returned

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An IP Space can be [imported][docs-import] into this resource via supplying path for it. An example
is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_ip_space_ip_allocation.ip org-name.ip-space-name.ip-allocation-type.ip-allocation-ip
```

e.g.

```
terraform import vcloud_ip_space_ip_allocation.ip my-org.my-ip-space.FLOATING_IP.10.10.10.1
```

`ip-allocation-type` reflects the value of field `type` and must be one of its values (`FLOATING_IP`
or `IP_PREFIX`)

The above would import the `10.10.10.1` IP Allocation of type `FLOATING_IP` in IP Space
`my-ip-space` withing Org `my-org`.
