---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_org_oidc"
sidebar_current: "docs-vcloud-data-source-org-oidc"
description: |-
  Provides a data source to read the OpenID Connect (OIDC) configuration of an Organization in Viettel IDC Cloud.
---

# vcloud\_org\_oidc

Provides a data source to read the OpenID Connect (OIDC) configuration of an Organization in Viettel IDC Cloud.

Supported in provider *v3.13+*.

## Example Usage

```hcl
data "vcloud_org" "my_org" {
  name = "my-org"
}

data "vcloud_org_oidc" "oidc_settings" {
  org_id = data.vcloud_org.my_org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) - ID of the organization containing the OIDC settings

## Attribute Reference

All the arguments and attributes from [the `vcloud_org_oidc` resource](/providers/vmware/vcloud/latest/docs/resources/org_oidc) are available as read-only.
