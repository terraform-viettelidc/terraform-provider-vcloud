---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_ipsec_vpn_tunnel"
sidebar_current: "docs-vcd-data-source-nsxt-ipsec-vpn-tunnel"
description: |-
  Provides a data source to read NSX-T IPsec VPN Tunnel. You can configure site-to-site connectivity between an NSX-T Data
  Center Edge Gateway and remote sites. The remote sites must use NSX-T Data Center, have third-party hardware routers,
  or VPN gateways that support IPSec.
---

# vcd\_nsxt\_ipsec\_vpn\_tunnel

Supported in provider *v3.3+* and VCD 10.1+ with NSX-T backed VDCs.

Provides a data source to read NSX-T IPsec VPN Tunnel. You can configure site-to-site connectivity between an NSX-T Data
Center Edge Gateway and remote sites. The remote sites must use NSX-T Data Center, have third-party hardware routers,
or VPN gateways that support IPSec.

## Example Usage

```hcl
data "vcloud_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "my-org"

  edge_gateway_id = data.vcloud_nsxt_edgegateway.existing.id

  name = "tunnel-1"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using `vcloud_nsxt_edgegateway`
  data source
* `name` - (Required)  - Name of existing IPsec VPN Tunnel

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_ipsec_vpn_tunnel`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_ipsec_vpn_tunnel) resource are available.
