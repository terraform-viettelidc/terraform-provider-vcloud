package viettelidc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVdcGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVdcGroupRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "Name of VDC group",
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "id"},
				Description:  "VDC group ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "VDC group description",
			},
			"dfw_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Distributed firewall status",
			},
			"default_policy_status": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Default Policy Status",
			},
			"error_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "More detailed error message when VDC group has error status",
			},
			"local_egress": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Status whether local egress is enabled for a universal router belonging to a universal VDC group",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of used network pool",
			},
			"network_pool_universal_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The network provider’s universal id that is backing the universal network pool",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Defines the networking provider backing the VDC Group",
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "The status that the group can be in (e.g. 'SAVING', 'SAVED', 'CONFIGURING'," +
					" 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND'," +
					" 'UNCONFIGURED')",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Defines the group as LOCAL or UNIVERSAL",
			},
			"universal_networking_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "True means that a VDC group router has been created",
			},
			"participating_org_vdcs": {
				Computed:    true,
				Type:        schema.TypeSet,
				Description: "The list of organization VDCs that are participating in this group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fault_domain_tag": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Represents the fault domain of a given organization VDC",
						},
						"network_provider_scope": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Specifies the network provider scope of the VDC",
						},
						"is_remote_org": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Specifies whether the VDC is local to this VCD site",
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "The status that the VDC can be in e.g. 'SAVING', 'SAVED', 'CONFIGURING'," +
								" 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND'," +
								" 'UNCONFIGURED')",
						},
						"org_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization VDC belongs",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization VDC belongs",
						},
						"site_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Site VDC belongs",
						},
						"site_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Site VDC belongs",
						},
						"vdc_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VDC name",
						},
						"vdc_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VDC ID",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdVdcGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	// get by ID when it's available
	var vdcGroup *govcd.VdcGroup
	name := d.Get("name").(string)
	if name != "" {
		vdcGroup, err = adminOrg.GetVdcGroupByName(name)
	} else if d.Get("id").(string) != "" {
		vdcGroup, err = adminOrg.GetVdcGroupById(d.Get("id").(string))
	} else {
		return diag.Errorf("Id or Name value is missing %s", err)
	}
	if err != nil {
		return diag.Errorf("[VDC group read] : %s", err)
	}

	d.SetId(vdcGroup.VdcGroup.Id)
	err = setVdcGroupConfigurationData(vdcGroup.VdcGroup, d, nil)
	if err != nil {
		return diag.Errorf("[VDC group read] : %s", err)
	}

	return nil
}
