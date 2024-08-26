---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_security_group"
sidebar_current: "docs-vcloud-resource-nsxt-security-group"
description: |-
  Provides a resource to manage NSX-T Security Group. Security Groups are groups of data center
  group networks to which distributed firewall rules apply. Grouping networks helps you to reduce
  the total number of distributed firewall rules to be created.
---

# vcloud\_nsxt\_security\_group

Supported in provider *v3.3+* and Vcloud 10.1+ with NSX-T backed VDCs.

Provides a resource to manage NSX-T Security Group. Security Groups are groups of data center group
networks to which distributed firewall rules apply. Grouping networks helps you to reduce the total
number of distributed firewall rules to be created.

-> Starting with **v3.6.0** `vcloud_nsxt_security_group` added support for VDC Groups.
The `vdc` field (in resource or inherited from provider configuration) is deprecated, as `vcloud_nsxt_security_group` will
inherit the VDC Group or VDC membership from a parent Edge Gateway specified in the `edge_gateway_id` field.
More about VDC Group support in a [VDC Groups guide](/providers/terraform-viettelidc/vcloud/latest/docs/guides/vdc_groups).

## Example Usage 1 (Security Group with member networks)

```hcl
data "vcloud_nsxt_edgegateway" "main" {
  org  = "my-org" # Optional
  name = "main-edge"
}

resource "vcloud_nsxt_security_group" "frontend-servers" {
  org = "my-org" # Optional

  # Referring to a data source for existing NSX-T Edge Gateway
  edge_gateway_id = data.vcloud_nsxt_edgegateway.main.id

  name        = "frontend-servers"
  description = "Security Group for a network connecting the frontend servers"

  member_org_network_ids = [vcloud_network_routed_v2.frontend.id]
}
```

## Example Usage 2 (Empty Security Group)
```hcl
data "vcloud_nsxt_edgegateway" "main" {
  org  = "my-org" # Optional
  name = "main-edge"
}

resource "vcloud_nsxt_security_group" "group1" {
  org = "my-org" # Optional

  # Referring to a data source for existing NSX-T Edge Gateway
  edge_gateway_id = data.vcloud_nsxt_edgegateway.existing.id

  name        = "precreated security group"
  description = "Members to be added later"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Deprecated; Optional) The name of VDC to use, optional if defined at provider level. **Deprecated**
  in favor of `edge_gateway_id` field.
* `name` - (Required) A unique name for Security Group
* `description` - (Optional) An optional description of the Security Group
* `edge_gateway_id` - (Required) The ID of the Edge Gateway (NSX-T only). Can be looked up using
  `vcloud_nsxt_edgegateway` data source
* `member_org_network_ids` - (Optional) A set of Org Network IDs

## Attribute Reference
* `member_vms` A set of member VMs (if exist). see [Member VMs](#member-vms) below for details.

<a id="member-vms"></a>
## Member VMs

Each member VM contains following attributes:

* `vm_id` - Member VM ID
* `vm_name` - Member VM name
* `vapp_id` - Parent vApp ID for member VM (empty for standalone VMs)
* `vapp_name` - Parent vApp Name for member VM (empty for standalone VMs)

~> There may be cases where Org Networks and Security Groups are already created, but
not all VMs are already created and not shown in this structure. Additional `depends_on` can ensure
that Security Group is created only after all networks and VMs are there.

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Security Group configuration can be [imported][docs-import] into this resource
via supplying the full dot separated path for your Security Group name. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_nsxt_security_group.imported my-org.my-org-vdc.my-nsxt-edge-gateway.my-security-group-name
or
terraform import vcloud_nsxt_security_group.imported my-org.my-org-vdc-group-name.my-nsxt-edge-gateway.my-security-group-name
```

The above would import the `my-security-group-name` Security Group config settings that are defined
on NSX-T Edge Gateway `my-nsxt-edge-gateway` which is configured in organization named `my-org` and
VDC named `my-org-vdc` or VDC Group `my-vdc-group-name.
