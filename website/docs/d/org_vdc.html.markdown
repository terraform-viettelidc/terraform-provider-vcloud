---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_org_vdc"
sidebar_current: "docs-vcd-data-source-org-vdc"
description: |-
  Provides an organization VDC data source.
---

# vcloud\_org\_vdc

Provides a Viettel IDC Cloud Organization VDC data source. An Organization VDC can be used to
reference a VDC and use its data within other resources or data sources.

-> **Note:** This resource supports NSX-T and NSX-V based Org VDCs

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcloud_org_vdc" "my-org-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

output "provider_vdc" {
  value = data.vcloud_org_vdc.my-org-vdc.provider_vdc_name
}

```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `name` - (Required) Organization VDC name

## Attribute reference

* `edge_cluster_id` - (*v3.8+*, *VCLOUD 10.3+*) An ID of NSX-T Edge Cluster which should provide vApp
  Networking Services or DHCP for Isolated Networks. This value **might be unavailable** in data
  source if user has insufficient rights.

All other attributes are defined in [organization VDC
resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org_vdc#attribute-reference) are supported.

