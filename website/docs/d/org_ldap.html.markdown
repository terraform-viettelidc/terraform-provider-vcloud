---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_org_ldap"
sidebar_current: "docs-vcd-data-source-org-ldap"
description: |-
  Provides a data source to read LDAP configuration for an organization.
---

# vcd\_org\_ldap

Supported in provider *v3.8+*.

Provides a data source to read LDAP configuration for an organization.

## Example Usage

```hcl
data "vcloud_org" "my-org" {
  name = "my-org"
}

data "vcloud_org_ldap" "first" {
  org_id = data.vcloud_org.my-org.id
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required)  - ID of the organization containing the LDAP settings

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_org_ldap`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org_ldap) resource are available.
