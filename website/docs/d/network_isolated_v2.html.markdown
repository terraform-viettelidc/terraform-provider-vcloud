---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_network_isolated_v2"
sidebar_current: "docs-vcloud-data-source-network-isolated-v2"
description: |-
  Provides a Viettel IDC Cloud Org VDC isolated Network data source to read data or reference existing network.
---

# vcloud\_network\_isolated\_v2

Provides a Viettel IDC Cloud Org VDC isolated Network data source to read data or reference existing network.

Supported in provider *v3.2+* for both NSX-T and NSX-V VDCs.

## Example Usage (Looking up Isolated Network in VDC)

```hcl
data "vcloud_org_vdc" "main" {
  org  = "my-org"
  name = "main-edge"
}

data "vcloud_network_isolated_v2" "net" {
  org      = "my-org"
  owner_id = data.vcloud_org_vdc.main.id
  name     = "my-net"
}
```

## Example Usage (Looking up Isolated Network in VDC Group)

```hcl
data "vcloud_vdc_group" "main" {
  org  = "my-org"
  name = "main-group"
}

data "vcloud_network_isolated_v2" "net" {
  org      = "my-org"
  owner_id = data.vcloud_vdc_group.main.id
  name     = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `owner_id` - (Optional) VDC or VDC Group ID. Always takes precedence over `vdc` fields (in resource
and inherited from provider configuration)
* `vdc` - (Deprecated; Optional) The name of VDC to use. **Deprecated**  in favor of new field
  `owner_id` which supports VDC and VDC Group IDs.
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional) Retrieves the data source using one or more filter parameters. **Note**
  filters do not support searching for networks in VDC Groups.

## Attribute reference

All attributes defined in [isolated network resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/network_isolated_v2#attribute-reference) are supported.

## Filter arguments

* `name_regex` - (Optional) matches the name using a regular expression.
* `ip` - (Optional) matches the IP of the resource using a regular expression.

See [Filters reference](/providers/terraform-viettelidc/vcloud/latest/docs/guides/data_source_filters) for details and examples.
