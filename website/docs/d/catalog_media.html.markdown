---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_catalog_media"
sidebar_current: "docs-vcd-data-source-catalog-media"
description: |-
  Provides a catalog media data source.
---

# vcd\_catalog\_media

Provides a VMware Cloud Director Catalog media data source. A Catalog media can be used to reference a catalog media and use its 
data within other resources or data sources.

Supported in provider *v2.5+*

## Example Usage

```hcl
data "vcloud_catalog" "my-catalog" {
  org  = "my-org"
  name = "my-catalog"
}

data "vcloud_catalog_media" "existing-media" {
  org        = "my-org"
  catalog_id = data.vcloud_catalog.my-catalog.id
  name       = "my-media"
}

output "media_size" {
  value = data.vcloud_catalog_media.existing-media.size
}

output "type_is_iso" {
  value = data.vcloud_catalog_media.existing-media.is_iso
}


```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `catalog` - (Optional; Deprecated) The name of the catalog to which media file belongs. It's mandatory if `catalog_id` is not used.
* `catalog_id` - (Optional; *v3.8.2+*) The ID of the catalog to which the media file belongs. It's mandatory if `catalog` field is not used.
* `name` - (Required) Media name in catalog (optional when `filter` is used)
* `filter` - (Optional; *2.9+*) Retrieves the data source using one or more filter parameters
* `download_to_file` - (Optional; *3.11+*) Will download the contents of the media item into the given file

-> NOTE: downloading of media items can take unexpectedly long amounts of time for large items. The ability of
downloading media items is supplied here to solve a specific problem, i.e. saving small files in the VCD as help
to other workflows. For example, you could save into a media item an HCL file used to configure the VCD, the file 
`terraform.tfstate`, planning documents, an image of the deployment topology, and so on.

## Attribute reference

All attributes defined in [catalog_media](/providers/vmware/vcd/latest/docs/resources/catalog_media#attribute-reference) are supported.

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

