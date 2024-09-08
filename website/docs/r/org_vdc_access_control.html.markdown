---
layout: "vcd"
page_title: "Viettel IDC Cloud: org_vdc_access_control"
sidebar_current: "docs-vcd-resource-org-vdc-access-control"
description: |-
Provides a Viettel IDC Cloud Org VDC access control resource. This can be
used to share VDC across users or groups.
---

# vcd\_org\_vdc\_access\_control

Provides a Viettel IDC Cloud Org VDC access control resource. This can be
used to share VDC across users and/or groups.

Supported in provider *v3.7+*

-> **Note:** This resource requires either system or org administrator privileges.

## Example Usage

### Example Usage 1 (Giving VDC read only access to a couple of users)
```hcl
data "vcloud_org_user" "my-user" {
  org  = "my-org" # Optional
  name = "my-user"
}

data "vcloud_org_user" "my-user2" {
  org  = "my-org" # Optional
  name = "my-user2"
}

resource "vcloud_org_vdc_access_control" "my_access_control" {
  org                  = "my-org" # Optional
  vdc                  = "my-vdc" # Optional
  shared_with_everyone = false

  shared_with {
    user_id      = data.vcloud_org_user.my-user.id
    access_level = "ReadOnly"
  }

  shared_with {
    user_id      = data.vcloud_org_user.my-user2.id
    access_level = "ReadOnly"
  }
}
```

### Example Usage 2 (Giving VDC read only access to everybody)
```hcl
resource "vcloud_org_vdc_access_control" "my_access_control" {
  org                   = "my-org" # Optional
  vdc                   = "my-vdc" # Optional
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
```

### Example Usage 3 (Creating a VDC and setting VDC read only access to everybody)
```hcl
resource "vcloud_org_vdc" "my_vdc" {
  name = "my-vdc" # Optional
  org  = "my-org" # Optional

  allocation_model  = "Flex"
  network_pool_name = "my-network-pool"
  provider_vdc_name = "my-provider-vdc"

  compute_capacity {
    cpu {
      allocated = "1024"
      limit     = "1024"
    }

    memory {
      allocated = "1024"
      limit     = "1024"
    }
  }

  storage_profile {
    name    = "my-storage-profile"
    enabled = true
    limit   = 10240
    default = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity                 = false
  include_vm_memory_overhead = false
}

resource "vcloud_org_vdc_access_control" "my_access_control" {
  org                   = "my-org" # Optional
  vdc                   = "my-vdc" # Optional
  shared_with_everyone  = true
  everyone_access_level = "ReadOnly"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organizations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `shared_with_everyone` - (Required) Whether the VDC is shared with everyone.
* `everyone_access_level` - (Optional) Access level when the VDC is shared with everyone (only `ReadOnly` is available). Required when shared_with_everyone is set.
* `shared_with` - (Optional) one or more blocks defining a subject to which we are sharing.
  See [shared_with](#shared_with) below for detail. It cannot be used if `shared_with_everyone` is set.

~> **Note:** Users must either set sharing for everybody using `shared_with_everyone` and `everyone_access_level` arguments or per user/group access using `shared_with` argument. Setting both will make the resource to error.

<a id="shared_with"></a>
## shared_with

* `user_id` - (Optional) The ID of a user which we are sharing with. Required if `group_id` is not set.
* `group_id` - (Optional) The ID of a group which we are sharing with. Required if `user_id` is not set.
* `access_level` - (Required) The access level for the user or group to which we are sharing. (Only `ReadOnly` is available)

## Attribute reference

* `shared_with.subject_name` - The name of the subject (group or user) which we are sharing with.

# Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing VDC access control can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name
An example is below:

```
terraform import vcloud_org_vdc_access_control.my_access_control my-org.my-vdc
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCLOUD_IMPORT_SEPARATOR


[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the security tag as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the security tag stored properties.
