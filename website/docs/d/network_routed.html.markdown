---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_network_routed"
sidebar_current: "docs-vcd-data-source-network-routed"
description: |-
  Provides a Viettel IDC Cloud Org VDC routed Network. This can be used to reference internal networks for vApps to connect.
---

# vcd\_network\_routed

Provides a Viettel IDC Cloud Org VDC routed Network data source. This can be used to reference internal networks for vApps to connect.

Supported in provider *v2.5+*

~> **Note:** This data source supports only NSX-V backed Org VDC networks.
Please use newer [`vcloud_network_routed_v2`](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/network_routed_v2)
data source which is compatible with NSX-T.

## Example Usage

```hcl
data "vcloud_network_routed" "net" {
  org  = "my-org" # Optional
  vdc  = "my-vdc" # Optional
  name = "my-net"
}

output "edge_gateway" {
  value = data.vcloud_network_routed.net.edge_gateway
}

output "gateway" {
  value = data.vcloud_network_routed.net.gateway
}
output "dhcp_start_address" {
  value = tolist(data.vcloud_network_routed.net.dhcp_pool)[0].start_address
}

output "dhcp_end_address" {
  value = tolist(data.vcloud_network_routed.net.dhcp_pool)[0].end_address
}

output "static_ip_start_address" {
  value = tolist(data.vcloud_network_routed.net.static_ip_pool)[0].start_address
}

output "static_ip_end_address" {
  value = tolist(data.vcloud_network_routed.net.static_ip_pool)[0].end_address
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [routed network resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/network_routed#attribute-reference) are supported.

## Filter arguments

(Supported in provider *v2.9+*)

* `name_regex` - (Optional) matches the name using a regular expression.
* `ip` - (Optional) matches the IP of the resource using a regular expression.
* `metadata` - (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/terraform-viettelidc/vcloud/latest/docs/guides/data_source_filters) for details and examples.

