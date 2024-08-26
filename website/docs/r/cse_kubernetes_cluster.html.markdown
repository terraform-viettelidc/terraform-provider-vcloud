---
layout: "vcloud"
page_title: "Viettel IDC Cloud: vcloud_cse_kubernetes_cluster"
sidebar_current: "docs-vcloud-resource-cse-kubernetes-cluster"
description: |-
  Provides a resource to manage Kubernetes clusters in Viettel IDC Cloud with Container Service Extension installed and running.
---

# vcloud\_cse\_kubernetes\_cluster

Provides a resource to manage Kubernetes clusters in Viettel IDC Cloud with Container Service Extension (CSE) installed and running.

Supported in provider *v3.12+*

Supports the following **Container Service Extension** versions:

* [4.1.0](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1/rn/vmware-cloud-director-container-service-extension-41-release-notes/index.html)
* [4.1.1 / 4.1.1a](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1.1/rn/vmware-cloud-director-container-service-extension-411-release-notes/index.html)
* [4.2.0](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2/rn/vmware-cloud-director-container-service-extension-42-release-notes/index.html)
* [4.2.1](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2.1/rn/vmware-cloud-director-container-service-extension-421-release-notes/index.html)

-> To install CSE in Viettel IDC Cloud, please follow [this guide](/providers/terraform-viettelidc/vcloud/latest/docs/guides/container_service_extension_4_x_install)

## Example Usage

```hcl
data "vcloud_catalog" "tkg_catalog" {
  org  = "solutions_org" # The catalog is shared with 'tenant_org', so it is visible for tenants
  name = "tkgm_catalog"
}

# Fetch a valid Kubernetes template OVA. If it's not valid, cluster creation will fail.
data "vcloud_catalog_vapp_template" "tkg_ova" {
  org        = data.vcloud_catalog.tkg_catalog.org
  catalog_id = data.vcloud_catalog.tkg_catalog.id
  name       = "ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc"
}

data "vcloud_org_vdc" "vdc" {
  org  = "tenant_org"
  name = "tenant_vdc"
}

data "vcloud_nsxt_edgegateway" "egw" {
  org      = data.vcloud_org_vdc.vdc.org
  owner_id = data.vcloud_org_vdc.vdc.id
  name     = "tenant_edgegateway"
}

data "vcloud_network_routed_v2" "routed" {
  org             = data.vcloud_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcloud_nsxt_edgegateway.egw.id
  name            = "tenant_net_routed"
}

# Fetch a valid Sizing policy created during CSE installation.
# Refer to the CSE installation guide for more information.
data "vcloud_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

data "vcloud_storage_profile" "sp" {
  org  = data.vcloud_org_vdc.vdc.org
  vdc  = data.vcloud_org_vdc.vdc.name
  name = "*"
}

# The token file is required, and it should be safely stored
resource "vcloud_api_token" "token" {
  name             = "myClusterToken"
  file_name        = "/home/Bob/safely_stored_token.json"
  allow_token_file = true
}

resource "vcloud_cse_kubernetes_cluster" "my_cluster" {
  cse_version            = "4.2.1"
  runtime                = "tkg"
  name                   = "test2"
  kubernetes_template_id = data.vcloud_catalog_vapp_template.tkg_ova.id
  org                    = data.vcloud_org_vdc.vdc.org
  vdc_id                 = data.vcloud_org_vdc.vdc.id
  network_id             = data.vcloud_network_routed_v2.routed.id
  api_token_file         = vcloud_api_token.token.file_name

  control_plane {
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcloud_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcloud_storage_profile.sp.id
  }

  worker_pool {
    name               = "node-pool-1"
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcloud_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcloud_storage_profile.sp.id
  }

  default_storage_class {
    name               = "sc-1"
    storage_profile_id = data.vcloud_storage_profile.sp.id
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  auto_repair_on_errors = true
  node_health_check     = true

  operations_timeout_minutes = 0
}

output "kubeconfig" {
  value     = vcloud_cse_kubernetes_cluster.my_cluster.kubeconfig
  sensitive = true
}
```

## Argument Reference

The following arguments are supported:

