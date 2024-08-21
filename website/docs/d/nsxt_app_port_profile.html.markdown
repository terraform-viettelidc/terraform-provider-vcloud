---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_app_port_profile"
sidebar_current: "docs-vcloud-data-source-nsxt-app-port-profile"
description: |-
  Provides a data source to read NSX-T Application Port Profiles. Application Port Profiles include 
  a combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
  services on the Edge Gateway.
---

# vcloud\_nsxt\_app\_port\_profile

Supported in provider *v3.3+* and Vcloud 10.1+ with NSX-T backed VDCs.

Provides a data source to read NSX-T Application Port Profiles. Application Port Profiles include a
combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
services on the Edge Gateway.

## Example Usage 1 (Find an Application Port Profile defined by Provider)

```hcl
data "vcloud_nsxt_app_port_profile" "custom" {
  org        = "System"
  context_id = data.vcloud_nsxt_manager.first.id
  name       = "WINS"
  scope      = "PROVIDER"
}
```

## Example Usage 2 (Find an Application Port Profile defined by Tenant in a VDC Group)

```hcl
data "vcloud_vdc_group" "g1" {
  org  = "myOrg"
  name = "myVDC"
}

data "vcloud_nsxt_app_port_profile" "custom" {
  org        = "my-org"
  context_id = data.vcloud_vdc_group.g1.id
  name       = "SSH-custom"
  scope      = "TENANT"
}
```

## Example Usage 3 (Find a System defined Application Port Profile)

```hcl
data "vcloud_org_vdc" "vdc1" {
  org  = "myOrg"
  name = "myVDC"
}

data "vcloud_nsxt_app_port_profile" "custom" {
  context_id = data.vcloud_org_vdc.vdc1.id

  scope = "SYSTEM"
  name  = "SSH"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Deprecated; Optional) The name of VDC to use, optional if defined at provider level.
  Deprecated and replaced by `context_id`
* `context_id` - (Optional) ID of NSX-T Manager, VDC or VDC Group. Replaces deprecated field `vdc`. Required if using more than one NSX-T Manager.
* `name` - (Required)  - Unique name of existing Security Group.
* `scope` - (Required)  - `SYSTEM`, `PROVIDER`, or `TENANT`.

## Attribute Reference

All the arguments and attributes defined in
[`vcloud_nsxt_app_port_profile`](/providers/vmware/vcloud/latest/docs/resources/nsxt_app_port_profile) resource
are available.
