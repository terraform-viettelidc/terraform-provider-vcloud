package viettelidc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdVmAffinityRule() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVmAffinityRuleRead,
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
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "rule_id"},
				Description:  "VM affinity rule name. Used to retrieve a rule only when the name is unique",
			},
			"rule_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VM affinity rule ID. It's the preferred way of identifying a rule",
			},
			"polarity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of 'Affinity', 'Anti-Affinity'",
			},
			"required": {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "True if this affinity rule is required. When a rule is mandatory, " +
					"a host failover will not power on the VM if doing so would violate the rule",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this affinity rule is enabled",
			},
			"vm_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of VM IDs assigned to this rule",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// datasourceVcdVmAffinityRuleRead reads a data source VM affinity rule
func datasourceVcdVmAffinityRuleRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVmAffinityRuleRead(d, meta, "datasource")
}
