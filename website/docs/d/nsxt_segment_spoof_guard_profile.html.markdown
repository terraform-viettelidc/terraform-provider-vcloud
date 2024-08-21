---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_segment_spoof_guard_profile"
sidebar_current: "docs-vcd-data-source-nsxt-segment-spoof-guard-profile"
description: |-
  Provides a Viettel IDC Cloud NSX-T Spoof Guard Profile data source. This can be used to read NSX-T Segment Profile definitions.
---

# vcd\_nsxt\_segment\_spoof\_guard\_profile

Provides a Viettel IDC Cloud Spoof Guard Profile data source. This can be used to read NSX-T Segment Profile definitions.

Supported in provider *v3.11+*.

## Example Usage (IP Discovery Profile)

```hcl
data "vcloud_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcloud_nsxt_segment_spoof_guard_profile" "first" {
  name            = "spoof-guard-profile-0"
  nsxt_manager_id = data.vcloud_nsxt_manager.nsxt.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Segment Profile
* `nsxt_manager_id` - (Optional) Segment Profile search context. Use when searching by NSX-T manager
* `vdc_id` - (Optional) Segment Profile search context. Use when searching by VDC
* `vdc_group_id` - (Optional) Segment Profile search context. Use when searching by VDC group

-> Note: only one of `nsxt_manager_id`, `vdc_id`, `vdc_group_id` can be used

## Attribute reference

* `description` - Description of Spoof Guard profile
* `is_address_binding_whitelist_enabled` - Whether Spoof Guard is enabled. If true, it only allows
  VM sending traffic with the IPs in the whitelist
