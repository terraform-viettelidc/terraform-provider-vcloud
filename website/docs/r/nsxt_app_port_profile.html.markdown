---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxt_app_port_profile"
sidebar_current: "docs-vcloud-resource-nsxt-app-port-profile"
description: |-
  Provides a resource to manage NSX-T Application Port Profiles. Application Port Profiles include a
  combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
  services on the Edge Gateway. In addition to the default Port Profiles that are preconfigured for
  NSX-T Data Center, you can create custom Application Port Profiles.
---

# vcloud\_nsxt\_app\_port\_profile

Supported in provider *v3.3+* and Vcloud 10.1+ with NSX-T backed VDCs.

Provides a resource to manage NSX-T Application Port Profiles. Application Port Profiles include a
combination of a protocol and a port, or a group of ports, that is used for Firewall and NAT
services on the Edge Gateway. In addition to the default Port Profiles that are preconfigured for
NSX-T Data Center, you can create custom Application Port Profiles.

## Example Usage 1 (Define Provider wide Application Port Profile)

```hcl
resource "vcloud_nsxt_app_port_profile" "icmpv4" {
  org         = "System"
  name        = "ICMP custom profile"
  description = "Application port profile for ICMPv4"

  scope      = "PROVIDER"
  context_id = data.vcloud_nsxt_manager.first.id

  app_port {
    protocol = "ICMPv4"
  }
}
```

## Example Usage 2 (Define Application Port Profile for particular NSX-T VDC 'vdc1')
```hcl
data "vcloud_org_vdc" "v1" {
  org  = "my-org"
  name = "vdc1"
}

resource "vcloud_nsxt_app_port_profile" "custom-app" {
  org        = "my-org"
  context_id = data.vcloud_org_vdc.v1.id

  name        = "custom app profile"
  description = "Application port profile for custom application"

  scope = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }

  app_port {
    protocol = "TCP"
    port     = ["2000", "2010-2020", "12345", "65000"]
  }

  app_port {
    protocol = "UDP"
    port     = ["40000-60000"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
  when connected as sysadmin working across different organisations.
* `vdc` - (Deprecated; Optional) The name of VDC to use, optional if defined at provider level.
  Deprecated and replaced by `context_id`
* `context_id` - (Optional) ID of NSX-T Manager, VDC or VDC Group. Replaces deprecated fields `vdc`
  and `nsxt_manager_id`. It accepts VDC, VDC Group or NSX-T Manager ID. 
* `name` - (Required) A unique name for Security Group
* `scope` - (Required) Application Port Profile scope - `PROVIDER`, `TENANT`
* `nsxt_manager_id` - (Deprecated; Optional) Required only when `scope` is `PROVIDER`. Deprecated
  and replaced by `context_id`
* `app_port` - (Required) At least one block of [Application Port definition](#app-port)


<a id="app-port"></a>
## Application Port

Each Application port must have at least `protocol` and optionally `port`:

* `protocol` - (Required) One of protocols `ICMPv4`, `ICMPv6`, `TCP`, `UDP`
* `port` - (Optional) A set of port numbers or port ranges (e.g. `"10000"`, `"20000-20010"`)


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

There are 2 different import paths based on `scope`:
* `PROVIDER` scoped import path is:
```
terraform import vcloud_nsxt_app_port_profile.imported my-nsxt-manager-name.my-app-port-profile-name
```
This would import NSX-T Application Port Profile named `my-app-port-profile-name` defined in NSX-T manager
named `my-nsxt-manager-name`.

->  Only **System** user can manage `Provider` scoped NSX-T Application Port Profile.

* `TENANT` scoped import path is:
```
terraform import vcloud_nsxt_app_port_profile.imported my-org.my-nsxt-vdc-or-vdc-group.my-app-port-profile-name
```

This would import NSX-T Application Port Profile named `my-app-port-profile-name` defined in Org `my-org` and NSX-T
VDC/VDC Group - `my-nsxt-vdc-or-vdc-group`
