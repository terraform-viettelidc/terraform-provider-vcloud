package vcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVappOrgNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVappOrgNetworkCreate,
		ReadContext:   resourceVappOrgNetworkRead,
		UpdateContext: resourceVappOrgNetworkUpdate,
		DeleteContext: resourceVappAndVappOrgNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVappOrgNetworkImport,
		},

		Schema: map[string]*schema.Schema{
			"vapp_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network name",
			},
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
			"org_network_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization network name to which vApp network is connected to",
			},
			"is_fenced": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Fencing allows identical virtual machines in different vApp networks connect to organization VDC networks that are accessed in this vApp",
			},
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
			"reboot_vapp_on_removal": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the vApp should be rebooted when the vApp network is removed. Default is false.",
			},
		},
	}
}

func resourceVappOrgNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return diag.Errorf("error finding vApp: %s and err: %s", vappName, err)
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		RetainIpMacEnabled: addrOf(d.Get("retain_ip_mac_enabled").(bool)),
	}

	orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(d.Get("org_network_name").(string), true)
	if err != nil {
		return diag.FromErr(err)
	}

	vAppNetworkConfig, err := vapp.AddOrgNetwork(vappNetworkSettings, orgNetwork.OrgVDCNetwork, d.Get("is_fenced").(bool))
	if err != nil {
		return diag.Errorf("error creating vApp org network. %#v", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == orgNetwork.OrgVDCNetwork.Name {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		return diag.Errorf("didn't find vApp network: %s", d.Get("name").(string))
	}

	// Parsing UUID from 'https://HOST/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset'
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF, false)
	if err != nil {
		return diag.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	return resourceVappOrgNetworkRead(ctx, d, meta)
}

func resourceVappOrgNetworkRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVappOrgNetworkRead(d, meta, "resource")
}

func genericVappOrgNetworkRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			log.Printf("vApp not found. Removing vApp network from state file: %s", err)
			d.SetId("")
			return nil
		}
		return diag.Errorf("error finding Vapp: %s", err)
	}

	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return diag.Errorf("error getting vApp networks: %s", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	var networkId string
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.Link != nil {
			networkId, err = govcd.GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return diag.Errorf("unable to get network ID from HREF: %s", err)
			}
			// name check needed for datasource to find network as don't have ID
			if extractUuid(d.Id()) == extractUuid(networkId) || networkConfig.NetworkName == d.Get("org_network_name").(string) {
				vAppNetwork = networkConfig
				break
			}
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		if origin == "resource" {
			log.Printf("[DEBUG] Network no longer exists. Removing from tfstate")
			d.SetId("")
			return nil
		}
		return diag.Errorf("[VAPP org network read] %s : %s", govcd.ErrorEntityNotFound, d.Get("org_network_name").(string))
	}

	// needs to set for datasource
	if d.Id() == "" {
		d.SetId(normalizeId("urn:vcloud:network:", networkId))
	}

	dSet(d, "retain_ip_mac_enabled", *vAppNetwork.Configuration.RetainNetInfoAcrossDeployments)

	isFenced := false
	if vAppNetwork.Configuration.FenceMode == types.FenceModeNAT {
		isFenced = true
	}
	dSet(d, "is_fenced", isFenced)
	return nil
}

func resourceVappOrgNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// reboot_vapp_on_removal does not have any effect on update therefore skipping update if only
	// this field was modified
	if !d.HasChangeExcept("reboot_vapp_on_removal") {
		return resourceVappOrgNetworkRead(ctx, d, meta)
	}
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappName := d.Get("vapp_name").(string)
	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return diag.Errorf("error finding vApp: %s and err:  %s", vappName, err)
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		ID:                 d.Id(),
		RetainIpMacEnabled: addrOf(d.Get("retain_ip_mac_enabled").(bool)),
	}

	_, err = vapp.UpdateOrgNetwork(vappNetworkSettings, d.Get("is_fenced").(bool))
	if err != nil {
		return diag.Errorf("error creating vApp network. %#v", err)
	}

	return resourceVappOrgNetworkRead(ctx, d, meta)
}

// resourceVcdVappOrgNetworkImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_org_network.org_network_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name.org-network-name
func resourceVcdVappOrgNetworkImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("[vApp org network import] resource name must be specified as org-name.vdc-name.vapp-name.org-network-name")
	}
	orgName, vdcName, vappName, networkName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] error retrieving vapp %s: %s", vappName, err)
	}
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("[vApp org network import] error retrieving vApp network configuration %s: %s", networkName, err)
	}

	vappNetworkToImport := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			vappNetworkToImport = networkConfig
			break
		}
	}

	if vappNetworkToImport == (types.VAppNetworkConfiguration{}) {
		return nil, fmt.Errorf("didn't find vApp org network: %s", networkName)
	}

	if govcd.IsVappNetwork(vappNetworkToImport.Configuration) {
		return nil, fmt.Errorf("found vApp network, not vApp org network: %s", networkName)
	}

	networkId, err := govcd.GetUuidFromHref(vappNetworkToImport.Link.HREF, false)
	if err != nil {
		return nil, fmt.Errorf("unable to get network ID from HREF: %s", err)
	}

	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	if vcdClient.Org != orgName {
		dSet(d, "org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		dSet(d, "vdc", vdcName)
	}
	dSet(d, "org_network_name", networkName)
	dSet(d, "vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}
