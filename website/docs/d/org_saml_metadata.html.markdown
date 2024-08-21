---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_org_saml_metadata"
sidebar_current: "docs-vcloud-data-source-org-saml_metadata"
description: |-
  Provides a data source to read SAML metadata for an organization.
---

# vcloud\_org\_saml\_metadata

Supported in provider *v3.10+*.

Provides a data source to read service provider SAML metadata for an organization.
This service provider metadata is used to configure the identity provider.

## Example Usage

```hcl
data "vcloud_org" "my-org" {
  name = "my-org"
}

data "vcloud_org_saml_metadata" "first" {
  org_id    = data.vcloud_org.my-org.id
  file_name = "vcloud-metadata.txt"
}

# The metadata will be stored in vcloud-metadata.txt
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) ID of the organization containing the SAML metadata
* `file_name` - (Optional) name of the file where to store the metadata.

## Attribute Reference

* `metadata_text` - the text of the metadata for this organization. 