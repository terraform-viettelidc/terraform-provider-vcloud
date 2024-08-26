---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_network_imported"
sidebar_current: "docs-vcloud-data-source-nsxt-network-imported"
description: |-
  Provides a Viettel IDC Cloud Org VDC NSX-T Imported Network data source to read data or reference existing network.
---

# vcloud\_nsxt\_network\_imported

Provides a Viettel IDC Cloud Org VDC NSX-T Imported Network data source to read data or reference existing network.

Supported in provider *v3.2+* for NSX-T VDCs only.

-> This is **not Terraform imported** data source, but a special **Imported** type of **Org VDC
network** in NSX-T VDC. Read more about Imported Network in [official Vcloud
documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-FB303D62-67EA-4209-BE4D-C3746481BCC8.html).

## Example Usage (Looking up Imported Network in VDC)

```hcl
data "vcloud_org_vdc" "main" {
  org  = "my-org"
  name = "main-edge"
}

data "vcloud_nsxt_network_imported" "net" {
  org      = "my-org"
  owner_id = data.vcloud_org_vdc.main.id
  name     = "my-net"
}
```

## Example Usage (Looking up Imported Network in VDC Group)

```hcl
data "vcloud_vdc_group" "main" {
  org  = "my-org"
  name = "main-group"
}

data "vcloud_nsxt_network_imported" "net" {
  org      = "my-org"
  owner_id = data.vcloud_vdc_group.main.id
  name     = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `owner_id` - (Optional) VDC or VDC Group ID. Always takes precedence over `vdc` fields (in resource
and inherited from provider configuration)
* `vdc` - (Deprecated; Optional) The name of VDC to use. **Deprecated**  in favor of new field
  `owner_id` which supports VDC and VDC Group IDs.
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [imported network resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/nsxt_network_imported#attribute-reference) are supported.

## Filter arguments

* `name_regex` - (Optional) matches the name using a regular expression.
* `ip` - (Optional) matches the IP of the resource using a regular expression.

See [Filters reference](/providers/terraform-viettelidc/vcloud/latest/docs/guides/data_source_filters) for details and examples.
