package viettelidc

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdLbServerPool() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLbServerPoolRead,
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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the LB Server Pool is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server Pool name for lookup",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Server pool description",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Balancing method for the service",
			},
			"algorithm_parameters": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Additional options for load balancing algorithm for http-header or url algorithms",
			},
			"monitor_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Load Balancer Service Monitor ID",
			},
			"enable_transparency": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Makes client IP addresses visible to the backend servers",
			},
			"member": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Pool member id (formatted as member-xx, where xx is a number)",
						},
						"condition": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Defines member state. One of enabled, drain, disabled.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of pool member",
						},
						"ip_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address of member in server pool",
						},
						"port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port at which the member is to receive traffic from the load balancer",
						},
						"monitor_port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port at which the member is to receive health monitor requests. Can be the same as port",
						},
						"weight": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Proportion of traffic this member is to handle. Must be an integer in the range 1-256",
						},
						"min_connections": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Minimum number of concurrent connections a member must always accept",
						},
						"max_connections": {
							Type:     schema.TypeInt,
							Computed: true,
							Description: "The maximum number of concurrent connections the member can handle. If exceeded " +
								"requests are queued and the load balancer waits for a connection to be released",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdLbServerPoolRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.GetLbServerPoolByName(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("unable to find load balancer server pool with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBPool.ID)
	return setLBPoolData(d, readLBPool)
}
