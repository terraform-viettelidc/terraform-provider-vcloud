---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_nsxv_distributed_firewall"
sidebar_current: "docs-vcloud-resource-nsxv-distributed-firewall"
description: |-
  The NSX-V Distributed Firewall allows user to segment organization virtual data center entities, such as
  virtual machines, edges, networks, based on several attributes
---

# vcloud\_nsxv\_distributed\_firewall

The Distributed Firewall allows user to segment organization virtual data center entities, such as
virtual machines, edges, networks, based on several attributes. This resource handles only **layer 3** rules.

Supported in provider *v3.9+*

## Example Usage

```hcl
data "vcloud_org_vdc" "my-vdc" {
  org  = "my-org"
  name = "my-vdc"
}

data "vcloud_nsxv_ip_set" "test-ipset" {
  org  = "my-org"
  vdc  = "my-vdc"
  name = "TestIpSet"
}

data "vcloud_vapp_vm" "vm1" {
  vdc       = data.vcloud_org_vdc.my-vdc.name
  vapp_name = "TestVapp"
  name      = "TestVm"
}

data "vcloud_network_routed" "net-r" {
  vdc  = data.vcloud_org_vdc.my-vdc.name
  name = "net-routed"
}

data "vcloud_edgegateway" "edge" {
  vdc  = data.vcloud_org_vdc.my-vdc.name
  name = "my-edge"
}

data "vcloud_nsxv_application" "application1" {
  vdc_id = data.vcloud_org_vdc.my-vdc.id
  name   = "POP3"
}

data "vcloud_nsxv_application_group" "application_group1" {
  vdc_id = data.vcloud_org_vdc.my-vdc.id
  name   = "MS Exchange 2010 Mailbox Servers"
}

resource "vcloud_nsxv_distributed_firewall" "dfw1" {
  vdc_id = data.vcloud_org_vdc.my-vdc.id

  rule {
    name      = "third"
    direction = "inout"
    action    = "allow"

    # Using an IP set as source
    source {
      name  = data.vcloud_nsxv_ip_set.test-ipset.name
      value = data.vcloud_nsxv_ip_set.test-ipset.id
      type  = "IPSet"
    }

    # Using an anonymous application
    application {
      protocol         = "TCP"
      source_port      = "20250"
      destination_port = "20251"
    }

    # Using a named application
    application {
      name  = data.vcloud_nsxv_application.application1.name
      value = data.vcloud_nsxv_application.application1.id
      type  = "Application"
    }

    # Using a named application group
    application {
      name  = data.vcloud_nsxv_application_group.application_group1.name
      value = data.vcloud_nsxv_application_group.application_group1.id
      type  = "ApplicationGroup"
    }

    # Applied to an edge gateway
    applied_to {
      name  = data.vcloud_edgegateway.edge.name
      type  = "Edge"
      value = data.vcloud_edgegateway.edge.id
    }
  }

  rule {
    name      = "second"
    direction = "inout"
    action    = "allow"

    # Defining a literal source
    source {
      name  = "10.10.1.0-10.10.1.100"
      value = "10.10.1.0-10.10.1.100"
      type  = "Ipv4Address"
    }

    # Defining a VM as source
    source {
      name  = data.vcloud_vapp_vm.vm1.name
      value = data.vcloud_vapp_vm.vm1.id
      type  = "VirtualMachine"
    }

    # Using a routed network as destination
    destination {
      name  = data.vcloud_network_routed.net-r.name
      value = data.vcloud_network_routed.net-r.id
      type  = "Network"
    }

    # Using an isolated network as destination
    destination {
      name  = data.vcloud_network_isolated.net-i.name
      value = data.vcloud_network_isolated.net-i.id
      type  = "Network"
    }

    # Applied to the current VDC
    applied_to {
      name  = data.vcloud_org_vdc.my-vdc.name
      type  = "VDC"
      value = data.vcloud_org_vdc.my-vdc.id
    }
  }

  # This rule is the main "deny-all" rule
  rule {
    name      = "first"
    direction = "inout"
    action    = "deny"

    # No source, destination, application: will be interpreted as `any`

    # Applied to the current VDC
    applied_to {
      name  = data.vcloud_org_vdc.my-vdc.name
      type  = "VDC"
      value = data.vcloud_org_vdc.my-vdc.id
    }
  }
}
```

~> NOTE: when we remove a `vcloud_nsxv_distributed_firewall`, the firewall gets disabled, and all rules (both layer 3 and layer 2) are removed.

## Argument Reference

The following arguments are supported:

