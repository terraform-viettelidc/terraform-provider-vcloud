package viettelidc

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdIpSpace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdIpSpaceRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Org ID for 'SHARED' IP spaces",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of IP space",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of IP space",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of IP space",
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of up internal scope IPs in CIDR format",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_range_quota": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP range quota",
			},
			"ip_range": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of IP Range",
						},
						"start_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Start address of the IP range",
						},
						"end_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "End address of the IP range",
						},
					},
				},
			},
			"ip_prefix": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP prefixes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"prefix": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "IP Prefix",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "ID of IP Prefix",
									},
									"first_ip": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "First IP",
									},
									"prefix_length": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Prefix length",
									},
									"prefix_count": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Prefix count",
									},
								},
							},
						},
						"default_quota": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Floating IP quota",
						},
					},
				},
			},
			"external_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External scope in CIDR format",
			},
			"route_advertisement_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag exposing if route advertisement is enabled",
			},
			"default_firewall_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag exposing whether default firewall rule creation should be enabled (VCD 10.5.0+)",
			},
			"default_no_snat_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag whether NO SNAT rule creation should be enabled (VCD 10.5.0+)",
			},
			"default_snat_rule_creation_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag whether SNAT rule creation should be enabled (VCD 10.5.0+)",
			},
		},
	}
}

func datasourceVcdIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space datasource read initiated")

	orgId := d.Get("org_id").(string)
	ipSpaceName := d.Get("name").(string)

	var ipSpace *govcd.IpSpace
	var err error

	if orgId != "" { // in case org_id is provided (PRIVATE IP Space)
		ipSpace, err = vcdClient.GetIpSpaceByNameAndOrgId(ipSpaceName, orgId)
		if err != nil {
			return diag.Errorf("error retrieving IP Space '%s' in Org ID '%s': %s", ipSpaceName, orgId, err)
		}
	} else {
		ipSpace, err = vcdClient.GetIpSpaceByName(ipSpaceName)
		if err != nil {
			return diag.Errorf("error retrieving IP Space '%s': %s", ipSpaceName, err)
		}
	}

	err = setIpSpaceData(d, ipSpace.IpSpace)
	if err != nil {
		return diag.Errorf("error storing IP Space state: %s", err)
	}

	d.SetId(ipSpace.IpSpace.ID)

	return nil
}
