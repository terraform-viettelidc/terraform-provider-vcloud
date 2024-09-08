---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_network_dhcp_binding"
sidebar_current: "docs-vcd-datasource-nsxt-network-dhcp-binding"
description: |-
  Provides a data source to read NSX-T Org VDC network DHCP bindings.
---

# vcd\_nsxt\_network\_dhcp\_binding

Provides a data source to read NSX-T Org VDC network DHCP bindings.

-> This data source requires VCD 10.3.1+

## Example Usage

```hcl
data "vcloud_nsxt_network_dhcp" "pools" {
  org = "cloud"

  org_network_id = data.vcloud_network_routed_v2.parent.id
}
data "vcloud_nsxt_network_dhcp_binding" "binding1" {
  org            = "cloud"
  org_network_id = data.vcloud_nsxt_network_dhcp.pools.id
  name           = "Binding-one"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization. Optional if defined at provider level
* `org_network_id` - (Required) The ID of an Org VDC network. **Note**  `.id` field of
  `vcloud_network_isolated_v2`, `vcloud_network_routed_v2` or `vcloud_nsxt_network_dhcp` can be referenced
  here
* `name` - (Required) A name of DHCP binding

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_network_dhcp_binding`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_dhcp_binding)
resource are available.
