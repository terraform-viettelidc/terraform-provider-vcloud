---
layout: "vcd"
page_title: "Viettel IDC Cloud: VDC Groups"
sidebar_current: "docs-vcd-guides-vdc-groups"
description: |-
  Provides guidance to VDC Group support
---

# VDC Groups in Viettel IDC Cloud 10.2+

## About 

Starting with version 10.2, Viettel IDC Cloud supports data center group networking backed by
NSX-T Data Center.

To create a network across multiple organization VDCs, you first group the VDCs and then create a
group network that is shared with them.

Data center group networks backed by NSX-T Data Center provide level-2 network sharing, single
active egress point configuration, and distributed firewall (DFW) rules that are applied across a
data center group.

### Data center group 

A data center group acts as a cross-VDC router that provides centralized networking administration,
egress point configuration, and east-west traffic between all networks within the group. A data
center group can contain between one and 16 VDCs that you configure to share an active egress point. 

### Availability zone 

An availability zone represents the compute clusters or compute fault domains that are available to
the network. By default, the availability zone is the provider VDC. 

### Egress point 

An existing NSX-T Data Center Edge Gateway that you configure to connect a data center group to an
external network. 

## Requirements

VDC Group support requires:

* Terraform Provider VCLOUD 3.6+
* Viettel IDC Cloud 10.2+

-> For changed fields (these are usually `vdc` and `owner_id`) the previous behavior is deprecated,
but still supported. To use VDC Groups though, one needs to migrate to new configuration, which
shouldn't require rebuilding infrastructure.

## Terraform Provider VCLOUD support 

This document describes features that were introduced in Terraform Provider VCLOUD 3.6.0+ for VDC Group
support. Earlier versions of Terraform Provider VCLOUD do not support VDC Groups.

Major new approach for VDC Group support is the use of new field `owner_id` (except for routed
network, which inherits parent VDC/VDC Group from Edge Gateway) field instead of `vdc`. `owner_id`
field **always takes precedence** above `vdc` field in resource and inherited from `provider`
section.

### List of resources that support VDC Groups (NSX-T only)

The following list of resources (and their corresponding data sources) support NSX-T VDC Groups (no
NSX-V VDC Group support is provided):

* [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_edgegateway)
* [vcloud_network_routed_v2](/providers/terraform-viettelidc/vcloud/latest/docs/resources/network_routed_v2)
* [vcloud_network_isolated_v2](/providers/terraform-viettelidc/vcloud/latest/docs/resources/network_isolated_v2)
* [vcloud_nsxt_network_imported](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_imported)
* [vcloud_nsxt_ip_set](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_ip_set)
* [vcloud_nsxt_app_port_profile](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_app_port_profile)
* [vcloud_nsxt_security_group](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_security_group)
* [vcloud_nsxt_network_dhcp](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_dhcp)
* [vcloud_nsxt_distributed_firewall](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_distributed_firewall)
* [vcloud_nsxt_network_context_profile](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/nsxt_network_context_profile)
* [vcloud_nsxt_firewall](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_firewall)
* [vcloud_nsxt_nat_rule](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_nat_rule)
* [vcloud_nsxt_ipsec_vpn_tunnel](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_ipsec_vpn_tunnel)
* [vcloud_nsxt_alb_settings](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_settings)
* [vcloud_nsxt_alb_edgegateway_service_engine_group](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_pool)
* [vcloud_nsxt_alb_virtual_service](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_virtual_service)
* [vcloud_nsxt_alb_pool](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_alb_pool)

The next sub-sections will cover some specifics for resources that have it. Resources that are not
explicitly mentioned here simply introduce `owner_id` field over deprecated `vdc` field.

#### Resource vcloud_nsxt_edgegateway

New fields for handling both VDCs and VDC Groups:

* `owner_id` (replaces deprecated `vdc` field in resource and inherited from provider
  configuration). This field now supports both - VDC and VDC Group IDs. 
* `starting_vdc_id` is an optional field and is only useful if `owner_id` is a VDC Group. NSX-T Edge
  Gateway cannot be created directly in VDC Group - at first it must originate in a VDC (which is a
  member of destination VDC Group). The initial VDC defines Egress point for traffic and picking
  right VDC might be important when VDC Group spans multiple availability zones in different
  locations. When this field is not specified, a random member of destination VDC Group will be
  picked for Edge Gateway creation and then immediately moved to VDC Group as specified in
  `owner_id`.

#### Resource vcloud_network_routed_v2

Terraform Provider VCLOUD 3.6.0 changes behavior of `vcloud_network_routed_v2` resource. It __does not
require__ to specify `vdc` or `owner_id` fields. Instead, it inherits VDC or VDC Group membership
directly from parent Edge Gateway (specified in `edge_gateway_id`). The reason for this is that
routed Org VDC networks travel to and from VDC Groups with parent Edge Gateway and this does not
work well with Terraform concept.

#### Resource vcloud_nsxt_app_port_profile

NSX-T Application Port Profiles that can be used in regular and Distributed Firewalls can be defined
in multiple contexts - VDC, VDC Group and NSX-T Manager (network provider). This resource introduced
a new field `context_id` which accepts IDs for mentioned entities. 