* `cse_version` - (Required) Specifies the CSE version to use. Accepted versions: `4.1.0`, `4.1.1` (also for *4.1.1a*), `4.2.0` and `4.2.1`
* `runtime` - (Optional) Specifies the Kubernetes runtime to use. Defaults to `tkg` (Tanzu Kubernetes Grid)
* `name` - (Required) The name of the Kubernetes cluster. It must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `kubernetes_template_id` - (Required) The ID of the vApp Template that corresponds to a Kubernetes template OVA
* `org` - (Optional) The name of organization that will host the Kubernetes cluster, optional if defined in the provider configuration
* `vdc_id` - (Required) The ID of the VDC that hosts the Kubernetes cluster
* `network_id` - (Required) The ID of the network that the Kubernetes cluster will use
* `owner` - (Optional) The user that creates the cluster and owns the API token specified in `api_token`.
  It must have the `Kubernetes Cluster Author` role that was created during CSE installation.
  If not specified, it assumes it's the user from the provider configuration
* `api_token_file` - (Required) Must be a file generated by [`vcloud_api_token` resource](/providers/terraform-viettelidc/vcloud/latest/docs/resources/api_token),
  or a file that follows the same formatting, that stores the API token used to create and manage the cluster,
  owned by the user specified in `owner`. Be careful about this file, as it contains sensitive information
