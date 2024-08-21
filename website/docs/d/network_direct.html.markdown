---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_network_direct"
sidebar_current: "docs-vcloud-data-source-network-direct"
description: |-
  Provides a Viettel IDC Cloud Org VDC Network attached to an external one. This can be used to reference internal networks for vApps to connect.
---

# vcloud\_network\_direct

Provides a Viettel IDC Cloud Org VDC Network data source directly connected to an external network. This can be used to reference
internal networks for vApps to connect.

Supported in provider *v2.5+*


## Example Usage

```hcl
data "vcloud_network_direct" "net" {
  org  = "my-org"
  vdc  = "my-vdc"
  name = "my-net"
}

# Get the name of the external network from the data source
# and use it to establish a second data source
output "external_network" {
  value = data.vcloud_network_direct.net.external_network
}

data "vcloud_external_network" "external_network1" {
  name = data.vcloud_network_direct.net.external_network
}

# From the second data source we extract the basic networking info
output "gateway" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.gateway
}
# equivalent to
output "external_network_gateway" {
  value = data.vcloud_network_direct.net.external_network_gateway
}

output "netmask" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.netmask
}
# equivalent to
output "external_network_netmask" {
  value = data.vcloud_network_direct.net.external_network_netmask
}

output "DNS" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.dns1
}
# equivalent to
output "external_network_dns" {
  value = data.vcloud_network_direct.net.external_network_dns1
}

output "external_ip" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.static_ip_pool.0.start_address
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute Reference

* `external_network` -  The name of the external network.
* `shared` -  Defines if this network is shared between multiple vDCs in the vOrg.

## Filter arguments

(Supported in provider *v2.9+*)

* `name_regex` - (Optional) matches the name using a regular expression.
* `ip` - (Optional) matches the IP of the resource using a regular expression.
* `metadata` - (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/vmware/vcloud/latest/docs/guides/data_source_filters) for details and examples.
