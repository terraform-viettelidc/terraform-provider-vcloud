---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_solution_landing_zone"
sidebar_current: "docs-vcd-data-source-solution-landing-zone"
description: |-
  Provides a data source to read VCLOUD Solution Add-on Landing Zone.
---

# vcloud\_solution\_landing\_zone

Supported in provider *v3.13+* and VCLOUD 10.4.1+.

Provides a data source to read VCLOUD Solution Add-on Landing Zone.

~> Only `System Administrator` can read this configuration.

## Example Usage

```hcl
data "vcloud_solution_landing_zone" "slz" {}
```

## Argument Reference

No arguments are required because this is a global configuration for VCD

## Attribute Reference

All the attributes defined in
[`vcloud_solution_landing_zone`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/solution_landing_zone)
resource are available.
