package vcloud

import (
	"context"
	_ "embed"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdCseKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdCseKubernetesRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"cluster_id", "name"},
				Description:  "The unique ID of the Kubernetes cluster to read",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"cluster_id", "name"},
				RequiredWith: []string{"cse_version", "org_id"},
				Description:  "The name of the Kubernetes cluster to read. If there is more than one Kubernetes cluster with the same name, searching by name will fail",
			},
			"org_id": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"cse_version", "name"},
				Description:  "The ID of organization that owns the Kubernetes cluster, only required if 'name' is set",
			},
			"cse_version": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"name", "org_id"},
				Description:  "The CSE version used by the cluster, only required if 'name' is set",
			},
			"runtime": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Kubernetes runtime used by the cluster",
			},
			"kubernetes_template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the vApp Template that corresponds to a Kubernetes template OVA",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the VDC that hosts the Kubernetes cluster",
			},
			"network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the network that the Kubernetes cluster uses",
			},
			"owner": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user that created the cluster",
			},
			"ssh_public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SSH public key used to login into the cluster nodes",
			},
			"control_plane": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Defines the control plane for the cluster",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"machine_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of nodes that the control plane has",
						},
						"disk_size_gi": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Disk size, in Gibibytes (Gi), of the control plane nodes",
						},
						"sizing_policy_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VM Sizing policy of the control plane nodes",
						},
						"placement_policy_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VM Placement policy of the control plane nodes",
						},
						"storage_profile_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Storage profile of the control plane nodes",
						},
						"ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP of the control plane",
						},
					},
				},
			},
			"worker_pool": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Defines a node pool for the cluster",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of this node pool",
						},
						"machine_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The number of nodes that this node pool has",
						},
						"disk_size_gi": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Disk size, in Gibibytes (Gi), of the control plane nodes",
						},
						"sizing_policy_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VM Sizing policy of the control plane nodes",
						},
						"placement_policy_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VM Placement policy of the control plane nodes",
						},
						"vgpu_policy_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "vGPU policy of the control plane nodes",
						},
						"storage_profile_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Storage profile of the control plane nodes",
						},
						"autoscaler_max_replicas": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Maximum replicas of the autoscaling capabilities of this worker pool",
						},
						"autoscaler_min_replicas": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum replicas of the autoscaling capabilities of this worker pool",
						},
					},
				},
			},
			"default_storage_class": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The default storage class of the cluster, if any",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_profile_id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "ID of the storage profile used by the storage class",
						},
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of the storage class",
						},
						"reclaim_policy": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "'delete' deletes the volume when the PersistentVolumeClaim is deleted. 'retain' does not, and the volume can be manually reclaimed",
						},
						"filesystem": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Filesystem of the storage class, can be either 'ext4' or 'xfs'",
						},
					},
				},
			},
			"pods_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CIDR that the Kubernetes pods use",
			},
			"services_cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "CIDR that the Kubernetes services use",
			},
			"virtual_ip_subnet": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual IP subnet of the cluster",
			},
			"auto_repair_on_errors": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "If errors occur before the Kubernetes cluster becomes available, and this argument is 'true', CSE Server will automatically attempt to repair the cluster",
			},
			"node_health_check": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "After the Kubernetes cluster becomes available, nodes that become unhealthy will be remediated according to unhealthy node conditions and remediation rules",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of Kubernetes installed in this cluster",
			},
			"tkg_product_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of TKG installed in this cluster",
			},
			"capvcd_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of CAPVCD used by this cluster",
			},
			"cluster_resource_set_bindings": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The cluster resource set bindings of this cluster",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpi_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the Cloud Provider Interface used by this cluster",
			},
			"csi_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the Container Storage Interface used by this cluster",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The state of the cluster, can be 'provisioning', 'provisioned', 'deleting' or 'error'. Useful to check whether the Kubernetes cluster is in a stable status",
			},
			"kubeconfig": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contents of the kubeconfig of the Kubernetes cluster, only available when 'state=provisioned'",
				Sensitive:   true,
			},
			"supported_upgrades": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of vApp Template names that could be used to upgrade the existing cluster",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"events": {
				Type:        schema.TypeList, // Order matters here, as they're ordered by date
				Computed:    true,
				Description: "A set of events that happened during the Kubernetes cluster lifecycle",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of the event",
						},
						"resource_id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "ID of the resource that caused the event",
						},
						"type": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Type of the event, either 'event' or 'error'",
						},
						"occurred_at": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "When the event happened",
						},
						"details": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Details of the event",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdCseKubernetesRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	vcdClient := meta.(*VCDClient)
	var cluster *govcd.CseKubernetesCluster
	var err error
	if id, ok := d.GetOk("cluster_id"); ok {
		cluster, err = vcdClient.CseGetKubernetesClusterById(id.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else if name, ok := d.GetOk("name"); ok {
		cseVersion, err := semver.NewVersion(d.Get("cse_version").(string))
		if err != nil {
			return diag.Errorf("could not parse cse_version='%s': %s", cseVersion, err)
		}

		orgId := d.Get("org_id").(string)
		org, err := vcdClient.GetOrgById(orgId)
		if err != nil {
			return diag.Errorf("could not find an Organization with ID '%s': %s", orgId, err)
		}

		clusters, err := org.CseGetKubernetesClustersByName(*cseVersion, name.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		if len(clusters) != 1 {
			return diag.Errorf("expected one Kubernetes cluster with name '%s', got %d. Try to use 'cluster_id' instead of 'name'", name, len(clusters))
		}
		cluster = clusters[0]
	}

	// These fields are specific to the data source
	dSet(d, "org_id", cluster.OrganizationId)
	dSet(d, "cluster_id", cluster.ID)

	warns, err := saveClusterDataToState(d, vcdClient, cluster, "datasource")
	if err != nil {
		return diag.Errorf("could not save Kubernetes cluster data into Terraform state: %s", err)
	}
	for _, warning := range warns {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  warning.Error(),
		})
	}

	if len(diags) > 0 {
		return diags
	}
	return nil
}
