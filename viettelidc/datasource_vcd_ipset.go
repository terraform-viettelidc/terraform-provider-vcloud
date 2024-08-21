package viettelidc

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdIpSet() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdIpSetRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP set name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP set description",
			},
			"is_inheritance_allowed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Allows visibility in underlying scopes (Default is true)",
			},
			"ip_addresses": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP address, CIDR, IP range objects",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
