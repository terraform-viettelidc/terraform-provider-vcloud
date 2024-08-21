---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_external_network_v2"
sidebar_current: "docs-vcd-resource-external-network-v2"
description: |-
  Provides a Viettel IDC Cloud External Network resource (version 2). New version of this resource
  uses new VCD API and is capable of creating NSX-T backed external networks as well as port group
  backed ones.
---

# vcd\_external\_network\_v2

Provides a Viettel IDC Cloud External Network resource (version 2). New version of this resource 
uses new VCD API and is capable of creating NSX-T backed external networks as well as port group
backed ones.

-> This resource manages NSX-T **External Networks**, NSX-V **External Networks**, and **NSX-T
Provider Gateways**

This resource supports **IP Spaces** - read [IP Spaces guide
page](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/ip_spaces) for more
information.

## Example Usage (NSX-T Tier 0 Router backed External Network backed by IP Spaces)

```hcl
resource "vcloud_external_network_v2" "ext-net-nsxt-t0" {
  name        = "nsxt-external-network"
  description = "IP Space backed"

  nsxt_network {
    nsxt_manager_id      = data.vcloud_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcloud_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
  # optional argument to dedicate network to a particular Org
  dedicated_org_id = data.vcloud_org.org1.id
}
```

## Example Usage (NSX-T Tier 0 Router backed External Network)

```hcl
data "vcloud_nsxt_manager" "main" {
  name = "nsxManager"
}

data "vcloud_nsxt_tier0_router" "router" {
  name            = "first-router"
  nsxt_manager_id = data.vcloud_nsxt_manager.main.id
}

resource "vcloud_external_network_v2" "ext-net-nsxt-t0" {
  name        = "nsxt-external-network"
  description = "First NSX-T Tier 0 router backed network"

  nsxt_network {
    nsxt_manager_id      = data.vcloud_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcloud_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = false
    gateway       = "88.88.88.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }

  ip_scope {
    # enabled       = true # by default
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }

    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}
```

## Example Usage (NSX-T Segment backed External Network with a Direct Org VDC network)

-> NSX-T **Segment backed External Network** is similar to **Imported Org VDC network**. The difference is that
**External Network can consume one NSX-T Segment and then many VDCs can use it by using NSX-T Direct Network**, 
while Org VDC Imported network directly requires one NSX-T Segment

```hcl
data "vcloud_nsxt_manager" "main" {
  name = "nsxManager"
}

resource "vcloud_external_network_v2" "ext-net-nsxt-segment" {
  name        = "nsxt-external-network"
  description = "First NSX-T segment backed network"

  nsxt_network {
    nsxt_manager_id   = data.vcloud_nsxt_manager.main.id
    nsxt_segment_name = "existing-nsxt-segment"
  }

  ip_scope {
    enabled       = false
    gateway       = "88.88.88.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }

  ip_scope {
    # enabled       = true # by default
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }

    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}

resource "vcloud_network_direct" "net" {
  vdc = "nsxt-vdc"

  name             = "direct-net"
  external_network = vcloud_external_network_v2.ext-net-nsxt-segment.name

  depends_on = [vcloud_external_network_v2.ext-net-nsxt]
}
```

## Example Usage (NSX-V backed external network)
```hcl
data "vcloud_vcenter" "vc" {
  name = "vc1"
}

data "vcloud_portgroup" "sw" {
  name = "TestbedPG"
  type = "DV_PORTGROUP"
}

resource "vcloud_external_network_v2" "ext-net-nsxv" {
  name        = "nsxv-external-network"
  description = "NSX-V based external network"

  vsphere_network {
    vcenter_id   = data.vcloud_vcenter.vc.id
    portgroup_id = data.vcloud_portgroup.sw.id
  }

  ip_scope {
    gateway       = "192.168.30.49"
    prefix_length = "24"
    dns1          = "192.168.0.164"
    dns2          = "192.168.0.196"
    dns_suffix    = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the network
* `description` - (Optional) Network friendly description
* `use_ip_spaces` - (Optional; *v3.10+*; *VCD 10.4.1+*) Defines if the network uses IP Spaces. Do
  not specify `ip_scope` when using IP Spaces. (default `false`)
* `dedicated_org_id` - (Optional; *v3.10+*; *VCD 10.4.1+*) An Org ID that this network should be
  dedicated to. Only applicable when `use_ip_spaces=true`
* `ip_scope` - (Optional) One or more IP scopes for the network. See [IP Scope](#ipscope) below for details.
* `vsphere_network` - (Optional) One or more blocks of [vSphere Network](#vspherenetwork)..
* `nsxt_network` - (Optional) NSX-T network definition. See [NSX-T Network](#nsxtnetwork) below for details.

<a id="ipscope"></a>
## IP Scope

* `gateway` - (Required) Gateway of the network.
* `prefix_length` - (Required) Network prefix.
* `static_ip_pool` - (Required) IP ranges used for static pool allocation in the network.  See [IP Pool](#ip-pool) below for details.
* `dns1` - (Optional) Primary DNS server. Provider version *v3.9+* also supports NSX-T
* `dns2` - (Optional) Secondary DNS server. Provider version *v3.9+* also supports NSX-T
* `dns_suffix` - (Optional) A FQDN for the virtual machines on this network. Provider version
  *v3.9+* also supports NSX-T
* `enabled` - (Optional) Default is `true`.

<a id="ip-pool"></a>
## IP Pool

* `start_address` - (Required) Start address of the IP range
* `end_address` - (Required) End address of the IP range

<a id="vspherenetwork"></a>
## vSphere Network

* `vcenter_id` - (Required) vCenter ID. Can be looked up using [`vcloud_vcenter`](/providers/vmware/vcd/latest/docs/data-sources/vcenter) data source.
* `portgroup_id` - (Required) vSphere portgroup ID. Can be looked up using  [`vcloud_portgroup`](/providers/vmware/vcd/latest/docs/data-sources/portgroup) data source.

<a id="nsxtnetwork"></a>
## NSX-T Network

* `nsxt_manager_id` - (Required) NSX-T manager ID. Can be looked up using [`vcloud_nsxt_manager`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_manager) data source.
* `nsxt_tier0_router_id` - (Optional) NSX-T Tier-0 router ID. Can be looked up using
  [`vcloud_nsxt_tier0_router`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_tier0_router) data source.
* `nsxt_segment_name` - (Optional; *v3.4+*; *VCD 10.3+*) Existing NSX-T segment name.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing external network can be [imported][docs-import] into this resource via supplying the path for an external network. Since the external network is
at the top of the vCD hierarchy, the path corresponds to the external network name.
For example, using this structure, representing an existing external network that was **not** created using Terraform:

```hcl
resource "vcloud_external_network_v2" "tf-external-network" {
  name = "my-ext-net"
}
```

You can import such external network into terraform state using this command

```
terraform import vcloud_external_network_v2.tf-external-network my-ext-net
```

[docs-import]:https://www.terraform.io/docs/import/

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR

While the above structure is the minimum needed to get an import, it is not sufficient to run `terraform plan`,
as it lacks several mandatory fields. To use the imported resource, you will need to add the missing properties
using the data in `terraform.tfstate` as a reference. If the resource does not need modifications, consider using
an [external network data source](/providers/vmware/vcd/latest/docs/data-sources/external_network_v2) instead. 
