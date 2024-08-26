---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_ip_space_uplink"
sidebar_current: "docs-vcloud-datasource-ip-space-uplink"
description: |-
  Provides a data source to read IP Space Uplinks in External Networks (Provider Gateways).
---

# vcloud\_ip\_space\_uplink

Provides a data source to read IP Space Uplinks in External Networks (Provider Gateways).

## Example Usage

```hcl
data "vcloud_ip_space_uplink" "u1" {
  name                = "ip-space-uplink-1"
  external_network_id = vcloud_external_network_v2.provider-gateway.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) - Name of IP Space Uplink
* `external_network_id` - (Required) Parent External Network ID

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_ip_space_uplink`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/ip_space_uplink) resource are available.
