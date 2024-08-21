---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_external_network_v2"
sidebar_current: "docs-vcloud-data-source-external-network-v2"
description: |-
  Provides a Viettel IDC Cloud External Network data source (version 2). New version of this data source
  uses new Vcloud API and is capable of creating NSX-T backed external networks as well as port group
  backed ones.
---

# vcloud\_external\_network\_v2

Provides a Viettel IDC Cloud External Network data source (version 2). New version of this data source uses new Vcloud
API and is capable of handling NSX-T backed external networks as well as port group backed ones.

-> **Note:** This resource uses new Viettel IDC Cloud
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least Vcloud *10.0+*. It supports both NSX-T and NSX-V backed networks (NSX-T *3.0+* requires Vcloud *10.1.1+*)

Supported in provider *v3.0+*.

## Example Usage

```hcl
data "vcloud_external_network_v2" "ext_net" {
  name = "my-nsxt-net"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) external network name

## Attribute Reference

All properties defined in [vcloud_external_network_v2](/providers/vmware/vcloud/latest/docs/resources/external_network_v2)
resource are available.
