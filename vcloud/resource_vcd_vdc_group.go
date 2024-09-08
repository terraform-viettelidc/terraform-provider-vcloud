package vcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var participatingOrgVdcsResource = &schema.Resource{
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
}

func resourceVdcGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdVdcGroupRead,
		CreateContext: resourceVcdVdcGroupCreate,
		UpdateContext: resourceVcdVdcGroupUpdate,
		DeleteContext: resourceVcdVdcGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVdcGroupImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of VDC group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "VDC group description",
			},
			"dfw_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Distributed firewall status",
			},
			"starting_vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Starting VDC ID",
			},
			"participating_vdc_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Participating VDC IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_policy_status": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Default Policy Status",
			},
			"remove_default_firewall_rule": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "A flag to remove default firewall rule when DFW and Default Policy are both enabled ",
			},
			"error_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "More detailed error message when VDC group has error status",
			},
			"local_egress": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status whether local egress is enabled for a universal router belonging to a universal VDC group",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of used network pool",
			},
			"network_pool_universal_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The network provider’s universal id that is backing the universal network pool",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the networking provider backing the VDC Group",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The status that the group can be in (e.g. 'SAVING', 'SAVED', 'CONFIGURING'," +
					" 'REALIZED', 'REALIZATION_FAILED', 'DELETING', 'DELETE_FAILED', 'OBJECT_NOT_FOUND'," +
					" 'UNCONFIGURED')",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the group as LOCAL or UNIVERSAL",
			},
			"universal_networking_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True means that a VDC group router has been created",
			},
			"participating_org_vdcs": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The list of organization VDCs that are participating in this group",
				Elem:        participatingOrgVdcsResource,
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Forces deletion of VDC Group during destroy",
			},
		},
	}
}

// resourceVcdVdcGroupCreate covers Create functionality for resource
func resourceVcdVdcGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	diagErr := isInvalidPropertySetup(d.Get("dfw_enabled").(bool),
		d.Get("default_policy_status").(bool),
		d.Get("remove_default_firewall_rule").(bool))
	if diagErr != nil {
		return diagErr
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroupConfig := getVdcGroupConfigurationType(d)
	createdVdcGroup, err := adminOrg.CreateNsxtVdcGroup(vdcGroupConfig.Name, vdcGroupConfig.Description, vdcGroupConfig.StartingVdcId, vdcGroupConfig.ParticipatingVdcIds)
	if err != nil {
		return diag.Errorf("error creating VDC group: %s", err)
	}

	if d.Get("dfw_enabled").(bool) {
		createdVdcGroup, err = createdVdcGroup.ActivateDfw()
		if err != nil {
			return diag.Errorf("error enabling DFW for VDC group: %s", err)
		}
		// by default, default policy will be enabled when DFW is enabled, so only need code for disabling
		if !d.Get("default_policy_status").(bool) {
			createdVdcGroup, err = createdVdcGroup.DisableDefaultPolicy()
			if err != nil {
				return diag.Errorf("error disabling default policy for VDC group: %s", err)
			}
		}

		// If dfw_enabled and default_policy_status are both true - we can evaluate optional setting
		// to disable default firewall rule which might interfere with how
		if d.Get("default_policy_status").(bool) && d.Get("remove_default_firewall_rule").(bool) {
			err := createdVdcGroup.DeleteAllDistributedFirewallRules()
			if err != nil {
				return diag.Errorf("error removing default firewall rule: %s", err)
			}
		}

	}

	d.SetId(createdVdcGroup.VdcGroup.Id)
	return resourceVcdVdcGroupRead(ctx, d, meta)
}

func isInvalidPropertySetup(dfw_enabled, default_policy_status, remove_default_firewall_rule bool) diag.Diagnostics {
	if !dfw_enabled && default_policy_status {
		return diag.Errorf("`default_policy_status` must be `false` when `dfw_enabled` is `false`.")
	}

	if remove_default_firewall_rule && !dfw_enabled && !default_policy_status {
		return diag.Errorf("'remove_default_firewall_rule' can only be 'true' when 'dfw_enabled=true' and 'default_policy_status=true'")
	}
	return nil
}