Scope of Application Port Profiles can be one of `SYSTEM`, `TENANT` or `PROVIDER`. UI behaves a bit
differently and it has only two views - "Default Applications" and "Custom Applications". "Default
Applications" are the `SYSTEM` scoped ones, while "Custom Applications" show `TENANT` and `PROVIDER`
scoped applications.

In UI it also does not matter if the Application Port Profile is created in NSX-T Edge Gateway or
VDC Group - they are still shown in both views. 

#### Resource vcloud_nsxt_network_dhcp

`vcloud_nsxt_network_dhcp` continues to work how it worked before VDC Group rollout. It still requires
correct `org` and `vdc` fields. The `vdc` field must have a name of any `vdc` in the VDC Group.

## Complete example for configuration with VDC Groups

```hcl
variable "org_name" {
  type = string
}

variable "vdc_name" {
  type = string
}

variable "external_network_name" {
  type = string
}

variable "nsxt_segment_name" {
  type = string
}

variable "vdc_group_name" {
  type = string
}

data "vcloud_vdc_group" "main" {
  org  = var.org_name
  name = var.vdc_group_name
}

data "vcloud_external_network_v2" "nsxt-ext-net" {
  name = var.external_network_name
}

resource "vcloud_nsxt_edgegateway" "nsxt-edge" {
  org      = var.org_name
  owner_id = data.vcloud_vdc_group.main.id
  name     = "nsxt-edge-gateway"

  external_network_id = data.vcloud_external_network_v2.nsxt-ext-net.id

  subnet {
    gateway       = "10.10.10.253"
    prefix_length = "24"
    primary_ip    = "10.10.10.138"
    allocated_ips {
      start_address = "10.10.10.138"
      end_address   = "10.10.10.142"
    }
  }
}

resource "vcloud_network_routed_v2" "nsxt-backed" {
  org  = var.org_name
  name = "nsxt-routed-net-1"

  edge_gateway_id = vcloud_nsxt_edgegateway.nsxt-edge.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcloud_nsxt_network_dhcp" "pools" {
  vdc            = var.vdc_name
  org_network_id = vcloud_network_routed_v2.nsxt-backed.id

  pool {
    start_address = "1.1.1.111"
    end_address   = "1.1.1.112"
  }

  pool {
    start_address = "1.1.1.211"
    end_address   = "1.1.1.212"
  }
}

resource "vcloud_network_isolated_v2" "nsxt-backed" {
  org      = var.org_name
  owner_id = data.vcloud_vdc_group.main.id

  name = "nsxt-isolated-1"

  gateway       = "2.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "2.1.1.10"
    end_address   = "2.1.1.20"
  }
}

resource "vcloud_nsxt_network_imported" "nsxt-backed" {
  org      = var.org_name
  owner_id = data.vcloud_vdc_group.main.id

  name = "nsxt-imported-network"

  nsxt_logical_switch_name = var.nsxt_segment_name

  gateway       = "4.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "4.1.1.10"
    end_address   = "4.1.1.20"
  }
}

resource "vcloud_nsxt_app_port_profile" "custom" {
  org  = var.org_name
  name = "custom_app_prof"

  context_id = data.vcloud_vdc_group.main.id

  description = "Application port profile for custom"
  scope       = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }
}

resource "vcloud_nsxt_ip_set" "ipset1" {
  org             = var.org_name
  edge_gateway_id = vcloud_nsxt_edgegateway.nsxt-edge.id

  name         = "local-gateway"
  ip_addresses = ["10.1.1.1"]
}

resource "vcloud_nsxt_security_group" "net" {
  edge_gateway_id = vcloud_nsxt_edgegateway.nsxt-edge.id

  name        = "routed-network"
  description = "Security Group containing routed network members"

  member_org_network_ids = [vcloud_network_routed_v2.nsxt-backed.id]
}

data "vcloud_nsxt_network_context_profile" "av" {
  context_id = data.vcloud_vdc_group.main.id
  name       = "360ANTIV"
}

resource "vcloud_nsxt_distributed_firewall" "t1" {
  org          = var.org_name
  vdc_group_id = data.vcloud_vdc_group.main.id

  rule {
    name    = "rule1"
    action  = "ALLOW"
    comment = "Accept "

    source_ids      = [vcloud_nsxt_security_group.net.id]
    destination_ids = [vcloud_nsxt_ip_set.ipset1.id]

    network_context_profile_ids = [data.vcloud_nsxt_network_context_profile.av.id]
  }

  rule {
    name        = "rule5"
    action      = "ALLOW"
    ip_protocol = "IPV6"
    direction   = "IN"

    app_port_profile_ids = [vcloud_nsxt_app_port_profile.custom.id]
  }

  rule {
    name        = "drop all"
    action      = "DROP"
    ip_protocol = "IPV4"
  }
}
```
## References

* [Viettel IDC Cloud Documentation about VDC
  Groups](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-E8A8CD70-31AD-4592-B520-34E3B7DC4E6E.html)
