---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_vm"
sidebar_current: "docs-vcloud-resource-vm"
description: |-
  Provides a Viettel IDC Cloud standalone VM resource. This can be used to create, modify, and delete Standalone VMs.
---

# vcloud\_vm

Provides a Viettel IDC Cloud standalone VM resource. This can be used to create, modify, and delete Standalone VMs.

Supported in provider *v3.2+*

## Example Usage

```hcl
data "vcloud_catalog" "cat-datacloud" {
  org  = "test"
  name = "cat-datacloud"
}

data "vcloud_catalog_vapp_template" "photon-hw11" {
  org        = "test"
  catalog_id = data.vcloud_catalog.cat-datacloud.id
  name       = "photon-hw11"
}

resource "vcloud_vm" "TestVm" {
  name = "TestVm"

  vapp_template_id = data.vcloud_catalog_vapp_template.photon-hw11.id
  cpus             = 2
  memory           = 2048

  network {
    name               = "net-datacloud-r"
    type               = "org"
    ip_allocation_mode = "POOL"
  }
}
```

## Arguments and attributes reference

This resource provides all arguments and attributes available for [`vcloud_vapp_vm`](/providers/vmware/vcloud/latest/docs/resources/vapp_vm),
with the only difference that the `vapp_name` should be left empty.

General notes:

* Although from the UI standpoint a standalone VM appears to exist without a vApp, in reality there is a hidden vApp that
  is generated automatically when the VM is created, and removed when the VM is terminated. The field `vapp_name` is populated
  with the hidden vApp name, and readable in Terraform state.

* The import path of the standalone VM does not need a vApp name. While a standard VM is retrieved with a path like 
`org-name.vdc-name.vapp-name.vm-name`, for a standalone VM you can use `org-name.vdc-name.vm-name`. If you know the vApp
  name (as retrieved through a data source, for example), you can safely use it in the path, as if it were a `vcloud_vapp_vm`.

* The VM name is unique **within the vApp**, which means that it is possible to create multiple standalone VMs with the same name.
  This fact has consequences when importing a resource, where we identify the VM by name. If there are duplicates, we get
  an error message containing the list of VMs that share the same name. To retrieve a specific VM in such scenario, we need
  to provide the VM ID instead of the name.

For example, given this input
```hcl
resource "vcloud_vm" "TestVm" {
  org  = "datacloud"
  vdc  = "vdc-datacloud"
  name = "TestVm"
}
```

If multiple VMs have the same name, we get an error like:

```
$ terraform import vcloud_vm.TestVm datacloud.vdc-datacloud.TestVm
vcloud_vm.TestVm: Importing from ID "datacloud.vdc-datacloud.TestVm"...

Error: [VM import] error retrieving VM TestVm by name: more than one VM found with name TestVm
ID                                                 Guest OS                       Network
-------------------------------------------------- ------------------------------ --------------------
urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003 Other 3.x Linux (64-bit)       (net-datacloud-r - 192.168.2.2)
urn:vcloud:vm:41d5d5a7-040e-49cb-a516-5a604211a395 Debian GNU/Linux 10 (32-bit)   (net-datacloud-i - DHCP)
```

We can achieve the goal by providing the ID instead of the name

```
$ terraform import vcloud_vm.TestVm datacloud.vdc-datacloud.urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003
vcloud_vm.TestVm: Importing from ID "datacloud.vdc-datacloud.urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003"...
vcloud_vm.TestVm: Import prepared!
  Prepared vcloud_vm for import
vcloud_vm.TestVm: Refreshing state... [id=urn:vcloud:vm:26c04f4d-2185-4a33-8ef9-019768d29003]

Import successful!
```
