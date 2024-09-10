---
layout: "vcd"
page_title: "Viettel IDC Cloud: vcloud_rde"
sidebar_current: "docs-vcd-data-source-rde"
description: |-
   Provides the capability of reading an existing Runtime Defined Entity in Viettel IDC Cloud.
---

# vcloud\_rde

Provides the capability of reading an existing Runtime Defined Entity in Viettel IDC Cloud.

-> VCLOUD allows to have multiple RDEs of the same [RDE Type](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde_type) with
the same name, meaning that the data source will not be able to fetch a RDE in this situation, as this data source
can only retrieve **unique RDEs**.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcloud_rde_type" "my_type" {
  vendor    = "bigcorp"
  namespace = "tech1"
  version   = "1.2.3"
}

data "vcloud_rde" "my_rde" {
  org         = "my-org"
  rde_type_id = data.vcloud_rde_type.my-type.id
  name        = "My custom RDE"
}

output "rde_output" {
  value = vcloud_rde.my_rde.entity
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) Name of the [Organization](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/org) that owns the RDE, optional if defined at provider level.
* `rde_type_id` - (Required) The ID of the [RDE Type](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/rde_type) of the RDE to fetch.
* `name` - (Required) The name of the Runtime Defined Entity.

## Attribute Reference

The following attributes are supported:

* `entity` - The entity JSON.
* `owner_user_id` - The ID of the [Organization user](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org_user) that owns this Runtime Defined Entity.
* `org_id` - The ID of the [Organization](/providers/terraform-viettelidc/vcloud/latest/docs/resources/org) to which the Runtime Defined Entity belongs.
* `state` - It can be `RESOLVED`, `RESOLUTION_ERROR` or `PRE_CREATED`.
* `metadata_entry` - A set of metadata entries that belong to the RDE.
  Read the [resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/rde#metadata) documentation for the details of the sub-attributes.
