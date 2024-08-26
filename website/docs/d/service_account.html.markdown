---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_service_account"
sidebar_current: "docs-vcloud-datasource-service-account"
description: |-
  Provides a data source to read Vcloud Service Accounts.
---

# vcloud\_service\_account

Provides a data source to read Vcloud Service Accounts.

Supported in provider *v3.10+* and Vcloud 10.4+.

## Example Usage

```hcl
data "vcloud_service_account" "example" {
  org  = "my-org"
  name = "my-parent-network"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `name` - (Required) Name of the Service Account in an organisation

## Attribute Reference

All the attributes defined in [`vcloud_service_account`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/service_account)
resource except `file_name` and `allow_token_file` are available.
