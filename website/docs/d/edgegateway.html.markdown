---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_edgegateway"
sidebar_current: "docs-vcd-data-source-edgegateway"
description: |-
  Provides an NSX-V edge gateway data source.
---

# vcd\_edgegateway

Provides a Viettel IDC Cloud NSX-V edge gateway data source, directly connected to one or more external networks. This can be used to reference
edge gateways for Org VDC networks to connect.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcloud_edgegateway" "mygw" {
  name = "mygw"
  org  = "myorg"
  vdc  = "myvdc"
}

output "edge_gateway_id" {
  value = data.vcloud_edgegateway.mygw.id
}

# Get the name of the default gateway from the data source
# and use it to establish a second data source
data "vcloud_external_network" "external_network1" {
  name = data.vcloud_edgegateway.mygw.external_network.name
}

# From the second data source we extract the basic networking info
output "gateway" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.gateway
}
output "netmask" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.netmask
}
output "DNS" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.dns1
}
output "external_ip" {
  value = data.vcloud_external_network.external_network1.ip_scope.0.static_ip_pool.0.start_address
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the edge gateway (optional when `filter` is used)
* `org` - (Optional) The name of organization to which the VDC belongs. Optional if defined at provider level.
* `vdc` - (Optional) The name of VDC that owns the edge gateway. Optional if defined at provider level. 
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute Reference

All attributes defined in [edge gateway resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/edgegateway#attribute-reference) are supported.

## Filter arguments

(Supported in provider *v2.9+*)

* `name_regex` - (Optional) matches the name using a regular expression.

See [Filters reference](/providers/terraform-viettelidc/vcloud/latest/docs/guides/data_source_filters) for details and examples.

