---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edgegateway"
sidebar_current: "docs-vcd-data-source-nsxt-edge-gateway"
description: |-
  Provides a Viettel IDC Cloud NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.
---

# vcd\_nsxt\_edgegateway

Provides a Viettel IDC Cloud NSX-T edge gateway data source. This can be used to read NSX-T edge gateway configurations.

Supported in provider *v3.1+*.

## Example Usage (NSX-T Edge Gateway belonging to VDC Group)

```hcl
data "vcloud_vdc_group" "group1" {
  name = "existing-group"
}

data "vcloud_nsxt_edgegateway" "t1" {
  org      = "myorg"
  owner_id = data.vcloud_vdc_group.group1.id
  name     = "nsxt-edge-gateway"
}
```

## Example Usage (NSX-T Edge Gateway belonging to VDC)

```hcl
data "vcloud_org_vdc" "vdc1" {
  name = "existing-vdc"
}

data "vcloud_nsxt_edgegateway" "t1" {
  org      = "myorg"
  owner_id = data.vcloud_org_vdc.vdc1.id
  name     = "nsxt-edge-gateway"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the NSX-T Edge Gateway belongs. Optional if
  defined at provider level.
* `vdc` - (Optional)  **Deprecated** - please use `owner_id` field. The name of VDC that owns the
  NSX-T Edge Gateway. Optional if defined at provider level.
* `owner_id` - (Optional, *v3.6+*,*VCLOUD 10.2+*) The ID of VDC or VDC Group. **Note:** Data sources
  [vcloud_vdc_group](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/vdc_group) or
  [vcloud_org_vdc](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/org_vdc) can be used to lookup IDs by
  name.

~> Only one of `vdc` or `owner_id` can be specified. `owner_id` takes precedence over `vdc`
definition at provider level.

* `name` - (Required) NSX-T Edge Gateway name.
* `ip_count_read_limit` - (Optional, *v3.13+*) Sets a limit of IPs to count for
  `used_ip_count` and `unused_ip_count` attributes to avoid exhausting compute resource while
  counting IPs in large IPv6 subnets. It does not affect operation of Edge Gateway configuration,
  only IP count reporting. Defaults to `1000000`. While it is unlikely that a single Edge Gateway
  can effectively manage more IPs, one can specify `0` for *unlimited* value. 

## Attribute reference

All properties defined in [vcloud_nsxt_edgegateway](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_edgegateway)
resource are available.
