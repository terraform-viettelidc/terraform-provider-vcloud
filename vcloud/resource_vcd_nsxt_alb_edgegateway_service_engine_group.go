package vcloud

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbEdgeGatewayServiceEngineGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupCreate,
		UpdateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupUpdate,
		ReadContext:   resourceVcdAlbEdgeGatewayServiceEngineGroupRead,
		DeleteContext: resourceVcdAlbEdgeGatewayServiceEngineGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbEdgeGatewayServiceEngineGroupImport,
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
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which ALB Service Engine Group should be located",
			},
			"service_engine_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Engine Group ID to attach to this NSX-T Edge Gateway",
			},
			"service_engine_group_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service Engine Group Name which is attached to NSX-T Edge Gateway",
			},
			"max_virtual_services": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of virtual services to be used in this Service Engine Group",
			},
			"reserved_virtual_services": {
				// This field could be TypeInt, but Terraform cannot differentiate if a value is
				// empty or '0'. TypeString solves this problem by differentiating empty string
				// and "0".
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Number of reserved virtual services for this Service Engine Group",
				ValidateFunc: IsIntAndAtLeast(0),
			},
			"deployed_virtual_services": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of deployed virtual services for this Service Engine Group",
			},
		},
	}
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignmentConfig := getAlbServiceEngineGroupAssignmentType(d)
	edgeAlbServiceEngineGroupAssignment, err := vcdClient.CreateAlbServiceEngineGroupAssignment(edgeAlbServiceEngineGroupAssignmentConfig)
	if err != nil {
		return diag.Errorf("error creating ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	d.SetId(edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return resourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}
	edgeAlbServiceEngineGroupAssignmentConfig := getAlbServiceEngineGroupAssignmentType(d)
	// Add correct ID for update
	edgeAlbServiceEngineGroupAssignmentConfig.ID = edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID
	updatedEdgeAlbServiceEngineGroupAssignment, err := edgeAlbServiceEngineGroupAssignment.Update(edgeAlbServiceEngineGroupAssignmentConfig)
	if err != nil {
		return diag.Errorf("error updating ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	d.SetId(updatedEdgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return resourceVcdAlbEdgeGatewayServiceEngineGroupRead(ctx, d, meta)
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("ALB Service Engine Group assignment not found. Removing from state file: %s", err)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading ALB Service Engine Group assignment: %s", err)
	}
	setAlbServiceEngineGroupAssignmentData(d, edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment)
	d.SetId(edgeAlbServiceEngineGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return nil
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := validateEdgeGatewayIdParent(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeAlbServiceEngineGroupAssignment, err := vcdClient.GetAlbServiceEngineGroupAssignmentById(d.Id())
	if err != nil {
		return diag.Errorf("error reading ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}

	err = edgeAlbServiceEngineGroupAssignment.Delete()
	if err != nil {
		return diag.Errorf("error deleting ALB Service Engine Group assignment to Edge Gateway: %s", err)
	}
	return nil
}

func resourceVcdAlbEdgeGatewayServiceEngineGroupImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T ALB Service Engine Group assignment import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.vdc-name.nsxt-edge-gw-name.se-group-name")
	}
	orgName, vdcOrVdcGroupName, edgeName, seGroupName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("this resource is only supported for NSX-T Edge Gateways")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	queryParams := url.Values{}
	queryParams.Add("filter", fmt.Sprintf("gatewayRef.id==%s", edge.EdgeGateway.ID))

	// GetFilteredAlbServiceEngineGroupAssignmentByName
	seGroupAssignment, err := vcdClient.GetFilteredAlbServiceEngineGroupAssignmentByName(seGroupName, queryParams)
	if err != nil {
		return nil, fmt.Errorf("errorr retrieving Servce Engine Group assignment to Edge Gateway with Name '%s': %s",
			seGroupName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	d.SetId(seGroupAssignment.NsxtAlbServiceEngineGroupAssignment.ID)
	return []*schema.ResourceData{d}, nil
}

func setAlbServiceEngineGroupAssignmentData(d *schema.ResourceData, t *types.NsxtAlbServiceEngineGroupAssignment) {
	dSet(d, "edge_gateway_id", t.GatewayRef.ID)
	dSet(d, "service_engine_group_id", t.ServiceEngineGroupRef.ID)
	dSet(d, "service_engine_group_name", t.ServiceEngineGroupRef.Name)
	dSet(d, "max_virtual_services", t.MaxVirtualServices)
	dSet(d, "deployed_virtual_services", t.NumDeployedVirtualServices)
	if t.MinVirtualServices != nil {
		dSet(d, "reserved_virtual_services", strconv.Itoa(*t.MinVirtualServices))
	}
}

func getAlbServiceEngineGroupAssignmentType(d *schema.ResourceData) *types.NsxtAlbServiceEngineGroupAssignment {
	edgeAlbServiceEngineAssignmentConfig := &types.NsxtAlbServiceEngineGroupAssignment{
		GatewayRef:            &types.OpenApiReference{ID: d.Get("edge_gateway_id").(string)},
		ServiceEngineGroupRef: &types.OpenApiReference{ID: d.Get("service_engine_group_id").(string)},
	}

	// Max Virtual Services and Reserved Virtual Services only work with SHARED Service Engine Group, but validation
	// enforcement is left for VCD API.
	if maxServicesInterface, isSet := d.GetOk("max_virtual_services"); isSet {
		edgeAlbServiceEngineAssignmentConfig.MaxVirtualServices = addrOf(maxServicesInterface.(int))
	}

	if reservedServicesInterface, isSet := d.GetOk("reserved_virtual_services"); isSet {
		reservedServicesInterfaceString := reservedServicesInterface.(string)
		// Ignoring error of `strconv.Atoi` because there is a validator enforced in schema field
		// 'reserved_virtual_services' - IsIntAndAtLeast(0),
		reservedServicesInt, _ := strconv.Atoi(reservedServicesInterfaceString)
		edgeAlbServiceEngineAssignmentConfig.MinVirtualServices = &reservedServicesInt
	}

	return edgeAlbServiceEngineAssignmentConfig
}

// validateEdgeGatewayIdParent validates if specified field `edge_gateway_id` exists in defined Org and VDC
func validateEdgeGatewayIdParent(d *schema.ResourceData, vcdClient *VCDClient) error {
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return fmt.Errorf("error retrieving Org and VDC")
	}

	_, err = vcdClient.GetNsxtEdgeGatewayFromResourceById(d, "edge_gateway_id")
	if err != nil {
		return fmt.Errorf("unable to locate NSX-T Edge Gateway with ID '%s' in Org '%s': %s",
			d.Get("edge_gateway_id").(string), org.Org.Name, err)
	}

	return nil
}
