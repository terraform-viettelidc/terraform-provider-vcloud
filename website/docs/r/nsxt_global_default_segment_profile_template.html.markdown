---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_global_default_segment_profile_template"
sidebar_current: "docs-vcloud-resource-nsxt-segment-profile-template"
description: |-
  Provides a resource to manage Global Default NSX-T Segment Profile Templates.
---

# vcloud\_nsxt\_global\_default\_segment\_profile\_template

Provides a resource to manage Global Default NSX-T Segment Profile Templates.

Supported in provider *v3.11+* and Vcloud 10.4.0+ with NSX-T. Requires System Administrator privileges.

-> This resource is a singleton - only one configuration exists in entire Vcloud instance. Having
multiple resource definitions will override each other.

## Example Usage

```hcl
resource "vcloud_nsxt_global_default_segment_profile_template" "singleton" {
  vdc_networks_default_segment_profile_template_id  = vcloud_nsxt_segment_profile_template.complete.id
  vapp_networks_default_segment_profile_template_id = vcloud_nsxt_segment_profile_template.empty.id
}
```

## Argument Reference

The following arguments are supported:

* `vdc_networks_default_segment_profile_template_id` - (Optional) Global Default Segment Profile
  Template ID for all VDC Networks
* `vapp_networks_default_segment_profile_template_id` - (Optional) Global Default Segment Profile
  Template ID for all vApp Networks


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing  global default Segment Profile Template configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_global_default_segment_profile_template.imported optional-dummy-id
```

The above would import the global default Segment Profile Template configuration. **Note**: the
`optional-dummy-id` is not mandatory but it may be useful for `import` definitions.
