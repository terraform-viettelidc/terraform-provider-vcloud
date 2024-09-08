---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_catalog"
sidebar_current: "docs-vcd-data-source-catalog"
description: |-
  Provides a catalog data source.
---

# vcd\_catalog

Provides a VMware Cloud Director Catalog data source. A Catalog can be used to manage catalog items and media items.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcloud_catalog" "my-cat" {
  org  = "my-org"
  name = "my-cat"
}

resource "vcloud_catalog_item" "myItem" {
  org     = data.vcloud_catalog.my-cat.org
  catalog = data.vcloud_catalog.my-cat.name

  name              = "myItem"
  description       = "Belongs to ${data.vcloud_catalog.my-cat.id}"
  ova_path          = "/path/to/test_vapp_template.ova"
  upload_piece_size = 5
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional, but required if not set at provider level) Org name 
* `name` - (Required) Catalog name (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters

## Attribute Reference

* `description` - Catalog description.
* `publish_enabled` - (*v3.6+*) Enable allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation. Default is `false`.
* `cache_enabled` - (*v3.6+*) Enable early catalog export to optimize synchronization. Default is `false`.
* `preserve_identity_information` - (*v3.6+*) Enable include BIOS UUIDs and MAC addresses in the downloaded OVF package. Preserving the identity information limits the portability of the package and you should use it only when necessary. Default is `false`.
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata.
* `metadata_entry` - A set of metadata entries assigned to this Catalog. See [Metadata](#metadata) section for details.
* `catalog_version` - (*v3.6+*) Version number from this catalog.
* `owner_name` - (*v3.6+*) Owner of the catalog.
* `number_of_vapp_templates` - (*v3.6+*) Number of vApp templates available in this catalog.
* `number_of_media` - (*v3.6+*) Number of media items available in this catalog.
* `vapp_template_list` - (*v3.8+*) List of vApp template names in this catalog, in alphabetical order.
* `media_item_list` - (*v3.8+*) List of media item names in this catalog, in alphabetical order.
* `is_shared` - (*v3.6+*) Indicates if the catalog is shared.
* `is_published` - (*v3.6+*) Indicates if this catalog is shared to all organizations.
* `is_local` - (*v3.8.1+*) Indicates if this catalog was created in the current organization.
* `created` - (*v3.6+*) Date and time of catalog creation
* `publish_subscription_type` - (*v3.6+*) Shows if the catalog is `PUBLISHED`, if it is a subscription from another one (`SUBSCRIBED`), or none of those (`UNPUBLISHED`).
* `publish_subscription_url` - (*v3.8+*) URL to which other catalogs can subscribe.

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

* `name_regex` (Optional) matches the name using a regular expression.
* `date` (Optional) is an expression starting with an operator (`>`, `<`, `>=`, `<=`, `==`), followed by a date, with
  optional spaces in between. For example: `> 2020-02-01 12:35:00.523Z`
  The filter recognizes several formats, but one of `yyyy-mm-dd [hh:mm[:ss[.nnnZ]]]` or `dd-MMM-yyyy [hh:mm[:ss[.nnnZ]]]`
  is recommended.
  Comparison with equality operator (`==`) need to define the date to the microseconds.
* `latest` (Optional) If `true`, retrieve the latest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the newest item.
* `earliest` (Optional) If `true`, retrieve the earliest item among the ones matching other parameters. If no other parameters
  are set, it retrieves the oldest item.
* `metadata` (Optional) One or more parameters that will match metadata contents.

See [Filters reference](/providers/vmware/vcd/latest/docs/guides/data_source_filters) for details and examples.

