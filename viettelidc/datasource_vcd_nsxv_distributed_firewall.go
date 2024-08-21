package viettelidc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func computedMap(input map[string]*schema.Schema) map[string]*schema.Schema {
	var output = make(map[string]*schema.Schema)
	for k, v := range input {
		v.Required = false
		v.Computed = true
		v.StateFunc = nil
		v.ValidateFunc = nil
		output[k] = v
	}
	return output
}

func datasourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvDistributedFirewallRead,
		Schema: map[string]*schema.Schema{
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "When true, the retrieved NSX-V distributed firewall is enabled",
			},
			"rule": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Ordered list of distributed firewall rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Firewall Rule ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rule is enabled",
						},
						"logged": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rule traffic is logged",
						},
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action of the rule (allow, deny)",
						},
						"direction": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Direction of the rule (in, out, inout)",
						},
						"packet_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Packet type of the rule (any, ipv4, ipv6)",
						},
						"source": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "List of source traffic for this rule. An empty value means 'any'",
							Elem: &schema.Resource{
								Schema: computedMap(sourceDef().Schema),
							},
						},
						"application": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Application definitions for this rule. An empty value means 'any'",
							Elem: &schema.Resource{
								Schema: computedMap(applicationDef().Schema),
							},
						},
						"exclude_source": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "If true, the content of the source elements is reversed",
						},
						"destination": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "List of destination traffic for this rule. An empty value means 'any'",
							Elem: &schema.Resource{
								Schema: computedMap(destinationDef().Schema),
							},
						},
						"exclude_destination": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "If true, the content of the destination elements is reversed",
						},
						"applied_to": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "List of elements to which this rule applies",
							Elem: &schema.Resource{
								Schema: computedMap(appliedToDef().Schema),
							},
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxvDistributedFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxvDistributedFirewallRead(ctx, d, meta, "datasource")
}