// resourceVcdVdcGroupUpdate covers Update functionality for resource
func resourceVcdVdcGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Return immediately if only 'force_delete' value was changed as it only affects delete
	// operation, but one must be able to update it so that it can be changed for existing resource
	if !d.HasChangeExcept("force_delete") {
		return resourceVcdVdcGroupRead(ctx, d, meta)
	}

	diagErr := isInvalidPropertySetup(d.Get("dfw_enabled").(bool),
		d.Get("default_policy_status").(bool),
		d.Get("remove_default_firewall_rule").(bool))
	if diagErr != nil {
		return diagErr
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[VDC group update] : %s", err)
	}

	if d.HasChanges("name", "description", "participating_vdc_ids") {
		vdcGroupConfig := getVdcGroupConfigurationType(d)
		vdcGroup, err = vdcGroup.Update(vdcGroupConfig.Name, vdcGroupConfig.Description, vdcGroupConfig.ParticipatingVdcIds)
		if err != nil {
			return diag.Errorf("[VDC group update] : %s", err)
		}
	}

	if d.HasChange("dfw_enabled") {
		if d.Get("dfw_enabled").(bool) {
			vdcGroup, err = vdcGroup.ActivateDfw()
		} else {
			vdcGroup, err = vdcGroup.DisableDefaultPolicy()
			// ignore if it isn't possible to change
			if err != nil && err.Error() != "DFW has to be enabled before changing Default policy" {
				return diag.Errorf("error disabling default policy for VDC group: %s", err)
			}
			vdcGroup, err = vdcGroup.DeactivateDfw()
		}
		if err != nil {
			return diag.Errorf("error activating/deactivating DFW for VDC group: %s", err)
		}
	}

	if d.HasChange("default_policy_status") || !d.Get("default_policy_status").(bool) {
		errDiag := applyDefaultPolicy(d, vdcGroup)
		if errDiag != nil {
			return errDiag
		}
	}

	if d.HasChange("remove_default_firewall_rule") && d.Get("remove_default_firewall_rule").(bool) && d.Get("default_policy_status").(bool) {
		// If dfw_enabled and default_policy_status are both true - we can evaluate optional setting
		// to disable default firewall rule which might interfere with how
		err := vdcGroup.DeleteAllDistributedFirewallRules()
		if err != nil {
			return diag.Errorf("error removing default firewall rule: %s", err)
		}
	}

	return resourceVcdVdcGroupRead(ctx, d, meta)
}

func applyDefaultPolicy(d *schema.ResourceData, vdcGroup *govcd.VdcGroup) diag.Diagnostics {
	var err error
	if !d.Get("default_policy_status").(bool) {
		_, err = vdcGroup.DisableDefaultPolicy()
	} else {
		_, err = vdcGroup.EnableDefaultPolicy()
	}
	// ignore if it isn't possible to change
	if err != nil && err.Error() != "DFW has to be enabled before changing Default policy" {
		return diag.Errorf("error disabling/enabling default policy for VDC group: %s", err)
	}
	return nil
}

func getVdcGroupConfigurationType(d *schema.ResourceData) vdcGroupConfig {
	vdcIds := convertSchemaSetToSliceOfStrings(d.Get("participating_vdc_ids").(*schema.Set))

	return vdcGroupConfig{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		StartingVdcId:       d.Get("starting_vdc_id").(string),
		ParticipatingVdcIds: vdcIds,
	}
}

func resourceVcdVdcGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[VDC group read] : %s", err)
	}

	defaultValueStatus, err := getDefaultPolicyStatus(vdcGroup)
	if err != nil {
		return diag.Errorf("[VDC group read] : %s", err)
	}

	err = setVdcGroupConfigurationData(vdcGroup.VdcGroup, d, defaultValueStatus)
	if err != nil {
		return diag.Errorf("[VDC group read] : %s", err)
	}

	var participatingVdcIds []interface{}
	for _, participatingVdc := range vdcGroup.VdcGroup.ParticipatingOrgVdcs {
		participatingVdcIds = append(participatingVdcIds, participatingVdc.VdcRef.ID)
	}
	if len(participatingVdcIds) > 0 {
		err = d.Set("participating_vdc_ids", participatingVdcIds)
		if err != nil {
			return diag.Errorf("[VDC group read] could not set participating_vdc_ids block: %s", err)
		}
	}
	return nil
}

func getDefaultPolicyStatus(vdcGroup *govcd.VdcGroup) (*bool, error) {
	dfwPolicies, err := vdcGroup.GetDfwPolicies()
	if err != nil {
		return nil, fmt.Errorf("[VDC group read] : %s", err)
	}
	var defaultValueStatus *bool
	if dfwPolicies != nil && dfwPolicies.DefaultPolicy != nil {
		defaultValueStatus = dfwPolicies.DefaultPolicy.Enabled
	}

	return defaultValueStatus, nil
}

