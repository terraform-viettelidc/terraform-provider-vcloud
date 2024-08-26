---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_org_vdc_nsxt_network_profile"
sidebar_current: "docs-vcloud-datasource-nsxt-network-segment-profile"
description: |-
  Provides a data source to read Segment Profile configuration for NSX-T Org VDC networks.
---

# vcloud\_nsxt\_network\_segment\_profile

Provides a data source to read Segment Profile configuration for NSX-T Org VDC networks.

Supported in provider *v3.11+* and Vcloud 10.4.0+ with NSX-T.

## Example Usage

```hcl
data "vcloud_nsxt_network_segment_profile" "custom-prof" {
  org            = "my-org"
  org_network_id = vcloud_network_routed_v2.net1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level 
* `org_network_id` - (Required) Org VDC Network ID

## Attribute Reference
 
All the arguments and attributes defined in
[`vcloud_nsxt_network_segment_profile`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_segment_profile)
resource are available.
