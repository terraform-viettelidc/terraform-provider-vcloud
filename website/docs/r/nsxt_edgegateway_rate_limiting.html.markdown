---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edgegateway_rate_limiting"
sidebar_current: "docs-vcd-resource-nsxt-edge-rate-limiting"
description: |-
  Provides a resource to manage NSX-T Edge Gateway Rate Limiting (QoS) configuration.
---

# vcd\_nsxt\_edgegateway\_rate\_limiting

Supported in provider *v3.9+* and VCLOUD 10.3.2+ with NSX-T.

Provides a resource to manage NSX-T Edge Gateway Rate Limiting (QoS) configuration.

## Example Usage

```hcl
data "vcloud_nsxt_manager" "nsxt" {
  name = "nsxManager1"
}

data "vcloud_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.vcloud_nsxt_manager.nsxt.id
  name            = "qos-profile-1"
}

data "vcloud_org_vdc" "v1" {
  org  = "datacloud"
  name = "nsxt-vdc-datacloud"
}

data "vcloud_nsxt_edgegateway" "testing-in-vdc" {
  org      = "datacloud"
  owner_id = data.vcloud_org_vdc.v1.id

  name = "nsxt-gw-datacloud"
}

resource "vcloud_nsxt_edgegateway_rate_limiting" "testing-in-vdc" {
  org             = "datacloud"
  edge_gateway_id = data.vcloud_nsxt_edgegateway.testing-in-vdc.id

  ingress_profile_id = data.vcloud_nsxt_edgegateway_qos_profile.qos-1.id
  egress_profile_id  = data.vcloud_nsxt_edgegateway_qos_profile.qos-1.id
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Required) Org in which the NSX-T Edge Gateway is located
* `edge_gateway_id` - (Required) NSX-T Edge Gateway ID
* `ingress_profile_id` - (Optional) A QoS profile to apply for ingress traffic. *Note* leaving empty
  means `unlimited`.
* `egress_profile_id` - (Optional) A QoS profile to apply for egress traffic. *Note* leaving empty
  means `unlimited`.

-> Ingress and Egress profile IDs can be looked up using
  [`vcloud_nsxt_edgegateway_qos_profile`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_edgegateway_qos_profile)
  data source. 

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing NSX-T Edge Gateway Rate Limiting configuration can be [imported][docs-import] into this
resource via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_edgegateway_rate_limiting.imported my-org.nsxt-vdc.nsxt-edge
```

The above would import the `nsxt-edge` Edge Gateway Rate Limiting configuration for this particular
Edge Gateway.
