---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_vapp"
sidebar_current: "docs-vcd-datasource-vapp"
description: |-
  Provides a Viettel IDC Cloud vApp data source. This can be used to reference vApps.
---

# vcd\_vapp

Provides a Viettel IDC Cloud vApp data source. This can be used to reference vApps.

Supported in provider *v2.5+*

## Example Usage


```hcl
data "vcloud_vapp" "test-tf" {
  name = "test-tf"
  org  = "tf"
  vdc  = "vdc-tf"
}

output "id" {
  value = data.vcloud_vapp.test-tf.id
}

output "name" {
  value = data.vcloud_vapp.test-tf.name
}

output "description" {
  value = data.vcloud_vapp.test-tf.description
}

output "href" {
  value = data.vcloud_vapp.test-tf.href
}

output "status_text" {
  value = data.vcloud_vapp.test-tf.status_text
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level

## Attribute reference

* `description` An optional description for the vApp
* `href` - The vApp Hyper Reference
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key value map of metadata assigned to this vApp. Key and value can be any string. 
* `metadata_entry` - A set of metadata entries assigned to this vApp. See [Metadata](#metadata) section for details.
* `guest_properties` -  Key value map of vApp guest properties.
* `status` -  The vApp status as a numeric code
* `status_text` -  The vApp status as text.
* `lease` - (*v3.5+*) - The information about the vApp lease. It includes the following fields:
  * `runtime_lease_in_sec` - How long any of the VMs in the vApp can run before the vApp is automatically powered off or suspended. 0 means never expires.
  * `storage_lease_in_sec` - How long the vApp is available before being automatically deleted or marked as expired. 0 means never expires.
* `inherited_metadata` - (*v3.11+*; *VCLOUD 10.5.1+*) A map that contains read-only metadata that is automatically added by VCLOUD (10.5.1+) and provides
  details on the origin of the vApp (e.g. `vapp.origin.id`, `vapp.origin.name`, `vapp.origin.type`).
* `vm_names` - (*v3.13.0+*) A list of VM names included in this vApp
* `vapp_network_names` - (*3.13.0+*) A list of vApp network names included in this vApp
* `vapp_org_network_names` - (*v3.13.0+*) A list of vApp Org network names included in this vApp

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.
