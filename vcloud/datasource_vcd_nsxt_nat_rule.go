package vcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtNatRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtNatRuleRead,

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
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NAT rule",
			},
			"rule_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Rule type - one of 'DNAT', 'NO_DNAT', 'SNAT', 'NO_SNAT'",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NAT rule",
			},
			"external_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address or CIDR of external network",
			},
			"internal_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address or CIDR of the virtual machines for which you are configuring NAT",
			},
			"app_port_profile_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application Port Profile ID applied for this rule",
			},
			"dnat_external_port": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "For DNAT only. Port into which the DNAT rule is translating for the packets inbound to the virtual machines.",
			},
			"snat_destination_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "For SNAT only. Limits SNAT rule by destination IP address or range in CIDR format.",
			},
			"logging": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable logging when this rule is applied",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enables or disables this rule",
			},
			"firewall_match": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "VCD 10.2.2+ Determines how the firewall matches the address during NATing if firewall stage is not skipped. One of 'MATCH_INTERNAL_ADDRESS', 'MATCH_EXTERNAL_ADDRESS', 'BYPASS'",
			},
			"priority": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "VCD 10.2.2+ If an address has multiple NAT rules, the rule with the highest priority is applied. A lower value means a higher precedence for this rule.",
			},
		},
	}
}

func datasourceVcdNsxtNatRuleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	natRuleName := d.Get("name").(string)

	existingRule, err := nsxtEdge.GetNatRuleByName(natRuleName)
	if err != nil {
		return diag.Errorf("unable to find NSX-T NAT rule with Name '%s': %s", natRuleName, err)
	}

	err = setNsxtNatRuleData(existingRule.NsxtNatRule, d, vcdClient)
	if err != nil {
		return diag.Errorf("error storing NSX-T NAT rule in statefile: %s", err)
	}
	d.SetId(existingRule.NsxtNatRule.ID)

	return nil
}