func setVdcGroupConfigurationData(config *types.VdcGroup, d *schema.ResourceData, defaultPolicyStatus *bool) error {
	dSet(d, "name", config.Name)
	dSet(d, "description", config.Description)
	dSet(d, "dfw_enabled", config.DfwEnabled)
	dSet(d, "error_message", config.ErrorMessage)
	dSet(d, "local_egress", config.LocalEgress)
	dSet(d, "network_pool_id", config.NetworkPoolId)
	dSet(d, "network_pool_universal_id", config.NetworkPoolUniversalId)
	dSet(d, "network_provider_type", config.NetworkProviderType)
	dSet(d, "status", config.Status)
	dSet(d, "type", config.Type)
	dSet(d, "universal_networking_enabled", config.UniversalNetworkingEnabled)
	if defaultPolicyStatus != nil {
		dSet(d, "default_policy_status", *defaultPolicyStatus)
	} else {
		dSet(d, "default_policy_status", false)
	}

	var candidateVdcsSlice []interface{}
	if len(config.ParticipatingOrgVdcs) > 0 {
		for _, candidateVdc := range config.ParticipatingOrgVdcs {

			candidateVdcMap := make(map[string]interface{})
			candidateVdcMap["fault_domain_tag"] = candidateVdc.FaultDomainTag
			candidateVdcMap["network_provider_scope"] = candidateVdc.NetworkProviderScope
			candidateVdcMap["is_remote_org"] = candidateVdc.RemoteOrg
			candidateVdcMap["status"] = candidateVdc.Status
			candidateVdcMap["org_name"] = candidateVdc.OrgRef.Name
			candidateVdcMap["org_id"] = candidateVdc.OrgRef.ID
			candidateVdcMap["site_name"] = candidateVdc.SiteRef.Name
			candidateVdcMap["site_id"] = candidateVdc.SiteRef.ID
			candidateVdcMap["vdc_name"] = candidateVdc.VdcRef.Name
			candidateVdcMap["vdc_id"] = candidateVdc.VdcRef.ID

			candidateVdcsSlice = append(candidateVdcsSlice, candidateVdcMap)
		}
	}

	err := d.Set("participating_org_vdcs", schema.NewSet(schema.HashResource(participatingOrgVdcsResource), candidateVdcsSlice))
	if err != nil {
		return fmt.Errorf("[VDC group read] could not set participating_org_vdcs block: %s", err)
	}
	return nil
}

func resourceVcdVdcGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrg, err)
	}

	vdcGroupToDelete, err := adminOrg.GetVdcGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[VDC group delete] : %s", err)
	}

	if vdcGroupToDelete.VdcGroup.DfwEnabled {
		vdcGroupToDelete, err = vdcGroupToDelete.DisableDefaultPolicy()
		if err != nil {
			return diag.Errorf("error disabling default policy for VDC group delete: %s", err)
		}
		vdcGroupToDelete, err = vdcGroupToDelete.DeactivateDfw()
		if err != nil {
			return diag.Errorf("error deactivating DFW for VDC group delete: %s", err)
		}
	}

	forceDelete := d.Get("force_delete").(bool)
	return diag.FromErr(vdcGroupToDelete.ForceDelete(forceDelete))
}

func resourceVdcGroupImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-group-name")
	}
	orgName, vdcGroupName := resourceURI[0], resourceURI[1]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("[VDC group import] error retrieving org %s: %s", orgName, err)
	}

	vdcGroup, err := adminOrg.GetVdcGroupByName(vdcGroupName)
	if err != nil {
		return nil, fmt.Errorf("error importing VDC group item: %s", err)
	}

	defaultValueStatus, err := getDefaultPolicyStatus(vdcGroup)
	if err != nil {
		return nil, fmt.Errorf("error importing VDC group item: %s", err)
	}

	d.SetId(vdcGroup.VdcGroup.Id)
	dSet(d, "org", orgName)
	err = setVdcGroupConfigurationData(vdcGroup.VdcGroup, d, defaultValueStatus)
	if err != nil {
		return nil, fmt.Errorf("[VDC group import] : %s", err)
	}

	return []*schema.ResourceData{d}, nil
}

// vdcGroupConfig is a minimal structure defining a VdcGroup in Organization
type vdcGroupConfig struct {
	Name                string
	Description         string
	ParticipatingVdcIds []string
	StartingVdcId       string
}
