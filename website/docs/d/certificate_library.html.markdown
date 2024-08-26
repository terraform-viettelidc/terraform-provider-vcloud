---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_library_certificate"
sidebar_current: "docs-vcloud-data-source-certificate-library"
description: |-
  Provides a data source to read certificate in System or Org library.
---

# vcloud\_certificate\_library
Supported in provider *v3.5+* and Vcloud 10.2+.

Provides a data source to read certificate in System or Org library and reference in other resources.

~> Only `System Administrator` can access System certificates using this data source.

## Example Usage

```hcl
data "vcloud_library_certificate" "certificate1" {
  org   = "myOrg"
  alias = "SAML Encryption"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Optional)  - alias (name) of certificate
* `id` - (Optional)  - ID of certificate

`alias` or `id` is required field.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_library_certificate`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/certificate_library) resource are available.