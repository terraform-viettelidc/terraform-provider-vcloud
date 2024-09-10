---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_version"
sidebar_current: "docs-vcd-data-source-version"
description: |-
  Provides a VCLOUD version data source.
---

# vcloud\_version

Provides a Viettel IDC Cloud version data source to fetch the VCLOUD version, the maximum supported API version and
perform some optional checks with version constraints.

Supported in provider *v3.12+*. Requires System Administrator privileges.

## Example Usage

```hcl
# This data source will assert that the VCLOUD version is exactly 10.5.1, otherwise it will fail
data "vcloud_version" "eq_1051" {
  condition         = "= 10.5.1"
  fail_if_not_match = true
}

# This data source will assert that the VCLOUD version is greater than or equal to 10.4.2, but it won't fail if it is not
data "vcloud_version" "gte_1042" {
  condition         = ">= 10.4.2"
  fail_if_not_match = false
}

output "is_gte_1042" {
  value = data.vcloud_version.gte_1042.matches_condition # Will show false if we're using a VCLOUD version < 10.4.2
}

# This data source will assert that the VCLOUD version is less than 10.5.0
data "vcloud_version" "lt_1050" {
  condition         = "< 10.5.0"
  fail_if_not_match = true
}

# This data source will assert that the VCLOUD version is 10.5.X
data "vcloud_version" "is_105" {
  condition         = "~> 10.5"
  fail_if_not_match = true
}

# This data source will assert that the VCLOUD version is not 10.5.1
data "vcloud_version" "not_1051" {
  condition         = "!= 10.5.1"
  fail_if_not_match = true
}
```

## Argument Reference

The following arguments are supported:

* `condition` - (Optional) A version constraint to check against the VCLOUD version
* `fail_if_not_match` - (Optional) Required if `condition` is set. Throws an error if the version constraint set in `condition` is not met

## Attribute Reference

* `matches_condition` - It is true if the VCLOUD version matches the constraint set in `condition`
* `vcloud_version` - The VCLOUD version
* `api_version` - The maximum supported API version
