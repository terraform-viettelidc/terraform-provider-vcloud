---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_external_network"
sidebar_current: "docs-vcd-data-source-external-network"
description: |-
  Provides an external network data source.
---

# external\_network

Provides a Viettel IDC Cloud external network data source. This can be used to reference external networks and their properties.

Supported in provider *v2.5+*

~> This resource is deprecated in favor of [`vcloud_external_network_v2`](/providers/vmware/vcd/latest/docs/data-sources/external_network_v2)

## Example Usage

```hcl
data "vcloud_external_network" "tf-external-network" {
  name = "my-extnet"
}

resource "vcloud_dnat" "tf-nat-rule" {
  org = "tf-org"
  vdc = "tf-vdc"
  # References the external network name from the data source
  network_name = data.vcloud_external_network.tf-external-network.name
  network_type = "ext"
  edge_gateway = "tf-gw"
  # References the first IP scope block. From that we extract the first static IP pool to retrieve the start address
  external_ip     = data.vcloud_external_network.extnet-datacloud.ip_scope[0].static_ip_pool[0].start_address
  port            = 7777
  protocol        = "tcp"
  internal_ip     = "10.10.102.60"
  translated_port = 77
  description     = "test run"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) external network name

## Attribute Reference

* `description` - Network friendly description
* `ip_scope` -  A list of IP scopes for the network. See [IP Scope](/providers/vmware/vcd/latest/docs/resources/external_network#ipscope)
   for details.
* `vsphere_network` -  A list of DV_PORTGROUP or NETWORK objects names that back this network. Each referenced 
  DV_PORTGROUP or NETWORK must exist on a vCenter server registered with the system.
  See [vSphere Network](/providers/vmware/vcd/latest/docs/resources/external_network#vspherenetwork) for details.
* `retain_net_info_across_deployments` -  Specifies whether the network resources such as IP/MAC of router will be 
  retained across deployments.

