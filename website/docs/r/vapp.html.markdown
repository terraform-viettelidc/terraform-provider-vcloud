---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_vapp"
sidebar_current: "docs-vcloud-resource-vapp"
description: |-
  Provides a Viettel IDC Cloud vApp resource. This can be used to create, modify, and delete vApps.
---

# vcloud\_vapp

Provides a Viettel IDC Cloud vApp resource. This can be used to create, modify, and delete vApps.

## Example of vApp with 2 VMs

Example with more than one VM under a vApp.

```hcl
resource "vcloud_network_direct" "direct-network" {
  name             = "net"
  external_network = "my-ext-net"
}

resource "vcloud_vapp" "web" {
  name = "web"

  metadata_entry {
    key   = "CostAccount"
    value = "Marketing Department"
  }
}

resource "vcloud_vapp_org_network" "direct-network" {
  vapp_name        = vcloud_vapp.web.name
  org_network_name = vcloud_network_direct.direct-network.name
}

data "vcloud_catalog" "my-catalog" {
  org  = "test"
  name = "my-catalog"
}

data "vcloud_catalog_vapp_template" "photon-os" {
  org        = "test"
  catalog_id = data.vcloud_catalog.my-catalog.id
  name       = "photon-os"
}

resource "vcloud_vapp_vm" "web1" {
  vapp_name = vcloud_vapp.web.name
  name      = "web1"

  vapp_template_id = data.vcloud_catalog_vapp_template.photon-os.id

  memory = 2048
  cpus   = 1

  network {
    type               = "org"
    name               = vcloud_vapp_org_network.direct-network.org_network_name
    ip_allocation_mode = "POOL"
  }

  guest_properties = {
    "vapp.property1" = "value1"
    "vapp.property2" = "value2"
  }

  lease {
    runtime_lease_in_sec = 60 * 60 * 24 * 30 # extends the runtime lease to 30 days
    storage_lease_in_sec = 60 * 60 * 24 * 7  # extends the storage lease to 7 days
  }
}

resource "vcloud_vapp_vm" "web2" {
  vapp_name = vcloud_vapp.web.name
  name      = "web2"

  vapp_template_id = data.vcloud_catalog_vapp_template.photon-os.id

  memory = 2048
  cpus   = 1

  network {
    type               = "org"
    name               = vcloud_vapp_org_network.direct-network.org_network_name
    ip_allocation_mode = "POOL"
  }
}
```

## Example of Empty vApp with no VMs

```hcl
resource "vcloud_vapp" "web" {
  name = "web"

  metadata_entry {
    key   = "boss"
    value = "Why is this vApp empty?"
  }

  metadata_entry {
    key   = "john"
    value = "I don't really know. Maybe somebody did forget to clean it up."
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name for the vApp
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level
* `description` (Optional; *v3.3*) An optional description for the vApp, up to 256 characters.
* `power_on` - (Optional) A boolean value stating if this vApp should be powered on. Default is `false`. Works only on update when vApp already has VMs.
* `metadata` - (Deprecated) Use `metadata_entry` instead. Key value map of metadata to assign to this vApp. Key and value can be any string. (Since *v2.2+* metadata is added directly to vApp instead of first VM in vApp)
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `guest_properties` - (Optional; *v2.5+*) Key value map of vApp guest properties
* `lease` - (Optional *v3.5+*) the information about the vApp lease. It includes the fields below. When this section is 
   included, both fields are mandatory. If lease values are higher than the ones allowed for the whole Org, the values
   are **silently** reduced to the highest value allowed.
  * `runtime_lease_in_sec` - How long any of the VMs in the vApp can run before the vApp is automatically powered off or suspended. 0 means never expires (or maximum allowed by Org). Regular values accepted from 3600+.
  * `storage_lease_in_sec` - How long the vApp is available before being automatically deleted or marked as expired. 0 means never expires (or maximum allowed by Org). Regular values accepted from 3600+.

## Attribute reference

* `href` - (Computed) The vApp Hyper Reference.
* `status` - (Computed; *v2.5+*) The vApp status as a numeric code.
* `status_text` - (Computed; *v2.5+*) The vApp status as text.
* `inherited_metadata` - (Computed; *v3.11+*; *Vcloud 10.5.1+*) A map that contains read-only metadata that is automatically added by Vcloud (10.5.1+) and provides
  details on the origin of the vApp (e.g. `vapp.origin.id`, `vapp.origin.name`, `vapp.origin.type`).

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcloud_vapp" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = true # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

To remove all metadata one needs to specify an empty `metadata_entry`, like:

```
metadata_entry {}
```

The same applies also for deprecated `metadata` attribute:

```
metadata = {}
```

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing vApp can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.vapp-name
For example, using this structure, representing a vApp that was **not** created using Terraform:

```hcl
resource "vcloud_vapp" "tf-vapp" {
  name = "my-vapp"
  org  = "my-org"
  vdc  = "my-vdc"
}
```

You can import such vapp into terraform state using this command

```
terraform import vcloud_vapp.tf-vapp my-org.my-vdc.my-vapp
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable vcloud_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.

