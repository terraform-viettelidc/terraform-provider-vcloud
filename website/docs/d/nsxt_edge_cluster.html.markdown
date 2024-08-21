---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_edge_cluster"
sidebar_current: "docs-vcloud-data-source-nsxt-edge-cluster"
description: |-
  Provides a data source for available NSX-T Edge Clusters.
---

# vcloud\_nsxt\_edge\_cluster

Provides a data source for available NSX-T Edge Clusters.

Supported in provider *v3.1+*

-> **Note:** This data source uses new Viettel IDC Cloud
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least Vcloud *10.1.1+* and NSX-T *3.0+*.

## Example Usage (with VDC ID)

```hcl
data "vcloud_org_vdc" "existing" {
  org  = "my-org"
  name = "nsxt-vdc-1"
}

data "vcloud_nsxt_edge_cluster" "first" {
  org    = "my-org"
  vdc_id = data.vcloud_org_vdc.existing.id
  name   = "edge-cluster-one"
}
```

## Example Usage (with VDC Group ID)

```hcl
data "vcloud_vdc_group" "existing" {
  org  = "my-org"
  name = "nsxt-vdc-group-1"
}

data "vcloud_nsxt_edge_cluster" "first" {
  org          = "my-org"
  vdc_group_id = data.vcloud_vdc_group.existing.id
  name         = "edge-cluster-one"
}
```

## Example Usage (with Provider VDC ID)

```hcl
data "vcloud_provider_vdc" "nsxt-pvdc" {
  name = "nsxt-provider-vdc"
}

data "vcloud_nsxt_edge_cluster" "first" {
  org             = "my-org"
  provider_vdc_id = data.vcloud_provider_vdc.nsxt-pvdc.id
  name            = "edge-cluster-one"
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which edge cluster belongs. Optional if defined at provider level.
* `vdc` - (Optional, Deprecated) The name of VDC that owns the edge cluster. Optional if defined at provider level.
* `vdc_id` - (Optional, *v3.8+*, *Vcloud 10.3+*) The ID of VDC for lookup. Data source `vcloud_org_vdc` can be used to get ID.
* `vdc_group_id` - (Optional, *v3.8+*, *Vcloud 10.3+*) The ID of VDC Group for lookup. Data source `vcloud_vdc_group` can be used to get ID.
* `provider_vdc_id` - (Optional, *v3.8+*, *Vcloud 10.3+*) The ID of VDC Group for lookup. Data source `vcloud_provider_vdc` can be used to get ID.
* `name` - (Required) NSX-T Edge Cluster name. **Note.** NSX-T does allow to have duplicate names therefore to be able
to correctly use this data source there should not be multiple NSX-T Edge Clusters with the same name defined.

## Attribute reference

* `description` - Edge Cluster description in NSX-T manager.
* `node_count` - Number of nodes in Edge Cluster.
* `node_type` - Type of nodes in Edge Cluster.
* `deployment_type` - Deployment type of Edge Cluster.
