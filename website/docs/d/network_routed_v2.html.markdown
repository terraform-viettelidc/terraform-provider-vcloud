---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_network_routed_v2"
sidebar_current: "docs-vcloud-data-source-network-routed-v2"
description: |-
  Provides a Viettel IDC Cloud Org VDC routed Network data source to read data or reference  existing network
  (backed by NSX-T or NSX-V).
---

# vcloud\_network\_routed\_v2

Provides a Viettel IDC Cloud Org VDC routed Network data source to read data or reference  existing network
(backed by NSX-T or NSX-V).

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

## Example Usage

```hcl
data "vcloud_nsxt_edgegateway" "main" {
  org  = "my-org"
  name = "main-edge"
}

data "vcloud_network_routed_v2" "net" {
  org             = "my-org" # Optional
  edge_gateway_id = data.vcloud_nsxt_edgegateway.main.id
  name            = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `edge_gateway_id` - (Optional; *v3.6+*) Replaces `vdc` field and helps to identify exact Org
  Network
* `vdc` - (Deprecated; Optional) The name of VDC to use, optional if defined at provider level. **Deprecated**
  in favor of `edge_gateway_id` field.
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional) Retrieves the data source using one or more filter parameters. **Note**
  filters do not support searching for networks in VDC Groups.

## Attribute reference

* `owner_id` - Parent VDC or VDC Group ID.

All attributes defined in [routed network v2
resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/network_routed_v2#attribute-reference) are
supported.

## Filter arguments

* `name_regex` - (Optional) matches the name using a regular expression.
* `ip` - (Optional) matches the IP of the resource using a regular expression.

See [Filters reference](/providers/terraform-viettelidc/vcloud/latest/docs/guides/data_source_filters) for details and examples.
