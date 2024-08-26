---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_dynamic_security_group"
sidebar_current: "docs-vcloud-data-source-nsxt-dynamic-security-group"
description: |-
  Provides a data source to read NSX-T Dynamic Security Groups. Dynamic Security Groups group Virtual
  Machines based on specific criteria (VM Names or Security tags) to which Distributed Firewall Rules
  apply.
---

# vcloud\_nsxt\_dynamic\_security\_group

Supported in provider *v3.7+* and Vcloud 10.3+ with NSX-T backed VDC Groups.

Provides a data source to read NSX-T Dynamic Security Groups. Dynamic Security Groups group Virtual
Machines based on specific criteria (VM Names or Security tags) to which Distributed Firewall Rules
apply.

## Example Usage 1 (Existing Dynamic Security Group lookup)

```hcl
data "vcloud_vdc_group" "group1" {
  org  = "cloud"
  name = "vdc-group-cloud"
}

data "vcloud_nsxt_dynamic_security_group" "group1" {
  org          = "cloud"
  vdc_group_id = data.vcloud_vdc_group.group1.id

  name = "cloud-dynamic-security-group"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc_group_id` - (Required) VDC Group ID hosting existing Dynamic Security Group.
* `name` - (Required) A unique name for existing Dynamic Security Group

All the arguments and attributes defined in
[`vcloud_nsxt_dynamic_security_group`](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_dynamic_security_group) resource are available.
