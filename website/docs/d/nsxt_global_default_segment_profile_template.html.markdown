---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_global_default_segment_profile_template"
sidebar_current: "docs-vcloud-data-source-nsxt-segment-profile-template"
description: |-
  Provides a data source to read Global Default NSX-T Segment Profile Templates.
---

# vcloud\_nsxt\_global\_default\_segment\_profile\_template

Provides a resource to manage Global Default NSX-T Segment Profile Templates.

Supported in provider *v3.11+* and Vcloud 10.4.0+ with NSX-T. Requires System Administrator privileges.

## Example Usage

```hcl
resource "vcloud_nsxt_global_default_segment_profile_template" "singleton" {
}
```
## Argument Reference

No arguments are required because this is a global Vcloud configuration

## Attribute Reference

The following attributes are supported:

* `vdc_networks_default_segment_profile_template_id` - Global Default Segment Profile
  Template ID for all VDC Networks
* `vapp_networks_default_segment_profile_template_id` - Global Default Segment Profile
  Template ID for all vApp Networks


