package vcloud

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVappNetworkStaticRouting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVappNetworkStaticRoutingCreate,
		DeleteContext: resourceVAppNetworkStaticRoutingDelete,
		ReadContext:   resourceVappNetworkStaticRoutingRead,
		UpdateContext: resourceVappNetworkStaticRoutingUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: vappNetworkStaticRoutingImport,
		},

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
			"vapp_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp identifier",
			},
			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network identifier",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enable or disable static Routing. Default is `true`.",
			},
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name for the static route.",
						},
						"network_cidr": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "network specification in CIDR.",
						},
						"next_hop_ip": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "IP Address of Next Hop router/gateway.",
						},
					},
				},
			},
		},
	}
}

func resourceVappNetworkStaticRoutingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVappNetworkStaticRoutingUpdate(ctx, d, meta)
}

func resourceVappNetworkStaticRoutingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}
	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	networkId := d.Get("network_id").(string)
	staticRouting, err := expandVappNetworkStaticRouting(d)
	if err != nil {
		return diag.Errorf("error expanding static routes: %s", err)
	}
	vappNetwork, err := vapp.UpdateNetworkStaticRouting(networkId, staticRouting, d.Get("enabled").(bool))
	if err != nil {
		log.Printf("[INFO] Error setting static routing: %s", err)
		return diag.Errorf("error setting static routing: %s", err)
	}

	d.SetId(vappNetwork.ID)

	return resourceVappNetworkStaticRoutingRead(ctx, d, meta)
}

func resourceVAppNetworkStaticRoutingDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}

	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	err = vapp.RemoveAllNetworkStaticRoutes(d.Get("network_id").(string))
	if err != nil {
		log.Printf("[INFO] Error deleting static routes: %s", err)
		return diag.Errorf("error deleting static routes: %s", err)
	}

	return nil
}

func resourceVappNetworkStaticRoutingRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}

	vappNetwork, err := vapp.GetVappNetworkById(d.Get("network_id").(string), false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("vApp network not found. Removing static routing from state file: %s", err)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error finding vApp network. %s", err)
	}

	var rules []map[string]interface{}
	if vappNetwork.Configuration.Features == nil || vappNetwork.Configuration.Features.StaticRoutingService == nil {
		log.Print("[INFO] no Static routes found.")
		dSet(d, "rule", nil)
	}

	for _, rule := range vappNetwork.Configuration.Features.StaticRoutingService.StaticRoute {
		singleRule := make(map[string]interface{})

		singleRule["name"] = rule.Name
		singleRule["network_cidr"] = rule.Network
		singleRule["next_hop_ip"] = rule.NextHopIP
		rules = append(rules, singleRule)
	}
	dSet(d, "enabled", vappNetwork.Configuration.Features.StaticRoutingService.IsEnabled)
	err = d.Set("rule", rules)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func expandVappNetworkStaticRouting(d *schema.ResourceData) ([]*types.StaticRoute, error) {
	var staticRoutes []*types.StaticRoute
	for _, singleRule := range d.Get("rule").([]interface{}) {
		configuredRule := singleRule.(map[string]interface{})
		rule := &types.StaticRoute{
			Network:   configuredRule["network_cidr"].(string),
			Name:      configuredRule["name"].(string),
			NextHopIP: configuredRule["next_hop_ip"].(string),
		}
		staticRoutes = append(staticRoutes, rule)
	}

	return staticRoutes, nil
}

// vappNetworkStaticRoutingImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_static_routing.my_existing_static_routing_rules
// Example import path (_the_id_string_): org.my_existing_vdc.vapp_name.network_name or org.my_existing_vdc.vapp_id.network_id
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func vappNetworkStaticRoutingImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return vappNetworkRuleImport(d, meta, "vcd_vapp_static_routing")
}
