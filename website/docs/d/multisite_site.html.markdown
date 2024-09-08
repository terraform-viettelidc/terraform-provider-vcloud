---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_multisite_site"
sidebar_current: "docs-vcd-data-source-multisite-site"
description: |-
  Provides a data source to read a Viettel IDC Cloud Site in the context of multi-site operations.
---

# vcd\_multisite\_site

Provides a data source to read a Viettel IDC Cloud Site in the context of multi-site operatioos

Supported in provider *v3.13+*

~> Note: this data source requires System Administrator privileges

## Example Usage

Note: there is only one site available for each VCD. No ID or name is necessary to identify it.

```hcl
data "vcloud_multisite_site" "current_site" {
}
```

## Argument Reference

None needed.

## Attribute Reference

* `id` - The identification of the site. Used when associated to a remote site.
* `name` - The name of the site, which usually corresponds to its host name.
* `description` - An optional description of the site.
* `number_of_associations` - The number of current associations with other sites.
* `associations` - An alphabetically sorted list of current associations.

## More information

See [Site and Org association](/providers/vmware/vcd/latest/docs/guides/site_org_association) for a broader description
of association workflows.