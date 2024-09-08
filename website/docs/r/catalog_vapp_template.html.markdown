---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_catalog_vapp_template"
sidebar_current: "docs-vcd-resource-catalog-vapp_template"
description: |-
  Provides a Viettel IDC Cloud vApp Template resource. This can be used to upload and delete OVA files inside a catalog.
---

# vcd\_catalog\_vapp\_template

Provides a Viettel IDC Cloud vApp Template resource. This can be used to upload OVA to catalog and delete it.

Supported in provider *v3.8+*

## Example Usage (OVA upload)

```hcl
data "vcloud_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcloud_catalog_vapp_template" "myNewVappTemplate" {
  org        = "my-org"
  catalog_id = data.vcloud_catalog.my-catalog.id

  name              = "my ova"
  description       = "new vapp template"
  ova_path          = "/home/user/file.ova"
  upload_piece_size = 10

  lease {
    storage_lease_in_sec = 60 * 60 * 24 * 7 # set storage lease for 7 days (60 seconds * 60 minutes * 24 hour * 7 days)
  }

  metadata_entry {
    key         = "license"
    value       = "public"
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "version"
    value       = "v1"
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

-> If vApp Template upload fails, or you need to re-upload it, you can do a `terraform apply -replace=vcloud_catalog_vapp_template.myNewVappTemplate`.

## Example Usage (Capturing from existing vApp)

```hcl
data "vcloud_catalog" "cat" {
  org  = "v51"
  name = "cat-v51-nsxt-backed"
}

resource "vcloud_catalog_vapp_template" "from-vapp" {
  org        = "v51"
  catalog_id = data.vcloud_catalog.cat.id

  name = "from-vapp"

  capture_vapp {
    source_id                = vcloud_vapp.web.id
    customize_on_instantiate = false
  }

  lease {
    storage_lease_in_sec = 3600 * 24 * 3
  }

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
  }

  depends_on = [vcloud_vapp_vm.emptyVM] # Ensuring all VMs are present in vApp
}
```

## Example Usage (Capturing from existing Standalone VM)

```hcl
data "vcloud_catalog" "cat" {
  org  = "v51"
  name = "cat-v51-nsxt-backed"
}

resource "vcloud_catalog_vapp_template" "from-standalone-vm" {
  org        = "v51"
  catalog_id = data.vcloud_catalog.cat.id

  name = "captured-vApp"

  capture_vapp {
    source_id                = vcloud_vm.standalone.vapp_id # Parent hidden vApp must be referenced
    customize_on_instantiate = true                      # Can only be `true` if source vApp is powered off
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `catalog_id` - (Required) ID of the Catalog where to upload the OVA file
* `name` - (Required) vApp Template name in Catalog
* `description` - (Optional) Description of the vApp Template. Not to be used with `ovf_url` when target OVA has a description
* `ova_path` - (Optional) Absolute or relative path to file to upload
* `ovf_url` - (Optional) URL to OVF file. Only OVF (not OVA) files are supported by VCLOUD uploading by URL
* `capture_vapp` - (Optional; *v3.12+*) A configuration [block to create template from existing
  vApp](#capture-vapp) (Standalone VM or vApp)
* `upload_piece_size` - (Optional) - Size in MB for splitting upload size. It can possibly impact upload performance. Default 1MB
* `metadata` -  (Deprecated) Use `metadata_entry` instead. Key/value map of metadata to assign to the associated vApp Template
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.
* `lease` - (Optional *v3.11+*) The information about the vApp Template lease. It includes the field below. When this section is
  included, the field is mandatory. If lease value is higher than the one allowed for the whole Org, we get an error
   * `storage_lease_in_sec` - How long the vApp Template is available before being automatically deleted or marked as expired. 0 means never expires (or maximum allowed by parent Org). Regular values accepted from 3600+.

## Attribute Reference

* `vdc_id` - The VDC ID to which this vApp Template belongs
* `vm_names` - Set of VM names within the vApp template
* `created` - Timestamp of when the vApp Template was created
* `catalog_item_id` - Catalog Item ID

<a id="capture-vapp"></a>
## Capture vApp template from existing vApp or Standalone VM

* `source_id` - (Required) Source vApp ID (can be referenced by `vcloud_vapp.id` or
  `vcloud_vm.vapp_id`/`vcloud_vapp_vm.vapp_id`)
* `overwrite_catalog_item_id` - (Optional) Optionally newly created template can overwrite. It can
  either be `id` of `vcloud_catalog_item` resource or `catalog_item_id` of
  `vcloud_catalog_vapp_template` resource
* `customize_on_instantiate` - (Optional) Default `false` - means "Make identical copy". `true`
  means "Customize VM settings". *Note* `true` can only be set when source vApp is powered off

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
resource "vcloud_catalog_vapp_template" "example" {
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

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing vApp Template can be [imported][docs-import] into this resource via supplying the full dot separated path for a
vApp Template. For example, using this structure, representing an existing vAppTemplate that was **not** created using Terraform:

```hcl
data "vcloud_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

resource "vcloud_catalog_vapp_template" "my-vapp-template" {
  org        = "my-org"
  catalog_id = data.vcloud_catalog.my-catalog.id
  name       = "my-vapp-template"
  ova_path   = "guess"
}
```

You can import such vApp Template into terraform state using this command

```
terraform import vcloud_catalog_vapp_template.my-vapp-template my-org.my-catalog.my-vapp-template
```

You can also import a vApp Template using a VDC name instead of a Catalog name:

```
terraform import vcloud_catalog_vapp_template.my-vapp-template my-org.my-vdc.my-vapp-template
```


NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCLOUD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the vApp Template as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the vApp Template's stored properties.
