package vcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbVirtualService() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbVirtualServiceRead,

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
				Description: "Edge gateway ID in which ALB Virtual Service is",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of ALB Virtual Service",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of ALB Virtual Service",
			},
			"pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Pool ID to use for this Virtual Service",
			},
			"service_engine_group_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service Engine Group ID",
			},
			"ca_certificate_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of certificate in library if used",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Virtual Service is enabled or disabled (default true)",
			},
			"virtual_ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual IP address (VIP) for Virtual Service",
			},
			"ipv6_virtual_ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IPv6 Virtual IP address (VIP) for Virtual Service (VCD 10.4.0+)",
			},
			"application_profile_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HTTP, HTTPS, L4, L4_TLS",
			},
			"service_port": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Starting port in the range",
						},
						"end_port": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Last port in the range",
						},
						"ssl_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Starting port in the range",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "One of 'TCP_PROXY', 'TCP_FAST_PATH', 'UDP_FAST_PATH'",
						},
					},
				},
			},
			"is_transparent_mode_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Preserves Client IP on a Virtual Service when enabled (VCD 10.4.1+)",
			},
		},
	}
}

func datasourceVcdAlbVirtualServiceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	nsxtEdge, err := org.GetNsxtEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	albVirtualService, err := vcdClient.GetAlbVirtualServiceByName(nsxtEdge.EdgeGateway.ID, d.Get("name").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T ALB Virtual Service '%s': %s", d.Get("name").(string), err)
	}

	err = setNsxtAlbVirtualServiceData(d, albVirtualService.NsxtAlbVirtualService)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Virtual Service data: %s", err)
	}
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return nil
}
