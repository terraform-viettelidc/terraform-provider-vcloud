---
layout: "vcd"
page_title: "VMware Cloud Director: vcloud_org_vdc_template_instance"
sidebar_current: "docs-vcd-resource-org-vdc-template-instance"
description: |-
  Provides a resource to instantiate VDCs from a VDC Template in VMware Cloud Director.
---

# vcd\_org\_vdc\_template\_instance

Provides a resource to instantiate VDCs from a [VDC Template](/providers/vmware/vcd/latest/docs/resources/org_vdc_template) in VMware Cloud Director.
Supported in provider *v3.13+*

## Example Usage

```hcl
data "vcloud_org" "org" {
  name = "my_org"
}

data "vcloud_provider_vdc" "pvdc1" {
  name = "nsxTPvdc1"
}

data "vcloud_provider_vdc" "pvdc2" {
  name = "nsxTPvdc2"
}

data "vcloud_external_network_v2" "ext_net" {
  name = "nsxt-extnet"
}

data "vcloud_network_pool" "np1" {
  name = "NSX-T Overlay 1"
}

resource "vcloud_org_vdc_template" "tmpl" {
  name               = "myTemplate"
  tenant_name        = "myAwesomeTemplate"
  description        = "Requires System privileges"
  tenant_description = "Any tenant can use this"
  allocation_model   = "AllocationVApp"

  compute_configuration {
    cpu_limit         = 0
    cpu_guaranteed    = 20
    cpu_speed         = 256
    memory_limit      = 1024
    memory_guaranteed = 30
  }

  provider_vdc {
    id                  = data.vcloud_provider_vdc.pvdc1.id
    external_network_id = data.vcloud_external_network_v2.ext_net.id
  }

  provider_vdc {
    id                  = data.vcloud_provider_vdc.pvdc2.id
    external_network_id = data.vcloud_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "*"
    default = true
    limit   = 1024
  }

  network_pool_id = data.vcloud_network_pool.np1.id

  readable_by_org_ids = [
    data.vcloud_org.org.id
  ]
}

resource "vcloud_org_vdc_template_instance" "my_instance" {
  org_vdc_template_id = vcloud_org_vdc_template.tmpl.id
  name                = "myInstantiatedVdc"
  description         = "A new VDC"
  org_id              = data.vcloud_org.org.id

  # This guarantees that removing this resource from HCL won't remove
  # the instantiated VDC. Set it to "true" to remove the VDC when this
  # resource is removed.
  delete_instantiated_vdc_on_removal = false
  delete_force                       = false
  delete_recursive                   = false
}
```

## Argument Reference

The following arguments are supported:

* `org_vdc_template_id` - (Required) The ID of the VDC Template to instantiate
* `name` - (Required) Name to give to the instantiated Organization VDC
* `description` - (Optional) Description of the instantiated Organization VDC
* `org_id` - (Required) ID of the Organization where the VDC will be instantiated
* `delete_instantiated_vdc_on_removal` - (Required) If this flag is set to `true`, removing this resource will attempt to delete the instantiated VDC
* `delete_force` - (Optional) Defaults to `false`. If this flag is set to `true`, it forcefully deletes the VDC, only when `delete_instantiated_vdc_on_removal=true`
* `delete_recursive` - (Optional) Defaults to `false`. If this flag is set to `true`, it recursively deletes the VDC, only when `delete_instantiated_vdc_on_removal=true`

## Attribute Reference

There are no read-only attributes. However, after the `vcloud_org_vdc_template_instance` resource is created successfully,
the identifier of the new VDC is saved in the Terraform state, as the `id` of the `vcloud_org_vdc_template_instance` resource
(example: `vcloud_org_vdc_template_instance.my_instance.id`).

## Deletion of the vcd\_org\_vdc\_template\_instance resource

When configuring the `vcloud_org_vdc_template_instance`, one must set the required `delete_instantiated_vdc_on_removal` argument.

* When set to `true`, removing this resource will attempt to delete the VDC that it instantiated.

  The flags `delete_force` and `delete_recursive` should be considered in this scenario, as they behave the same way as in [`vcloud_org_vdc`](/providers/vmware/vcd/latest/docs/resources/org_vdc).

* When set to `false`, removing this resource will leave the instantiated VDC behind. This is useful when the VDC is being managed
by Terraform after importing it to a `vcloud_org_vdc` (see section below), therefore this resource is not needed anymore.

-> When changing `delete_instantiated_vdc_on_removal`, `delete_force` or `delete_recursive`, take into account that you need to perform a `terraform apply` to
save the changes in these flags.

## How to manage the VDC instantiated from a VDC Template using Terraform

If users want to modify the new instantiated VDC, they can [import](/providers/vmware/vcd/latest/docs/guides/importing_resources#semi-automated-import-terraform-v15) it.
In the same `.tf` file (once the VDC has been instantiated), or in a new one, we can place the following snippet: 

```hcl
import {
  to = vcloud_org_vdc.imported
  id = "my_org.myInstantiatedVdc" # Using the same names from the example
}
```

Note that this importing mechanism still does not support `${}` placeholders, so the Organization and VDC name must be explicitly
written. When running the `terraform plan -generate-config-out=generated_resources.tf`, Terraform will generate the new file
`generated_resources.tf` with the instantiated VDC code.

With a subsequent `terraform apply`, the instantiated VDC will be managed by Terraform as a normal `vcloud_org_vdc` resource.

-> After importing, bear in mind that `vcloud_org_vdc` will have the arguments `delete_force` and `delete_recursive` set to `false`.
They should be modified accordingly.

## Importing

There is no importing for this resource, as it should be used only on creation.
The instantiated VDC can be imported using `vcloud_org_vdc` by following the steps of the section above.
