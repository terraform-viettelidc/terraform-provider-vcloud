---
layout: "vcloud"
page_title: "Viettel IDC Cloud: Container Service Extension v3.1.x"
sidebar_current: "docs-vcloud-guides-cse-3-1-x"
description: |-
  Provides guidance on configuring Vcloud to be able to install Container Service Extension v3.1.x.
---

# Container Service Extension v3.1.x

~> This CSE installation method is **deprecated** in favor of CSE v4.x. Please have a look at the new guide
[here](https://registry.terraform.io/providers/vmware/vcloud/latest/docs/guides/container_service_extension_4_x_install)

## About

This guide describes the required steps to configure Vcloud to install the Container Service Extension (CSE) v3.1.x, that
will allow tenant users to deploy **Tanzu Kubernetes Grid Multi-cloud (TKGm)** clusters on Vcloud using the UI. For that purpose, after completing the steps described below you
will need also to **publish the Container UI Plugin** to the desired tenants and **run the CSE server** in your infrastructure.

To know more about CSE v3.1.x, you can explore [the official website](https://vmware.github.io/container-service-extension/).

## Pre-requisites

In order to complete the steps described in this guide, please be aware:

* CSE v3.1.x is supported from Vcloud v10.3.1 or above, make sure your Vcloud appliance matches the criteria.
* Terraform provider needs to be v3.8.0 or above.
* All CSE elements use NSX-T backed resources, NSX-V **is not** is supported.
* Some steps require the usage of `cse` extension for `vcloud` CLI. Make sure you have them installed and working.
  Go [here](http://vmware.github.io/vcloud-cli/install.html) for `vcloud` CLI installation,
  and [here](https://vmware.github.io/container-service-extension/cse3_0/INSTALLATION.html#getting_cse) to install `cse`.

## Installation process

-> You can find examples of a fully automated CSE installation in the [Examples](#examples) section below.

To start installing CSE v3.1.x in a Vcloud appliance, you must use **v3.7.0 or above** of the Vcloud Terraform Provider:

```hcl
provider "vcloud" {
  user                 = "administrator"
  password             = var.vcloud_pass
  auth_type            = "integrated"
  sysorg               = "System"
  url                  = var.vcloud_url
  max_retry_timeout    = var.vcloud_max_retry_timeout
  allow_unverified_ssl = var.vcloud_allow_unverified_ssl
}
```

As you will be creating several administrator-scoped resources like Orgs, VDCs, Provider Gateways, etc; make sure you provide 
**System administrator** credentials.

### Step 1: Initialization

This step assumes that you want to install CSE in a brand new [Organization](/providers/vmware/vcloud/latest/docs/resources/org)
with no [VDCs](/providers/vmware/vcloud/latest/docs/resources/org_vdc), or that is a fresh installation of Vcloud.
Otherwise, please skip this step and configure `org` and `vdc` attributes in the provider configuration above or use an
available data source to fetch them.

~> The target VDC needs to be backed by **NSX-T** for CSE to work.

Here is an example that creates both the Organization and the VDC:

```hcl
resource "vcloud_org" "cse_org" {
  name             = "cse_org"
  full_name        = "Organization to deploy Kubernetes clusters with CSE"
  is_enabled       = true
  delete_force     = true
  delete_recursive = true
}

resource "vcloud_org_vdc" "cse_vdc" {
  name = "cse_vdc"
  org  = vcloud_org.cse_org.name

  allocation_model  = "AllocationVApp"
  network_pool_name = "NSX-T Overlay"
  provider_vdc_name = "nsxTPvdc1"
  network_quota     = 50

  compute_capacity {
    cpu {
      limit = 0
    }

    memory {
      limit = 0
    }
  }

  storage_profile {
    name    = "*"
    enabled = true
    limit   = 0
    default = true
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = false
  delete_force             = true
  delete_recursive         = true
}
```

### Step 2: Configure networking

For the Kubernetes clusters to be functional, you need to provide some networking resources to the target VDC:

* [Tier-0 Gateway](/providers/vmware/vcloud/latest/docs/resources/external_network_v2)
* [Edge Gateway](/providers/vmware/vcloud/latest/docs/resources/nsxt_edgegateway)
* [Routed Network](/providers/vmware/vcloud/latest/docs/resources/network_routed_v2)
* [SNAT rule](/providers/vmware/vcloud/latest/docs/resources/nsxt_nat_rule)

The [Tier-0 Gateway](/providers/vmware/vcloud/latest/docs/resources/external_network_v2) will provide access to the
outside world. For example, this will allow cluster users to communicate with Kubernetes API server with **kubectl** and
download required dependencies for the cluster to be created correctly.

Here is an example on how to configure this resource:

```hcl
data "vcloud_nsxt_manager" "main" {
  name = "my-nsxt-manager"
}

data "vcloud_nsxt_tier0_router" "router" {
  name            = "Vcloud T0 edgeCluster"
  nsxt_manager_id = data.vcloud_nsxt_manager.main.id
}

resource "vcloud_external_network_v2" "cse_external_network_nsxt" {
  name        = "extnet-cse"
  description = "NSX-T backed network for k8s clusters"

  nsxt_network {
    nsxt_manager_id      = data.vcloud_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcloud_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "88.88.88.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }
}
```

Create also an [Edge Gateway](/providers/vmware/vcloud/latest/docs/resources/nsxt_edgegateway) that will use the recently created
external network. This will act as the main router connecting our nodes in the internal network to the external (Provider Gateway) network:

```hcl
resource "vcloud_nsxt_edgegateway" "cse_egw" {
  org      = vcloud_org.cse_org.name
  owner_id = vcloud_org_vdc.cse_vdc.id

  name                = "cse-egw"
  description         = "CSE edge gateway"
  external_network_id = vcloud_external_network_v2.cse_external_network_nsxt.id

  subnet {
    gateway       = "88.88.88.1"
    prefix_length = "24"
    primary_ip    = "88.88.88.88"
    allocated_ips {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }
  depends_on = [vcloud_org_vdc.cse_vdc]
}
```

The above resource creates a basic Edge Gateway, but you can of course add more configurations like
[firewall rules](/providers/vmware/vcloud/latest/docs/resources/nsxt_firewall)
to fit with your organization requirements. Make sure that traffic is allowed, as the cluster creation process
requires software to be installed in the nodes, otherwise cluster creation will fail.

Create a [Routed Network](/providers/vmware/vcloud/latest/docs/resources/network_routed_v2) that will be using the recently
created Edge Gateway. This network is the one used by all the Kubernetes nodes in the cluster, so the used IP pool will determine
the number of nodes you can have in the cluster.

```hcl
resource "vcloud_network_routed_v2" "cse_routed" {
  org         = vcloud_org.cse_org.name
  name        = "cse_routed_net"
  description = "My routed Org VDC network backed by NSX-T"

  edge_gateway_id = vcloud_nsxt_edgegateway.cse_egw.id

  gateway       = "192.168.7.0"
  prefix_length = 24

  static_ip_pool {
    start_address = "192.168.7.1"
    end_address   = "192.168.7.100"
  }

  dns1 = "8.8.8.8"
  dns2 = "8.8.8.4"
}
```

To be able to reach the Kubernetes nodes within the routed network, you need also a
[SNAT rule](/providers/vmware/vcloud/latest/docs/resources/nsxt_nat_rule):

```hcl
resource "vcloud_nsxt_nat_rule" "snat" {
  org             = vcloud_org.cse_org.name
  edge_gateway_id = vcloud_nsxt_edgegateway.cse_egw.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  external_address = "88.88.88.89"    # A public IP from the external network
  internal_address = "192.168.7.0/24" # This is the routed network CIDR
  logging          = true
}
```

### Step 3: Configure ALB

Avi Load Balancers are required for CSE to be able to handle Kubernetes services and other internal capabilities.
You need the following resources:

* [ALB Controller](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_controller)
* [ALB Cloud](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_cloud)
* [ALB Service Engine Group](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_service_engine_group)
* [ALB Settings](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_settings)
* [ALB Edge Gateway Service Engine Group](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)
* [ALB Pool](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_pool)
* [ALB Virtual Service](/providers/vmware/vcloud/latest/docs/resources/nsxt_alb_virtual_service)

You can have a look at [this guide](/providers/vmware/vcloud/latest/docs/guides/nsxt_alb) as it explains every resource
and provides some examples of how to set up ALB in Vcloud. You can also have a look at the "[Examples](#examples)" section below
where the full ALB setup is provided.

### Step 4: Create a Service Account

It is **recommended** using a user with CSE Service Role for CSE server management.
The role comes with all the Vcloud rights that CSE needs to function:

```hcl
resource "vcloud_role" "cse-service-role" {
  name        = "CSE Service Role"
  description = "CSE Service Role has all the rights necessary for CSE to operate"

  rights = [
    "Access Control List: View",
    "Access Control List: Manage",
    "AMQP Settings: View",
    "Catalog: Add vApp from My Cloud",
    "Catalog: Create / Delete a Catalog",
    "Catalog: Edit Properties",
    "Catalog: Publish",
    "Catalog: Sharing",
    "Catalog: View ACL",
    "Catalog: View Private and Shared Catalogs",
    "Catalog: View Published Catalogs",
    "Content Library System Settings: View",
    "Custom entity: Create custom entity definitions",
    "Custom entity: Delete custom entity definitions",
    "Custom entity: Edit custom entity definitions",
    "Custom entity: View custom entity definitions",
    "Extension Services: View",
    "Extensions: View",
    "External Service: Manage",
    "External Service: View",
    "General: View Error Details",
    "Group / User: View",
    "Host: View",
    "Kerberos Settings: View",
    "Organization Network: View",
    "Organization vDC Compute Policy: Admin View",
    "Organization vDC Compute Policy: Manage",
    "Organization vDC Compute Policy: View",
    "Organization vDC Kubernetes Policy: Edit",
    "Organization vDC Network: Edit Properties",
    "Organization vDC Network: View Properties",
    "Organization vDC: Extended Edit",
    "Organization vDC: Extended View",
    "Organization vDC: View",
    "Organization: Perform Administrator Queries",
    "Organization: View",
    "Provider Network: View",
    "Provider vDC Compute Policy: Manage",
    "Provider vDC Compute Policy: View",
    "Provider vDC: View",
    "Right: Manage",
    "Right: View",
    "Rights Bundle: View",
    "Rights Bundle: Edit",
    "Role: Create, Edit, Delete, or Copy",
    "Service Configuration: Manage",
    "Service Configuration: View",
    "System Settings: View",
    "Task: Resume, Abort, or Fail",
    "Task: Update",
    "Task: View Tasks",
    "Token: Manage",
    "UI Plugins: Define, Upload, Modify, Delete, Associate or Disassociate",
    "UI Plugins: View",
    "vApp Template / Media: Copy",
    "vApp Template / Media: Create / Upload",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "vApp Template: Checkout",
    "vApp Template: Import",
    "vApp: Allow All Extra Config",
    "vApp: Allow Ethernet Coalescing Extra Config",
    "vApp: Allow Latency Extra Config",
    "vApp: Allow Matching Extra Config",
    "vApp: Allow NUMA Node Affinity Extra Config",
    "vApp: Create / Reconfigure",
    "vApp: Delete",
    "vApp: Edit Properties",
    "vApp: Edit VM CPU and Memory reservation settings in all VDC types",
    "vApp: Edit VM CPU",
    "vApp: Edit VM Compute Policy",
    "vApp: Edit VM Hard Disk",
    "vApp: Edit VM Memory",
    "vApp: Edit VM Network",
    "vApp: Edit VM Properties",
    "vApp: Manage VM Password Settings",
    "vApp: Power Operations",
    "vApp: Shadow VM View",
    "vApp: Upload",
    "vApp: Use Console",
    "vApp: VM Boot Options",
    "vApp: VM Check Compliance",
    "vApp: VM Migrate, Force Undeploy, Relocate, Consolidate",
    "vApp: View VM and VM's Disks Encryption Status",
    "vApp: View VM metrics",
    "vCenter: View",
    "vSphere Server: View",
    "vmware:tkgcluster: Administrator Full access",
    "vmware:tkgcluster: Administrator View",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View"
    # These rights are only needed after CSE is completely installed
    # "cse:nativeCluster: Administrator Full access",
    # "cse:nativeCluster: Administrator View",
    # "cse:nativeCluster: Full Access",
    # "cse:nativeCluster: Modify",
    # "cse:nativeCluster: View"
  ]
}
```

Once created, you can create a [User](/providers/vmware/vcloud/latest/docs/resources/org_user) and use it, as this will provide
more security and traceability to the CSE management operations, which is recommended:

```hcl
resource "vcloud_org_user" "cse-service-account" {
  name     = var.service-account-user
  password = var.service-account-password
  role     = vcloud_role.cse-service-role.name
}
```

To use this user in the subsequent operations, you can configure a new provider with an
[alias](https://www.terraform.io/language/providers/configuration#alias-multiple-provider-configurations):

```hcl
provider "vcloud" {
  alias    = "cse-service-account"
  user     = vcloud_org_user.cse-service-account.name
  password = vcloud_org_user.cse-service-account.password
  # ...
}
```

### Step 5: Configure catalogs and OVAs

You need to have a [Catalog](/providers/vmware/vcloud/latest/docs/resources/catalog) for vApp Templates and upload the corresponding
TKGm (Tanzu Kubernetes Grid) OVA files to be able to create Kubernetes clusters.

```hcl
data "vcloud_storage_profile" "cse_storage_profile" {
  org        = vcloud_org.cse_org.name
  vdc        = vcloud_org_vdc.cse_vdc.name
  name       = "*"
  depends_on = [vcloud_org.cse_org, vcloud_org_vdc.cse_vdc]
}

resource "vcloud_catalog" "cat-cse" {
  org         = vcloud_org.cse_org.name
  name        = "cat-cse"
  description = "CSE catalog"

  storage_profile_id = data.vcloud_storage_profile.cse_storage_profile.id

  delete_force     = true
  delete_recursive = true
  depends_on       = [vcloud_org_vdc.cse_vdc]
}
```

Then you can upload TKGm OVAs to this catalog. These can be downloaded from **VMware Customer Connect**.
To upload them, use the [Catalog Item](/providers/vmware/vcloud/latest/docs/resources/catalog_item) resource with
`metadata_entry`.

~> Only TKGm OVAs are supported. CSE is **not compatible** yet with PhotonOS

In the example below, the downloaded OVA corresponds to **TKGm v1.4.0** and uses Kubernetes v1.21.2. 

```hcl
resource "vcloud_catalog_item" "tkgm_ova" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this resource

  org     = vcloud_org.cse_org.name
  catalog = vcloud_catalog.cat-cse.name

  name              = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  description       = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  ova_path          = "/Users/johndoe/Download/ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322.ova"
  upload_piece_size = 100

  metadata_entry {
    key         = "kind"
    value       = "TKGm" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "kubernetes"
    value       = "TKGm" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "kubernetes_version"
    value       = split("-", var.tkgm-ova-name)[3] # The version comes in the OVA name downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "name"
    value       = replace(var.tkgm-ova-name, ".ova", "") # The name as it was in the OVA downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "os"
    value       = split("-", var.tkgm-ova-name)[0] # The OS comes in the OVA name downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "revision"
    value       = "1" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }
}
```

Notice that all the metadata entries from the set of `metadata_entry` are required for CSE to fetch the OVA file:
* `kind`: Needs to be set to `TKGm` in all cases, as *Native* is not supported yet.
* `kubernetes`: Same as above.
* `kubernetes_version`: When the OVA is downloaded from VMware Customer Connect, the version appears as part of the file name.
* `name`: OVA full file name. VMware Customer Connect should provide already the downloaded OVA with a proper canonical name.
* `os`: When the OVA is downloaded from VMware Customer Connect, the OS appears as part of the file name.
* `revision`: Needs to be always `1`. This information is internally used by CSE.

Alternatively, you can upload the OVA file using `cse` CLI. This command line tool is explained in the next step.

### Step 6: CSE cli

This step can be done manually, by executing the `cse` CLI in a given shell, or can be automated within the Terraform HCL.

-> To see an example of how `cse` cli is automated using Terraform
[`null_resource`](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) from the
[null provider](https://registry.terraform.io/providers/hashicorp/null/3.1.1), take a look in the "[Examples](#examples)"
section below.

In any case, you need to have installed the [CSE command line interface](https://vmware.github.io/container-service-extension/cse3_0/INSTALLATION.html#getting_cse)
and then provide a YAML configuration file with the entities that were created by Terraform.

You can check the documentation for this configuration file [here](https://vmware.github.io/container-service-extension/cse3_1/CSE_CONFIG.html).
An example file is provided below, with all the information from the snippets shown in previous steps:

```yaml
mqtt:
  verify_ssl: false
 
vcloud:
  host: my-vcloud-dev.company.com
  log: true
  password: "*****"
  port: 443
  username: cse-service-account
  verify: false
 
service:
  enforce_authorization: false
  legacy_mode: false
  log_wire: false
  no_vc_communication_mode: true
  processors: 15
  telemetry:
    enable: true
 
broker:
  catalog: cat-cse
  ip_allocation_mode: pool
  network: cse_routed_net
  org: cse_org
  remote_template_cookbook_url: https://raw.githubusercontent.com/vmware/container-service-extension-templates/master/template_v2.yaml
  storage_profile: '*'
  vdc: cse_vdc
```

If you choose to execute it manually in a shell, launch the following command:

```shell
cse install -c config.yaml
```

CSE will use the configuration file from above to install some new custom entities and rights among other required settings.
You can also refer to the command line [documentation](https://vmware.github.io/container-service-extension) to upload OVA files
if you skipped the upload with Terraform from previous step.

If you choose to execute it using Terraform HCL, as in the "[Examples](#examples)" section below, you need to use the `null_resource` with a
`local-exec` provisioner and transform the file `config.yaml` into a template (in the snippet below, `config.yaml.template`) with all
required values as placeholders:

```yaml
# ... (omitted content for brevity)
broker:
  catalog: "${catalog}"
  ip_allocation_mode: pool
  network: "${network}"
  org: "${org}"
  remote_template_cookbook_url: https://raw.githubusercontent.com/vmware/container-service-extension-templates/master/template_v2.yaml
  storage_profile: "${storage_profile}"
  vdc: "${vdc}"
```

```hcl
resource "null_resource" "cse-install-script" {
  triggers = {
    always_run = timestamp() # Force to always trigger
  }

  provisioner "local-exec" {
    on_failure = continue # Ignores failures to allow re-creating the whole HCL after a destroy, as cse doesn't have an uninstall option.
    command = format("printf '%s' > config.yaml && chmod 0400 config.yaml && cse install -c config.yaml", templatefile("${path.module}/config.yaml.template", {
      vcloud_url         = replace(replace(var.vcloud-url, "/api", ""), "/http.*\\/\\//", "")
      vcloud_username    = vcloud_org_user.cse-service-account.name # Using CSE Service Account
      vcloud_password    = vcloud_org_user.cse-service-account.password
      catalog         = vcloud_catalog.cat-cse.name
      network         = vcloud_network_routed_v2.cse_routed.name
      org             = vcloud_org.cse_org.name
      vdc             = vcloud_org_vdc.cse_vdc.name
      storage_profile = data.vcloud_storage_profile.cse_sp.name
    }))
  }
}
```

When using the HCL option, take into account the following important aspects:
* The generated `config.yaml` needs to **not** to have read permissions for group and others (`chmod 0400`).
* `cse install` must run just once (in subsequent runs it should be `cse upgrade`), so a way to allowing Terraform to
  apply and destroy multiple times is to add `on_failure = continue` to the local-exec provisioner.
* As a consequence, if `config.yaml` is wrong or the `cse` command is not present, `on_failure = continue` will make
  Terraform continue on any failure. In this case, you'll see a failure in next steps, as `cse` installs several rights in Vcloud
  that are needed.

### Step 7: Rights and roles

You need to publish a new [Rights Bundle](/providers/vmware/vcloud/latest/docs/resources/rights_bundle) to your
Organization with the new rights that `cse install` command created in Vcloud.
The required new rights are listed in the example below. It creates a new bundle with a mix of the existent Default Rights Bundle rights and
the new ones.

```hcl
data "vcloud_rights_bundle" "default-rb" {
  name = "Default Rights Bundle"
}

resource "vcloud_rights_bundle" "cse-rb" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this resource

  name        = "CSE Rights Bundle"
  description = "Rights bundle to manage CSE"
  rights = setunion(data.vcloud_rights_bundle.default-rb.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify"
  ])
  publish_to_all_tenants = false
  tenants                = [vcloud_org.cse_org.name]
}
```

Once you have the new bundle created, you can now create a specific role for users that will be responsible for managing clusters.
Notice that the next example is assigning the new rights provided by the new published bundle:

```hcl
data "vcloud_role" "vapp_author" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this data source

  org  = vcloud_org.cse_org.name
  name = "vApp Author"
}

resource "vcloud_role" "cluster_author" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this resource

  org         = vcloud_org.cse_org.name
  name        = "Cluster Author"
  description = "Can read and create clusters"
  rights = setunion(data.vcloud_role.vapp_author.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: Configure NAT",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify",
    "Certificate Library: View" # Implicit role needed
  ])

  depends_on = [vcloud_rights_bundle.cse-rb]
}
```

You need also to publish the bundle that `cse install` command created, named **"cse:nativeCluster Entitlement"**. For that you can do the same
as above, create a clone. This is also recommended so doing `terraform destroy` will let the original rights bundle intact. 

```hcl
data "vcloud_rights_bundle" "cse-native-cluster-entl" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this data source

  name = "cse:nativeCluster Entitlement"

  depends_on = [null_resource.cse-install-script]
}

resource "vcloud_rights_bundle" "published-cse-rights-bundle" {
  provider = vcloud.cse-service-account # Using CSE Service Account for this resource

  name                   = "cse:nativeCluster Entitlement Published"
  description            = data.vcloud_rights_bundle.cse-native-cluster-entl.description
  rights                 = data.vcloud_rights_bundle.cse-native-cluster-entl.rights
  publish_to_all_tenants = false
  tenants                = [vcloud_org.cse_org.name]
}
```

### Final step

After applying all the above resources successfully, make sure you publish the **Container UI Plugin** to the desired tenants.
To do this, login in Vcloud as System administrator, click on "More" in the top bar, then "Customize Portal".
You will see a list of plugins, you need to publish **Container UI Plugin** to the target tenant.

Finally, **run the CSE server** in your infrastructure, by executing `cse run -c config.yaml`. Take into account that server
will start running indefinitely, so plan to execute this command in a dedicated place.

The `cse run` command should fetch all resources and OVAs and allow the tenant users to provision Kubernetes clusters in Vcloud web UI.
If they have the required rights from the role created in previous step, they should now be able to see the "Kubernetes Container Clusters"
option in the "More" option in the top bar.

## Examples

There are available examples in the [GitHub repository](https://github.com/vmware/terraform-provider-vcloud/tree/main/examples/container-service-extension-3.1.x).
