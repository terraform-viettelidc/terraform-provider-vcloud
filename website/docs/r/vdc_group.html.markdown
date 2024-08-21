---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vdc_group"
sidebar_current: "docs-vcd-resource-vdc-group"
description: |-
  Provides a VDC group resource.
---

# vcd\_vdc\_group
Supported in provider *v3.5+* and VCD 10.2+.

Provides a resource to manage NSX-T VDC groups.

~> Only `System Administrator` and `Org Users` with rights `View VDC Group`, `Configure VDC Group`, `vDC Group: Configure Logging`, `Organization vDC Distributed Firewall: Enable/Disable` can manage VDC groups using this resource.

## Example Usage

```hcl
data "vcloud_org_vdc" "startVdc" {
  name = "existingVdc"
}

data "vcloud_org_vdc" "additionalVdc" {
  name = "oneMoreVdc"
}

resource "vcloud_vdc_group" "new-vdc-group" {
  org                   = "myOrg"
  name                  = "newVdcGroup"
  description           = "my description"
  starting_vdc_id       = data.vcloud_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcloud_org_vdc.startVdc.id, data.vcloud_org_vdc.additionalVdc.id]
  dfw_enabled           = true
  default_policy_status = true
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations
* `name` - (Required) The name for VDC group
* `description` - (Optional) VDC group description
* `starting_vdc_id` - (Required) With selecting a starting VDC you will be able to create a group in which this VDC can participate. **Note**: `starting_vdc_id` must be included in `participating_vdc_ids` to participate in this group.
* `participating_vdc_ids` - (Required) The list of organization VDCs that are participating in this group. **Note**: `starting_vdc_id` isn't automatically included in this list.
* `dfw_enabled` - (Optional) Whether Distributed Firewall is enabled for this VDC group.
* `default_policy_status` - (Optional) Whether this security policy is enabled. `dfw_enabled` must be `true`.
* `remove_default_firewall_rule` - (Optional, *3.10+*) Marks whether default firewall rule should be
  removed after activating. Both `dfw_enabled` and `default_policy_status` must be true. **Note.**
  This is mainly useful when using
  [`vcloud_nsxt_distributed_firewall_rule`](/providers/vmware/vcd/latest/docs/resources/nsxt_distributed_firewall_rule)
  resource as it cannot remove the default rule.
* `force_delete` - (Optional, *3.11+*) When `true`, will request VCD to force VDC Group deletion. It
  should clean up child components. Default `false` (VCD may fail removing VDC Group if there are
  child components remaining). **Note:** when setting it to `true` for existing resource, it will
  cause a plan change (update), but this will not alter the resource in any way.

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The VDC group ID
* `error_message` - More detailed error message when VDC group has error status
* `local_egress` - Status whether local egress is enabled for a universal router belonging to a universal VDC group.
* `network_pool_id` - ID of used network pool.
* `network_pool_universal_id` - The network provider’s universal id that is backing the universal network pool.
* `network_provider_type` - Defines the networking provider backing the VDC group.
* `status` - The status that the group can be in (e.g. 'SAVING', 'SAVED', 'CONFIGURING', 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND', 'UNCONFIGURED').
* `type` - Defines the group as LOCAL or UNIVERSAL.
* `universal_networking_enabled` - True means that a VDC group router has been created.
* `participating_org_vdcs` - A list of blocks providing organization VDCs that are participating in this group details. See [Participating Org VDCs](#participatingOrgVdcs) below for details.

<a id="participatingOrgVdcs"></a>
## Participating Org VDCs

* `vdc_id` - VDC ID.
* `vdc_name` - VDC name.
* `site_id` - Site ID.
* `site_name` - Site name.
* `org_id` - Organization ID.
* `org_name` - Organization name.
* `status` - "The status that the VDC can be in e.g. 'SAVING', 'SAVED', 'CONFIGURING', 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND', 'UNCONFIGURED')."
* `is_remote_org` - Specifies whether the VDC is local to this VCD site.
* `network_provider_scope` - Specifies the network provider scope of the VDC.
* `fault_domain_tag` - Represents the fault domain of a given organization VDC.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing VDC group can be [imported][docs-import] into this resource
via supplying the full dot separated path VDC group. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcloud_vdc_group.imported my-org.my-vdc-group
```

The above would import the VDC group named `my-vdc-group` which is configured in organization named `my-org`.
