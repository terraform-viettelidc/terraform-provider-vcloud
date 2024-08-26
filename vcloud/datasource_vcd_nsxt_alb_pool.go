package vcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbPoolRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID in which ALB Pool should be created",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of ALB Pool",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if NSX-T ALB Pool is enabled or disabled",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of ALB Pool",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Algorithm for choosing pool members (default LEAST_CONNECTIONS)",
			},

			"default_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Default Port defines destination server port used by the traffic sent to the member (default 80)",
			},
			"graceful_timeout_period": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum time in minutes to gracefully disable pool member (default 1)",
			},
			"member": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "ALB Pool Members",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Shows is the member is enabled or not",
						},
						"ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP Address of pool member",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Service port",
						},
						"ratio": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Load ratio",
						},
						"marked_down_by": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Marked down by provides a set of health monitors that marked the service down",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"health_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Health status",
						},
						"detailed_health_message": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Detailed health message",
						},
					},
				},
			},
			"member_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of Firewall Group to use for Pool Membership (VCD 10.4.1+)",
			},
			"health_monitor": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of health monitor",
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_defined": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"persistence_profile": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "System generated name of persistence profile",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of persistence strategy",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Value of attribute based on persistence type",
						},
					},
				},
			},
			"ca_certificate_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of root certificate IDs to use when validating certificates presented by pool members",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cn_check_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean flag if common name check of the certificate should be enabled",
			},
			"domain_names": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of domain names which will be used to verify common names",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"passive_monitoring_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Monitors if the traffic is accepted by node (default true)",
			},
			"ssl_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enables SSL - Necessary when CA certificates are used",
			},
			// Read only information
			"associated_virtual_service_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "IDs of associated virtual services",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"associated_virtual_services": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Names of associated virtual services",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of members in the pool",
			},
			"up_member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of members in the pool serving the traffic",
			},
			"enabled_member_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of enabled members in the pool",
			},
			"health_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Health message",
			},
		},
	}
}

func datasourceVcdAlbPoolRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	nsxtEdge, err := org.GetNsxtEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T nsxtEdge gateway with ID '%s': %s", d.Id(), err)
	}

	albPool, err := vcdClient.GetAlbPoolByName(nsxtEdge.EdgeGateway.ID, d.Get("name").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T ALB Pool '%s': %s", d.Get("name").(string), err)
	}

	err = setNsxtAlbPoolData(d, albPool.NsxtAlbPool)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Pool data: %s", err)
	}
	d.SetId(albPool.NsxtAlbPool.ID)

	return nil
}
