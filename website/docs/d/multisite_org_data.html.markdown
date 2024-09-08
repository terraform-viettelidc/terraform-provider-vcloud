---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_multisite_org_data"
sidebar_current: "docs-vcd-data-source-multisite-org-data"
description: |-
  Provides a data source to read a Viettel IDC Cloud Org association data to be used for association with another Org.
---

# vcd\_multisite\_org\_data

Provides a data source to read a Viettel IDC Cloud Org association data to be used for association with another Org.

Supported in provider *v3.13+*

## Example Usage 


```hcl
data "vcloud_org" "my-org" {
  name = "my-org"
}

data "vcloud_multisite_org_data" "current_org" {
  org_id           = data.vcloud_org.my-org.id
  download_to_file = "filename.xml"
}
```

## Argument Reference

* `org_id` - (Required) The ID of the organization for which we need to collect the data.
* `download_to_file` - (Optional) Name of the file that will contain the data needed to associate this Org to another one, 
  either on the same VCLOUD or in a different one.
  Contains the same data returned in `association_data`.

## Attribute Reference

* `association_data` - The data needed to associate this Org to another one. Contains the same data that would be saved into
  the file defined in `download_to_file`.
* `number_of_associations` - The number of current associations with other Orgs.
* `associations` - An alphabetically sorted list of current associations.

## More information

See [Site and Org association](/providers/terraform-viettelidc/vcloud/latest/docs/guides/site_org_association) for a broader description
of association workflows.