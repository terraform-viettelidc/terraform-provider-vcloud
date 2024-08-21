---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_solution_landing_zone"
sidebar_current: "docs-vcloud-data-source-solution-landing-zone"
description: |-
  Provides a data source to read Vcloud Solution Add-on Landing Zone.
---

# vcloud\_solution\_landing\_zone

Supported in provider *v3.13+* and Vcloud 10.4.1+.

Provides a data source to read Vcloud Solution Add-on Landing Zone.

~> Only `System Administrator` can read this configuration.

## Example Usage

```hcl
data "vcloud_solution_landing_zone" "slz" {}
```

## Argument Reference

No arguments are required because this is a global configuration for Vcloud

## Attribute Reference

All the attributes defined in
[`vcloud_solution_landing_zone`](/providers/vmware/vcloud/latest/docs/resources/solution_landing_zone)
resource are available.
