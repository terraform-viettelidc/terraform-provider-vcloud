---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_catalog_item"
sidebar_current: "docs-vcd-data-source-catalog-item"
description: |-
  Provides a catalog item data source.
---

# vcd\_catalog\_item

-> If you only need vApp Template features, you may use [`vcloud_catalog_vapp_template`](/providers/vmware/vcd/latest/docs/data-sources/catalog_vapp_template) instead.

Provides a VMware Cloud Director Catalog item data source. A Catalog item can be used to reference a catalog item and use its 
data within other resources or data sources.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcloud_catalog_item" "my-first-item" {
  org     = "my-org"
  catalog = "my-cat"
  name    = "my-first-item"
}

resource "vcloud_catalog_item" "my-second-item" {
  # Using the data source, two properties from another catalog items are
  # used in this resource.
  # You can read it as "use the org from catalog item `my-first-item`"
  # and "use the catalog from catalog item `my-first-item`"
  org     = data.vcloud_catalog_item.my-first-item.org
  catalog = data.vcloud_catalog_item.my-first-item.catalog

  name = "my-second-item"

  # The description uses the data source to create a dynamic text
  # The description will become "Belongs to my-cat"
  description       = "Belongs to ${data.vcloud_catalog_item.my-first-item.catalog}"
  ova_path          = "/path/to/test_vapp_template.ova"
  upload_piece_size = 5
  metadata          = data.vcloud_catalog_item.my-first-item.metadata
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `catalog` - (Required) Catalog name
* `name` - (Required) Catalog Item name (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute Reference

* `description` - Catalog item description.
* `metadata` - Key value map of metadata assigned to the associated vApp template.
* `catalog_item_metadata` - (Deprecated) Use `metadata_entry` instead. Key value map of metadata assigned to the catalog item.
* `metadata_entry` - A set of metadata entries assigned to the catalog item. See [Metadata](#metadata) section for details.

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - Key of this metadata entry.
* `value` - Value of this metadata entry.
* `type` - Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

## Filter arguments

(Supported in provider *v2.9+*)

* `name_regex` - (Optional) matches the name using a regular expression.
* `date` - (Optional) is an expression starting with an operator (`>`, `<`, `>=`, `<=`, `==`), followed by a date, with
  optional spaces in between. For example: `> 2020-02-01 12:35:00.523Z`
  The filter recognizes several formats, but one of `yyyy-mm-dd [hh:mm[:ss[.nnnZ]]]` or `dd-MMM-yyyy [hh:mm[:ss[.nnnZ]]]`
  is recommended.
  Comparison with equality operator (`==`) need to define the date to the microseconds.
* `latest` - (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `earliest` - (Optional) If `true`, retrieve the earliest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the oldest item.
* `metadata` - (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/vmware/vcd/latest/docs/guides/data_source_filters) for details and examples.