* `ssh_public_key` - (Optional) The SSH public key used to log in into the cluster nodes
* `control_plane` - (Required) See [**Control Plane**](#control-plane)
* `worker_pool` - (Required) See [**Worker Pools**](#worker-pools)
* `default_storage_class` - (Optional) See [**Default Storage Class**](#default-storage-class)
* `pods_cidr` - (Optional) A CIDR block for the pods to use. Defaults to `100.96.0.0/11`
* `services_cidr` - (Optional) A CIDR block for the services to use. Defaults to `100.64.0.0/13`
* `virtual_ip_subnet` - (Optional) A virtual IP subnet for the cluster
* `auto_repair_on_errors` - (Optional) If errors occur before the Kubernetes cluster becomes available, and this argument is `true`,
  CSE Server will automatically attempt to repair the cluster. Defaults to `false`.
  Since CSE 4.1.1, when the cluster is available/provisioned, this flag is set automatically to false.
* `node_health_check` - (Optional) After the Kubernetes cluster becomes available, nodes that become unhealthy will be
  remediated according to unhealthy node conditions and remediation rules. Defaults to `false`
* `operations_timeout_minutes` - (Optional) The time, in minutes, to wait for the cluster operations to be successfully completed.
  For example, during cluster creation, it should be in `provisioned` state before the timeout is reached, otherwise the
  operation will return an error. For cluster deletion, this timeout specifies the time to wait until the cluster is completely deleted.
  Setting this argument to `0` means to wait indefinitely (not recommended as it could hang Terraform if the cluster can't be created
  due to a configuration error if `auto_repair_on_errors=true`). Defaults to `60`

### Control Plane

The `control_plane` block is **required** and unique per resource, meaning that there must be **exactly one** of these
in every resource.

This block asks for the following arguments:

* `machine_count` - (Optional) The number of nodes that the control plane has. Must be an odd number and higher than `0`. Defaults to `3`
* `disk_size_gi` - (Optional) Disk size, in **Gibibytes (Gi)**, for the control plane VMs. Must be at least `20`. Defaults to `20`
* `sizing_policy_id` - (Optional) VM Sizing policy for the control plane VMs. Must be one of the ones made available during CSE installation
* `placement_policy_id` - (Optional) VM Placement policy for the control plane VMs
* `storage_profile_id` - (Optional) Storage profile for the control plane VMs
* `ip` - (Optional) IP for the control plane. It will be automatically assigned during cluster creation if left empty

### Worker Pools

The `worker_pool` block is **required**, and every cluster should have **at least one** of them.

Each block asks for the following arguments:

* `name` - (Required) The name of the worker pool. It must be unique per cluster, and must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `machine_count` - (Optional) The number of VMs that the worker pool has. Must be higher than `0`. Defaults to `1`
* `disk_size_gi` - (Optional) Disk size, in **Gibibytes (Gi)**, for the worker pool VMs. Must be at least `20`. Defaults to `20`
* `sizing_policy_id` - (Optional) VM Sizing policy for the control plane VMs. Must be one of the ones made available during CSE installation
* `placement_policy_id` - (Optional) VM Placement policy for the worker pool VMs. If this one is set, `vgpu_policy_id` must be empty
* `vgpu_policy_id` - (Optional) vGPU policy for the worker pool VMs. If this one is set, `placement_policy_id` must be empty
* `storage_profile_id` - (Optional) Storage profile for the worker pool VMs

### Default Storage Class

The `default_storage_class` block is **optional**, and every cluster should have **at most one** of them.

If defined, the block asks for the following arguments:

* `name` - (Required) The name of the default storage class. It must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `storage_profile_id` - (Required) Storage profile for the default storage class
* `reclaim_policy` - (Required) A value of `delete` deletes the volume when the PersistentVolumeClaim is deleted. `retain` does not,
  and the volume can be manually reclaimed
* `filesystem` - (Required) Filesystem of the storage class, can be either `ext4` or `xfs`

## Attribute Reference

The following attributes are available for consumption as read-only attributes after a successful cluster creation:

* `kubernetes_version` - The version of Kubernetes installed in this cluster
* `tkg_product_version` - The version of TKG installed in this cluster
* `capvcloud_version` - The version of CAPVcloud used by this cluster
* `cluster_resource_set_bindings` - The cluster resource set bindings of this cluster
* `cpi_version` - The version of the Cloud Provider Interface used by this cluster
* `csi_version` - The version of the Container Storage Interface used by this cluster
* `state` - The Kubernetes cluster status, can be `provisioning` when it is being created, `provisioned` when it was successfully
  created and ready to use, or `error` when an error occurred. `provisioning` can only be obtained when a timeout happens during
  cluster creation. `error` can only be obtained either with a timeout or when `auto_repair_on_errors=false`.
* `kubeconfig` - The ready-to-use Kubeconfig file **contents** as a raw string. Only available when `state=provisioned`
* `supported_upgrades` - A set of vApp Template names that can be fetched with a
  [`vcloud_catalog_vapp_template` data source](/providers/terraform-viettelidc/vcloud/latest/docs/data-sources/catalog_vapp_template) to upgrade the cluster.
* `events` - A set of events that happened during the Kubernetes cluster lifecycle. They're ordered from most recent to least. Each event has:
  * `name` - Name of the event
  * `resource_id` - ID of the resource that caused the event
  * `type` - Type of the event, either `event` or `error`
  * `details` - Details of the event
  * `occurred_at` - When the event happened

## Updating

Only the following arguments can be updated:

* `kubernetes_template_id`: The cluster must allow upgrading to the new TKG version. You can check `supported_upgrades` attribute to know
  the available OVAs.
* `machine_count` of the `control_plane`: Supports scaling up and down. Nothing else can be updated.
* `machine_count` of any `worker_pool`: Supports scaling up and down. Use caution when resizing down to 0 nodes.
  The cluster must always have at least 1 running node, or else the cluster will enter an unrecoverable error state.
* `auto_repair_on_errors`: Can only be updated in CSE 4.1.0, and it is recommended to set it to `false` when the cluster is created.
  In versions higher than 4.1.0, this is automatically done by the CSE Server, so this flag cannot be updated.
* `node_health_check`: Can be turned on/off.
* `operations_timeout_minutes`: Does not require modifying the existing cluster

You can also add more `worker_pool` blocks to add more Worker Pools to the cluster. **You can't delete Worker Pools**, but they can
be scaled down to zero.

Updating any other argument will delete the existing cluster and create a new one, when the Terraform plan is applied.

Upgrading CSE version with `cse_version` is not supported as this operation would require human intervention,
as stated [in the official documentation](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1/VMware-Cloud-Director-Container-Service-Extension-Using-Tenant-4.1/GUID-092C40B4-D0BA-4B90-813F-D36929F2F395.html).

## Accessing the Kubernetes cluster

To retrieve the Kubeconfig of a created cluster, you may set it as an output:

```hcl
output "kubeconfig" {
  value     = vcloud_cse_kubernetes_cluster.my_cluster.kubeconfig
  sensitive = true
}
```

Then, creating a file turns out to be trivial:

```shell
terraform output -raw kubeconfig > $HOME/kubeconfig
```

The Kubeconfig can now be used with `kubectl` and the Kubernetes cluster can be used.

## Importing

An existing Kubernetes cluster can be [imported][docs-import] into this resource via supplying the **Cluster ID** for it.
The ID can be easily obtained in Vcloud UI, in the CSE Kubernetes Container Clusters plugin.

An example is below. During import, none of the mentioned arguments are required, but they will be in subsequent Terraform commands
such as `terraform plan`. Each comment in the code gives some context about how to obtain them to have a completely manageable cluster:

```hcl
# This is just a snippet of code that will host the imported cluster that already exists in Vcloud.
# This must NOT be created with Terraform beforehand, it is just a shell that will receive the information
# None of the arguments are required during the Import phase, but they will be asked when operating it afterwards
resource "vcloud_cse_kubernetes_cluster" "imported_cluster" {
  name                   = "test2"                                   # The name of the existing cluster
  cse_version            = "4.2.1"                                   # The CSE version installed in your Vcloud
  kubernetes_template_id = data.vcloud_catalog_vapp_template.tkg_ova.id # See below data sources
  vdc_id                 = data.vcloud_org_vdc.vdc.id                   # See below data sources
  network_id             = data.vcloud_network_routed_v2.routed.id      # See below data sources
  node_health_check      = true                                      # Whether the existing cluster has Machine Health Check enabled or not, this can be checked in UI

  control_plane {
    machine_count      = 5                                      # This is optional, but not setting it to the current value will make subsequent plans to try to scale our existing cluster to the default one
    sizing_policy_id   = data.vcloud_vm_sizing_policy.tkg_small.id # See below data sources
    storage_profile_id = data.vcloud_storage_profile.sp.id         # See below data sources
  }

  worker_pool {
    name               = "node-pool-1"                          # The name of the existing worker pool of the existing cluster. Retrievable from UI
    machine_count      = 40                                     # This is optional, but not setting it to the current value will make subsequent plans to try to scale our existing cluster to the default one
    sizing_policy_id   = data.vcloud_vm_sizing_policy.tkg_small.id # See below data sources
    storage_profile_id = data.vcloud_storage_profile.sp.id         # See below data sources
  }

  # While optional, we cannot change the Default Storage Class after an import, so we need
  # to set the information of the existing cluster to avoid re-creation.
  # The information can be retrieved from UI
  default_storage_class {
    filesystem         = "ext4"
    name               = "sc-1"
    reclaim_policy     = "delete"
    storage_profile_id = data.vcloud_storage_profile.sp.id # See below data sources
  }
}

# The below data sources are needed to retrieve the required IDs. They are not needed
# during the Import phase, but they will be asked when operating it afterwards

# The VDC and Organization where the existing cluster is located
data "vcloud_org_vdc" "vdc" {
  org  = "tenant_org"
  name = "tenant_vdc"
}

# The OVA that the existing cluster is using. You can obtain the OVA by inspecting
# the existing cluster TKG/Kubernetes version.
data "vcloud_catalog_vapp_template" "tkg_ova" {
  org        = data.vcloud_catalog.tkg_catalog.org
  catalog_id = data.vcloud_catalog.tkg_catalog.id
  name       = "ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc"
}

# The network that the existing cluster is using
data "vcloud_network_routed_v2" "routed" {
  org             = data.vcloud_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcloud_nsxt_edgegateway.egw.id
  name            = "tenant_net_routed"
}

# The VM Sizing Policy of the existing cluster nodes
data "vcloud_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

# The Storage Profile that the existing cluster uses
data "vcloud_storage_profile" "sp" {
  org  = data.vcloud_org_vdc.vdc.org
  vdc  = data.vcloud_org_vdc.vdc.name
  name = "*"
}

data "vcloud_catalog" "tkg_catalog" {
  org  = "solutions_org" # The Organization that shares the TKGm OVAs with the tenants
  name = "tkgm_catalog"  # The Catalog name
}

data "vcloud_nsxt_edgegateway" "egw" {
  org      = data.vcloud_org_vdc.vdc.org
  owner_id = data.vcloud_org_vdc.vdc.id
  name     = "tenant_edgegateway"
}
```

```sh
terraform import vcloud_cse_kubernetes_cluster.imported_cluster urn:vcloud:entity:vmware:capvcloudCluster:1d24af33-6e5a-4d47-a6ea-06d76f3ee5c9
```

-> The ID is required as it is the only way to unequivocally identify a Kubernetes cluster inside Vcloud. To obtain the ID
you can check the Kubernetes Container Clusters UI plugin, where all the available clusters are listed.

After that, you can expand the configuration file and either update or delete the Kubernetes cluster. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Kubernetes cluster stored properties.

### Importing with Import blocks (Terraform v1.5+)

~> Terraform warns that this procedure is considered **experimental**. Read more [here](/providers/terraform-viettelidc/vcloud/latest/docs/guides/importing_resources)

Given a Cluster ID, like `urn:vcloud:entity:vmware:capvcloudCluster:f2d88194-3745-47ef-a6e1-5ee0bbce38f6`, you can write 
the following HCL block in your Terraform configuration:

```hcl
import {
  to = vcloud_cse_kubernetes_cluster.imported_cluster
  id = "urn:vcloud:entity:vmware:capvcloudCluster:f2d88194-3745-47ef-a6e1-5ee0bbce38f6"
}
```

Instead of using the suggested snippet in the section above, executing the command
`terraform plan -generate-config-out=generated_resources.tf` will generate a similar code, automatically.

Once the code is validated, running `terraform apply` will perform the import operation and save the Kubernetes cluster
into the Terraform state. The Kubernetes cluster can now be operated with Terraform.

[docs-import]:https://www.terraform.io/docs/import/
