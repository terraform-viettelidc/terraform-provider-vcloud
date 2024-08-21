---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_independent_disk"
sidebar_current: "docs-vcloud-independent-disk"
description: |-
  Provides a Viettel IDC Cloud independent disk resource. This can be used to create and delete independent disks.
---

# vcloud\_independent\_disk

Provides a Viettel IDC Cloud independent disk resource. This can be used to create and delete independent disks.
The resource is capable of updating independent disks attached to a VM. Update detaches the disks temporarily and attaches back after changes are done.

## Example Usage

```hcl
resource "vcloud_independent_disk" "myNewIndependentDisk" {
  vdc             = "my-vcloud"
  name            = "logDisk"
  size_in_mb      = "1024"
  bus_type        = "SCSI"
  bus_sub_type    = "VirtualSCSI"
  storage_profile = "external"
}

resource "vcloud_vapp_vm" "web2" {
  vapp_name = vcloud_vapp.web.name

  # ...

  disk {
    name        = vcloud_independent_disk.myNewIndependentDisk.name
    bus_number  = 1
    unit_number = 0
  }

}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) Disk name
* `size_in_mb` - (Required, *v3.0+*) Size of disk in MB.
* `bus_type` - (Optional) Disk bus type. Values can be: `IDE`, `SCSI`, `SATA`, (*v3.6+*) `NVME`. **Note** When the disk type is IDE then VM is required to be powered off
* `bus_sub_type` - (Optional) Disk bus subtype. Values can be: `buslogic`, `lsilogic`, `lsilogicsas`, `VirtualSCSI` for `SCSI`, `ahci` for `SATA` and (*v3.6+*) `nvmecontroller` for `NVME`
* `storage_profile` - (Optional) The name of storage profile where disk will be created
* `sharing_type` - (Optional, *v3.6+* and Vcloud 10.2+) This is the sharing type. Values can be: `DiskSharing`,`ControllerSharing`, or `None`
* `metadata` - (Deprecated; *v3.6+*) Use `metadata_entry` instead. Key value map of metadata to assign to this independent disk.
* `metadata_entry` - (Optional; *v3.8+*) A set of metadata entries to assign. See [Metadata](#metadata) section for details.

## Attribute reference

Supported in provider *v2.5+*

* `iops` - (Computed) IOPS request for the created disk
* `owner_name` - (Computed) The owner name of the disk
* `datastore_name` - (Computed) Data store name. Readable only for system user.
* `is_attached` - (Computed) True if the disk is already attached
* `encrypted` - (Computed, *v3.6+* and Vcloud 10.2+) True if disk is encrypted
* `uuid` - (Computed, *v3.6+* and Vcloud 10.2+) The UUID of this named disk's device backing
* `attached_vm_ids` - (Computed, *v3.6+*) IDs of VM which are using the disk

<a id="metadata"></a>
## Metadata

The `metadata_entry` (*v3.8+*) is a set of metadata entries that have the following structure:

* `key` - (Required) Key of this metadata entry.
* `value` - (Required) Value of this metadata entry.
* `type` - (Required) Type of this metadata entry. One of: `MetadataStringValue`, `MetadataNumberValue`, `MetadataDateTimeValue`, `MetadataBooleanValue`.
* `user_access` - (Required) User access level for this metadata entry. One of: `PRIVATE` (hidden), `READONLY` (read only), `READWRITE` (read/write).
* `is_system` - (Required) Domain for this metadata entry. true if it belongs to `SYSTEM`, false if it belongs to `GENERAL`.

~> Note that `is_system` requires System Administrator privileges, and not all `user_access` options support it.
   You may use `is_system = true` with `user_access = "PRIVATE"` or `user_access = "READONLY"`.

Example:

```hcl
resource "vcloud_independent_disk" "example" {
  # ...
  metadata_entry {
    key         = "foo"
    type        = "MetadataStringValue"
    value       = "bar"
    user_access = "PRIVATE"
    is_system   = true # Requires System admin privileges
  }

  metadata_entry {
    key         = "myBool"
    type        = "MetadataBooleanValue"
    value       = "true"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

To remove all metadata one needs to specify an empty `metadata_entry`, like:

```
metadata_entry {}
```

The same applies also for deprecated `metadata` attribute:

```
metadata = {}
```

## Importing

Supported in provider *v2.5+*

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing independent disk can be [imported][docs-import] into this resource via supplying its path.
The path for this resource is made of org-name.vdc-name.disk-id
For example, using this structure, representing a independent disk that was **not** created using Terraform:

```hcl
resource "vcloud_independent_disk" "tf-myDisk" {
  vdc  = "my-vdc"
  name = "my-disk"
}
```

You can import such independent disk into terraform state using this command

```
terraform import vcloud_independent_disk.tf-myDisk org-name.vdc-name.my-disk-id
```

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see the rest of the values and modify the script accordingly for
further operations.

### Listing independent disk IDs

If you want to list IDs there is a special command **`terraform import vcloud_independent_disk.imported list@org-name.vdc-name.my-independent-disk-name`**
or **`terraform import vcloud_independent_disk.imported list@org-name.vdc-name`**
where `org-name` is the organization used, `vdc-name` is vDC name and `my-independent-disk-name`
is independent disk name. The output for this command should look similar to the one below:

```shell
$ terraform import vcloud_independent_disk.imported list@org-name.vdc-name.my-independent-disk-name
vcloud_independent_disk.Disk_import: Importing from ID "list@org-name.vdc-name.my-independent-disk-name"...
Retrieving all disks by name
No  ID                                                      Name    Description Size
--  --                                                      ----    ------      ----
1  urn:vcloud:disk:1bbc273d-7701-4f06-97be-428b46b0805e     diskV2  loging      78946548
2  urn:vcloud:disk:6e1c996f-48b8-4e78-8111-a6407188d8b6     diskV2              5557452

Error: resource was not imported! resource id must be specified in one of these formats:
'org-name.vdc-name.my-independent-disk-id' to import by rule id
'list@org-name.vdc-name.my-independent-disk-name' to get a list of disks with their IDs
```

Now to import disk with ID urn:vcloud:disk:1bbc273d-7701-4f06-97be-428b46b0805e one could supply this command:

```shell
$ terraform import vcloud_independent_disk.imported list@org-name.vdc-name.urn:vcloud:disk:1bbc273d-7701-4f06-97be-428b46b0805e
```
