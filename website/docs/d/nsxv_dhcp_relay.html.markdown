---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxv_dhcp_relay"
sidebar_current: "docs-vcd-datasource-nsxv-dhcp-relay"
description: |-
  Provides an NSX edge gateway DHCP relay configuration data source.
---

# vcd\_nsxv\_dhcp\_relay

Provides a Viettel IDC Cloud Edge Gateway DHCP relay configuration data source. The DHCP relay
capability provided by NSX in Viettel IDC Cloud environment allows to leverage existing DHCP
infrastructure from within Viettel IDC Cloud environment without any interruption to the IP address
management in existing DHCP infrastructure. DHCP messages are relayed from virtual machines to the
designated DHCP servers in your physical DHCP infrastructure, which allows IP addresses controlled
by the NSX software to continue to be in sync with IP addresses in the rest of your DHCP-controlled
environments. 

Supported in provider *v2.6+*

## Example Usage 1

```hcl
data "vcloud_nsxv_dhcp_relay" "relay_config" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which DHCP relay is to be configured.

## Attribute Reference

All the attributes defined in [`vcloud_nsxv_dhcp_relay`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxv_dhcp_relay)
resource are available.
