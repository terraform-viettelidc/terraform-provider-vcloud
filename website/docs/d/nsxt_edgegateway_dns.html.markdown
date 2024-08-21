---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edgegateway_dns"
sidebar_current: "docs-vcd-data-source-nsxt-edgegateway-dns"
description: |-
  Provides a data source to read NSX-T Edge Gateway DNS forwarder configuration.
---

# vcd\_nsxt\_edgegateway\_dns

Supported in provider *v3.11+* and VCD *10.4+* with NSX-T.

Provides a data source to read NSX-T Edge Gateway DNS forwarder configuration.

## Example Usage
```hcl
data "vcloud_org_vdc" "existing" {
  name = "existing-vdc"
}

data "vcloud_nsxt_edgegateway" "testing" {
  owner_id = data.vcloud_org_vdc.existing.id
  name     = "server-testing"
}

data "vcloud_nsxt_edgegateway_dns" "dns-service" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at 
  provider level. Useful when connected as sysadmin working across different organisations
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). 
  Can be looked up using [`vcloud_nsxt_edgegateway`](/providers/vmware/vcd/latest/docs/data-sources/nsxt_edgegateway) data source

## Attribute Reference

All properties defined in [vcloud_nsxt_edgegateway_dns](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway_dns)
resource are available.

