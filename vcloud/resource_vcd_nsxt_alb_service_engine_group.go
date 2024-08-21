package vcloud

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbServiceEngineGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbServiceEngineGroupCreate,
		ReadContext:   resourceVcdAlbServiceEngineGroupRead,
		UpdateContext: resourceVcdAlbServiceEngineGroupUpdate,
		DeleteContext: resourceVcdAlbServiceEngineGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbServiceEngineGroupImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Service Engine Group name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T ALB Service Engine Group description",
			},
			"alb_cloud_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB backing Cloud ID",
			},
			"reservation_model": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "NSX-T ALB Service Engine Group reservation model. One of 'DEDICATED', 'SHARED'",
				ValidateFunc: validation.StringInSlice([]string{"DEDICATED", "SHARED"}, false),
			},
			// Ideally this should be a reference by ID and a data source for lookup. However, the Importable Service
			// Engine Group API endpoint does not return entities once they are consumed, and it is impossible to make
			// a data source.
			"importable_service_engine_group_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Importable Service Engine Group Name",
			},
			"max_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group maximum virtual services",
			},
			"reserved_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group reserved virtual services",
			},
			"deployed_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group deployed virtual services",
			},
			"ha_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group HA mode",
			},
			"overallocated": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Boolean value that shows if virtual services are overallocated",
			},
			"sync_on_refresh": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value that shows if sync should be performed on every refresh",
			},
			"supported_feature_set": {
				Type:         schema.TypeString,
				Optional:     true,
				Required:     false, // It should be required but for VCD < 10.4 compatibility it is not
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"STANDARD", "PREMIUM"}, false),
				Description:  "Feature set for this ALB Service Engine Group. One of 'STANDARD', 'PREMIUM'.",
			},
		},
	}
}

func resourceVcdAlbServiceEngineGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	// Lookup Importable Service Engine Group
	albImportableSeGroup, err := vcdClient.GetAlbImportableServiceEngineGroupByName(
		d.Get("alb_cloud_id").(string), d.Get("importable_service_engine_group_name").(string))
	if err != nil {
		return diag.Errorf("unable to find Importable Service Engine Group by Name '%s': %s",
			d.Get("importable_service_engine_group_name").(string), err)
	}

	albSeGroupConfig := getNsxtAlbServiceEngineGroupType(d, albImportableSeGroup.NsxtAlbImportableServiceEngineGroups.ID)
	createdAlbController, err := vcdClient.CreateNsxtAlbServiceEngineGroup(albSeGroupConfig)
	if err != nil {
		return diag.Errorf("error creating NSX-T ALB Service Engine Group '%s': %s", albSeGroupConfig.Name, err)
	}

	d.SetId(createdAlbController.NsxtAlbServiceEngineGroup.ID)

	return resourceVcdAlbServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbServiceEngineGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	// If the only value for update is for 'sync_on_refresh' flag - there is no need to perform any API calls
	if !d.HasChangeExcept("sync_on_refresh") {
		return resourceVcdAlbServiceEngineGroupRead(ctx, d, meta)
	}

	albSeGroup, err := vcdClient.GetAlbServiceEngineGroupById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Service Engine Group: %s", err)
	}

	// Feeding back in Importable Service Engine Group ID to avoid reading it by name again.
	updateSeGroupConfig := getNsxtAlbServiceEngineGroupType(d, albSeGroup.NsxtAlbServiceEngineGroup.ServiceEngineGroupBacking.BackingId)
	updateSeGroupConfig.ID = d.Id()

	_, err = albSeGroup.Update(updateSeGroupConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T ALB Service Engine Group: %s", err)
	}

	return resourceVcdAlbServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbServiceEngineGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albSeGroup, err := vcdClient.GetAlbServiceEngineGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to find NSX-T ALB Service Engine Group: %s", err)
	}

	// If "Sync" is requested for read operations - perform Sync operation and re-read the entity to get latest data
	if d.Get("sync_on_refresh").(bool) {
		err = albSeGroup.Sync()
		if err != nil {
			return diag.Errorf("error executing Sync operation for NSX-T ALB Service Engine Group '%s': %s",
				albSeGroup.NsxtAlbServiceEngineGroup.Name, err)
		}

		// re-read new values post sync
		albSeGroup, err = vcdClient.GetAlbServiceEngineGroupById(albSeGroup.NsxtAlbServiceEngineGroup.ID)
		if err != nil {
			return diag.Errorf("error re-reading NSX-T ALB Service Engine Group after Sync: %s", err)
		}
	}

	setNsxtAlbServiceEngineGroupData(d, albSeGroup.NsxtAlbServiceEngineGroup)

	return nil
}

func resourceVcdAlbServiceEngineGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albSeGroup, err := vcdClient.GetAlbServiceEngineGroupById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Service Engine Group: %s", err)
	}

	err = albSeGroup.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T ALB Service Engine Group: %s", err)
	}

	return nil
}

func resourceVcdAlbServiceEngineGroupImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("this resource is only supported for Providers")
	}

	resourceURI := d.Id()
	albSeGroup, err := vcdClient.GetAlbServiceEngineGroupByName("", resourceURI)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Service Engine Group with Name '%s': %s", d.Id(), err)
	}

	// This value is an internal flag and it cannot be read from resource itself. However, it makes sense to set it to
	// default value in configuration. That way plan after import should be clean
	dSet(d, "sync_on_refresh", false)

	d.SetId(albSeGroup.NsxtAlbServiceEngineGroup.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxtAlbServiceEngineGroupType(d *schema.ResourceData, impServiceEngineGroupId string) *types.NsxtAlbServiceEngineGroup {
	albControllerType := &types.NsxtAlbServiceEngineGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		ServiceEngineGroupBacking: types.ServiceEngineGroupBacking{
			BackingId: impServiceEngineGroupId,
			LoadBalancerCloudRef: &types.OpenApiReference{
				ID: d.Get("alb_cloud_id").(string),
			},
		},
		ReservationType:     d.Get("reservation_model").(string),
		SupportedFeatureSet: d.Get("supported_feature_set").(string),
	}
	return albControllerType
}

func setNsxtAlbServiceEngineGroupData(d *schema.ResourceData, albController *types.NsxtAlbServiceEngineGroup) {
	dSet(d, "name", albController.Name)
	dSet(d, "description", albController.Description)
	dSet(d, "reservation_model", albController.ReservationType)
	dSet(d, "alb_cloud_id", albController.ServiceEngineGroupBacking.LoadBalancerCloudRef.ID)

	dSet(d, "max_virtual_services", albController.MaxVirtualServices)
	dSet(d, "reserved_virtual_services", albController.ReservedVirtualServices)
	dSet(d, "deployed_virtual_services", albController.NumDeployedVirtualServices)
	dSet(d, "ha_mode", albController.HaMode)
	dSet(d, "overallocated", albController.OverAllocated)
	dSet(d, "supported_feature_set", albController.SupportedFeatureSet)
}
