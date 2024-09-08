package vcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
	"strings"
)

func resourceVcdNsxtRouteAdvertisement() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		ReadContext:   resourceVcdNsxtRouteAdvertisementRead,
		UpdateContext: resourceVcdNsxtRouteAdvertisementCreateUpdate,
		DeleteContext: resourceVcdNsxtRouteAdvertisementDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtRouteAdvertisementImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T Edge Gateway ID in which route advertisement is located",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Defines if route advertisement is active",
			},
			"subnets": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Set of subnets that will be advertised to Tier-0 gateway. Empty means none",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdNsxtRouteAdvertisementCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on a route advertisement is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[routed advertisement create/update] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	var subnets []string
	enableRouteAdvertisement := d.Get("enabled").(bool)
	subnetsFromSchema, ok := d.GetOk("subnets")

	if ok {
		subnets = convertSchemaSetToSliceOfStrings(subnetsFromSchema.(*schema.Set))
	}

	if !enableRouteAdvertisement && len(subnets) > 0 {
		return diag.Errorf("if enable is set to false, no subnets must be passed")
	}

	_, edgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "route advertisement")
	if err != nil {
		return diag.FromErr(err)
	}

	err = checkNSXTEdgeGatewayDedicated(edgeGateway)
	if err != nil {
		return diag.Errorf("error when configuring route advertisement on NSX-T Edge Gateway - %s", err)
	}

	_, err = edgeGateway.UpdateNsxtRouteAdvertisement(enableRouteAdvertisement, subnets)
	if err != nil {
		return diag.Errorf("error when creating/updating route advertisement - %s", err)
	}

	d.SetId(edgeGateway.EdgeGateway.ID)

	return resourceVcdNsxtRouteAdvertisementRead(ctx, d, meta)
}

func resourceVcdNsxtRouteAdvertisementRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return diag.Errorf("error when getting Org name - %s", err)
	}

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	routeAdvertisement, err := nsxtEdge.GetNsxtRouteAdvertisement()
	if err != nil {
		return diag.Errorf("error while retrieving route advertisement - %s", err)
	}

	dSet(d, "enabled", routeAdvertisement.Enable)

	subnetSet := convertStringsToTypeSet(routeAdvertisement.Subnets)
	err = d.Set("subnets", subnetSet)
	if err != nil {
		return diag.Errorf("error while setting subnets argument: %s", err)
	}

	return nil
}

func resourceVcdNsxtRouteAdvertisementDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on a route advertisement is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[route advertisement delete] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return diag.Errorf("error when getting Org name - %s", err)
	}

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	err = nsxtEdge.DeleteNsxtRouteAdvertisement()
	if err != nil {
		return diag.Errorf("error while deleting route advertisement - %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdNsxtRouteAdvertisementImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway Route Advertisement import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)

	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

// checkNSXTEdgeGatewayDedicated is a simple helper function that checks if "Using Dedicated Provider Router" option is enabled
// on NSX-T Edge Gateway so that route advertisement can be configured. If not it returns an error.
func checkNSXTEdgeGatewayDedicated(nsxtEdgeGw *govcd.NsxtEdgeGateway) error {
	if !nsxtEdgeGw.EdgeGateway.EdgeGatewayUplinks[0].Dedicated {
		return fmt.Errorf("NSX-T Edge Gateway is not using a dedicated provider router. Please enable this feature before configuring route advertisement")
	}

	return nil
}
