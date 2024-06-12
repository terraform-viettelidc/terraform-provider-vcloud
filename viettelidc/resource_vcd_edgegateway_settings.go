package viettelidc

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdEdgeGatewaySettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdEdgeGatewaySettingsCreate,
		Read:   resourceVcdEdgeGatewaySettingsRead,
		Update: resourceVcdEdgeGatewaySettingsUpdate,
		Delete: resourceVcdEdgeGatewaySettingsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdEdgeGatewaySettingsImport,
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
			"edge_gateway_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Name of the edge gateway. Required when 'edge_gateway_id' is not set",
				ExactlyOneOf: []string{"edge_gateway_id", "edge_gateway_name"},
			},
			"edge_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "ID of the edge gateway. Required when 'edge_gateway_name' is not set",
				ExactlyOneOf: []string{"edge_gateway_id", "edge_gateway_name"},
			},
			"lb_enabled": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancing. (Disabled by default)",
			},
			"lb_acceleration_enabled": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancer acceleration. (Disabled by default)",
			},
			"lb_logging_enabled": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancer logging. (Disabled by default)",
				// Due to a VCD bug, this field can only be changed by a system administrator
			},
			"lb_loglevel": {
				Type:         schema.TypeString,
				Default:      "info",
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Log level. One of 'emergency', 'alert', 'critical', 'error', " +
					"'warning', 'notice', 'info', 'debug'. ('info' by default)",
				// Due to a VCD bug, this field can only be changed by a system administrator
			},
			"fw_enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Enable firewall. Default 'true'",
			},
			"fw_default_rule_logging_enabled": {
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable logging for default rule. Default 'false'",
			},
			"fw_default_rule_action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "deny",
				Description:  "'accept' or 'deny'. Default 'deny'",
				ValidateFunc: validation.StringInSlice([]string{"accept", "deny"}, false),
			},
		},
	}
}

func resourceVcdEdgeGatewaySettingsCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdEdgeGatewaySettingsUpdate(d, meta)
}

func getVcdEdgeGateway(d *schema.ResourceData, meta interface{}) (*govcd.EdgeGateway, error) {

	log.Printf("[TRACE] edge gateway settings read initiated")

	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[getVcdEdgeGateway] error retrieving org and vdc: %s", err)
	}
	var edgeGateway *govcd.EdgeGateway

	// Preferred identification method is by ID
	identifier := d.Get("edge_gateway_id").(string)
	if identifier == "" {
		// Alternative method is by name
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, fmt.Errorf("[getVcdEdgeGateway] no identifier provided")
	}
	edgeGateway, err = vdc.GetEdgeGatewayByNameOrId(identifier, false)
	if err != nil {
		return nil, fmt.Errorf("[getVcdEdgeGateway] edge gateway %s not found: %s", identifier, err)
	}
	return edgeGateway, nil
}

func resourceVcdEdgeGatewaySettingsRead(d *schema.ResourceData, meta interface{}) error {
	edgeGateway, err := getVcdEdgeGateway(d, meta)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			log.Printf("[edgegateway settings read] edge gateway not found. Removing from state file: %s", err)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[edgegateway settings read] edge gateway not found: %s", err)
	}
	if err := setLoadBalancerData(d, *edgeGateway); err != nil {
		return err
	}

	if err := setFirewallData(d, *edgeGateway); err != nil {
		return err
	}

	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	dSet(d, "edge_gateway_name", edgeGateway.EdgeGateway.Name)
	d.SetId(edgeGateway.EdgeGateway.ID)

	log.Printf("[TRACE] edge gateway settings read completed: %#v", edgeGateway.EdgeGateway)
	return nil
}

func resourceVcdEdgeGatewaySettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	edgeGateway, err := getVcdEdgeGateway(d, meta)
	if err != nil {
		return err
	}

	if d.HasChange("lb_enabled") || d.HasChange("lb_acceleration_enabled") ||
		d.HasChange("lb_logging_enabled") || d.HasChange("lb_loglevel") {
		err := updateLoadBalancer(d, *edgeGateway)
		if err != nil {
			return err
		}
	}

	if d.HasChange("fw_enabled") || d.HasChange("fw_default_rule_logging_enabled") ||
		d.HasChange("fw_default_rule_action") {
		err := updateFirewall(d, *edgeGateway)
		if err != nil {
			return err
		}
	}

	log.Printf("[TRACE] edge gateway settings update completed: %#v", edgeGateway.EdgeGateway)
	return resourceVcdEdgeGatewaySettingsRead(d, meta)
}

func resourceVcdEdgeGatewaySettingsDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdEdgeGatewaySettingsUpdate(d, meta)
}

// resourceVcdEdgeGatewaySettingsImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_edgegateway_settings.my-edge-gateway-name
// Example import path (_the_id_string_): org.vdc.my-edge-gw
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
// Note: the edge gateway can be identified by either the name or the ID
func resourceVcdEdgeGatewaySettingsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[resourceVcdEdgeGatewaySettingsImport] resource name must be specified as org-name.vdc-name.edge-gw-name (or edge-gw-ID)")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find org %s: %s", orgName, err)
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
	}
	edgeGateway, err := vdc.GetEdgeGatewayByNameOrId(edgeName, false)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway_name", edgeGateway.EdgeGateway.Name)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(edgeGateway.EdgeGateway.ID)
	return []*schema.ResourceData{d}, nil
}