* `vdc_id` - (Required) The ID of VDC to manage the Distributed Firewall in. Can be looked up using a `vcloud_org_vdc` data source
* `rule` - (Optional) One or more blocks with [Firewall Rule](#firewall-rule) definitions. **Order
  defines firewall rule precedence**. If no rules are defined, all will be removed from the firewall

## Attribute Reference

* `enabled` - Shows whether the NSX-V Distributed Firewall is enabled.

## Firewall Rule

-> Order of `rule` blocks defines order of firewall rules in the system.

Each Firewall Rule contains the following attributes:

* `name` - (Optional) Explanatory name for firewall rule (uniqueness not enforced)
* `direction` - (Required) One of `in`, `out`, or `inout`
* `action` - (Required) Defines if it should `allow` or `deny` traffic 
* `enabled` - (Optional) Defines if the rule is enabled (default `true`)
* `logging` - (Optional) Defines if logging for this rule is enabled (default `false`)
* `source` - (Optional) A set of source objects. See below for [source or destination objects](#source-or-destination-objects)
Leaving it empty matches `any` (all)
* `destination` - (Optional) A set of destination objects. See below for [source or destination objects](#source-or-destination-objects). Leaving it empty matches `any` (all)
* `application` - (Optional) An optional set of applications to use for this rule. See below for [Application objects](#application-objects)
* `applied_to` - (Required) A set of objects to which the rule applies. See below for [Source or destination objects](#source-or-destination-objects) 
* `exclude_source` - (Optional) - reverses value of `source` for the rule to match everything except specified objects
* `exclude_destination` - (Optional) - reverses value of `destination` for the rule to match everything except specified objects

### Source or destination objects

Each element of the `source`, `destination`, or `applied_to` is identified by three elements:

* `name` - (Required) is the name of the object. When using a literal object (such as an IP or IP range), **the name must
  contain the same text as the `value`**
* `type` - (Required) is the type of the object. One of `Network`, `Edge`, `VirtualMachine`, `IPSet`, `VDC`, `Ipv4Address`.
   Note that the case of the type identifiers are relevant. Using `IpSet` instead of `IPSet` results in an error.
   Also note that `Ipv4Address` allows any of:
    * An IP address (example: `192.168.1.1`)
    * A list of IP addresses (example: `192.168.1.2,192.168.1.15`)
    * A range of IP addresses (example: `10.10.10.2-10.10.10.20`)
    * A CIDR (example: `10.10.10.1/24`)
* `value` - (Required) - When using a named object (such a VM or a network), this field will have the object ID. For a literal
   object, such as an IP or IP range, this will be the text of the IP reference.

### Application objects

An application object can be one of the three following things:

* A named application, identified by fields `name` and `value` with `type = "Application"`
* A named application group, identified by fields `name` and `value` with `type = "ApplicationGroup"`
* A literal application, identified by fields `protocol`, `ports`, `source_port`, `destination_port`

The following fields can be used:

* Named applications:
  * `name` (Optional) - Required if defining a named object or object group
  * `type` (Optional) - Required if defining a named object or object group. (One of `Application` or `ApplicationGroup`)
  * `value` (Optional) - Required if defining a named object or object group

* Literal applications:
  * `protocol` (Required) - Required when defining a literal object. (One of `TCP`, `UDP`, `ICMP`)
  * `ports` (Optional) - The ports used by the application. Could be a single port, a comma delimited list, or a range
  * `source_port` (Optional) - The source port used by the application, if any
  * `destination_port` (Optional) - The destination port used by the application, if any

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing NSX-V Distributed Firewall can be [imported][docs-import] into this resource via supplying its path.
The path for this resource can be one of:
* vdc-ID
* org-name.vdc-name

For example, using the structure in [Example Usage](#example-usage), representing a firewall that was **not** created using Terraform:

You can import such firewall into terraform state using one of the commands below:

```
terraform import vcloud_nsxv_distributed_firewall.dfw1 urn:vcloud:vdc:e5680ceb-1c15-48a8-9a54-e0bbc6fe909f
# or
terraform import vcloud_nsxv_distributed_firewall.dfw1 my-org.my-vdc
```

NOTE 1: To get the VDC ID, you can use a `vcloud_org_vdc` data source, and check its ID from the Terraform state file (`terraform.tfstate`).

NOTE 2: The default separator (.) can be changed using `Provider.import_separator` or the environmant variable `VCLOUD_IMPORT_SEPARATOR`

[docs-import]:https://www.terraform.io/docs/import/

After importing, if you run `terraform plan` you will see whether you need to change something in the script to match
the existing structure.
