package vcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNetworkDirect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNetworkDirectCreate,
		ReadContext:   resourceVcdNetworkDirectRead,
		UpdateContext: resourceVcdNetworkDirectUpdate,
		DeleteContext: resourceVcdNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNetworkDirectImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name for this network",
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
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description for the network",
			},
			"external_network": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the external network",
			},
			"external_network_gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway of the external network",
			},
			"external_network_netmask": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Net mask of the external network",
			},
			"external_network_dns1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Main DNS of the external network",
			},
			"external_network_dns2": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary DNS of the external network",
			},
			"external_network_dns_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix of the external network",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hypertext Reference",
			},
			"shared": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Description:   "Key value map of metadata to assign to this network. Key and value can be any string",
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
			},
			"metadata_entry": metadataEntryResourceSchemaDeprecated("Network"),
		},
	}
}

func resourceVcdNetworkDirectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("creation of a vcd_network_direct requires system administrator privileges")
	}
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	externalNetworkName := d.Get("external_network").(string)
	networkName := d.Get("name").(string)
	externalNetwork, err := vcdClient.GetExternalNetworkByName(externalNetworkName)
	if err != nil {
		return diag.Errorf("unable to find external network %s (%s)", externalNetworkName, err)
	}

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        networkName,
		Description: d.Get("description").(string),
		Configuration: &types.NetworkConfiguration{
			ParentNetwork: &types.Reference{
				HREF: externalNetwork.ExternalNetwork.HREF,
				Type: externalNetwork.ExternalNetwork.Type,
				Name: externalNetwork.ExternalNetwork.Name,
			},
			FenceMode:                 "bridged",
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return diag.Errorf("error: %s", err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, true)
	if err != nil {
		return diag.Errorf("error retrieving network %s after creation", networkName)
	}
	d.SetId(network.OrgVDCNetwork.ID)

	err = createOrUpdateMetadata(d, network, "metadata")
	if err != nil {
		return diag.Errorf("error adding metadata to direct network: %s", err)
	}

	return resourceVcdNetworkDirectRead(ctx, d, meta)
}

func resourceVcdNetworkDirectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNetworkDirectRead(ctx, d, meta, "resource")
}

func genericVcdNetworkDirectRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	var diags diag.Diagnostics
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[direct network read] "+errorRetrievingOrgAndVdc, err)
	}

	network, err := getNetwork(d, vcdClient, origin == "datasource", "direct")
	if err != nil {
		if origin == "resource" {
			networkName := d.Get("name").(string)
			log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", networkName)
			d.SetId("")
			return nil
		}
		return diag.Errorf("[direct network read] network not found: %s", err)
	}

	dSet(d, "name", network.OrgVDCNetwork.Name)
	dSet(d, "href", network.OrgVDCNetwork.HREF)
	dSet(d, "shared", network.OrgVDCNetwork.IsShared)

	// Getting external network data through network list, as a direct call to external network
	// structure requires system admin privileges.
	// Org Users can't create a direct network, but should be able to see the connection info.
	networkList, err := vdc.GetNetworkList()
	if err != nil {
		return diag.Errorf("error retrieving network list for VDC %s : %s", vdc.Vdc.Name, err)
	}
	var currentNetwork *types.QueryResultOrgVdcNetworkRecordType
	for _, net := range networkList {
		if net.Name == network.OrgVDCNetwork.Name {
			currentNetwork = net
		}
	}
	if currentNetwork == nil {
		return diag.Errorf("error retrieving network %s from network list", network.OrgVDCNetwork.Name)
	}
	dSet(d, "external_network", currentNetwork.ConnectedTo)
	dSet(d, "external_network_netmask", currentNetwork.Netmask)
	dSet(d, "external_network_dns1", currentNetwork.Dns1)
	dSet(d, "external_network_dns2", currentNetwork.Dns2)
	dSet(d, "external_network_dns_suffix", currentNetwork.DnsSuffix)
	// Fixes issue #450
	dSet(d, "external_network_gateway", currentNetwork.DefaultGateway)
	dSet(d, "description", network.OrgVDCNetwork.Description)
	d.SetId(network.OrgVDCNetwork.ID)

	diags = append(diags, updateMetadataInStateDeprecated(d, vcdClient, "vcd_network_direct", network)...)
	if diags != nil && diags.HasError() {
		log.Printf("[DEBUG] Unable to set direct network metadata: %v", diags)
		return diags
	}

	// This must be checked at the end as updateMetadataInStateDeprecated can throw Warning diagnostics
	if len(diags) > 0 {
		return diags
	}
	return nil
}

func resourceVcdNetworkDirectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("update of a vcd_network_direct requires system administrator privileges")
	}
	network, err := getNetwork(d, vcdClient, false, "direct")
	if err != nil {
		return diag.Errorf("[direct network update] error getting network: %s", err)
	}

	networkName := d.Get("name").(string)
	network.OrgVDCNetwork.Name = networkName
	network.OrgVDCNetwork.Description = d.Get("description").(string)
	network.OrgVDCNetwork.IsShared = d.Get("shared").(bool)

	err = network.Update()
	if err != nil {
		return diag.Errorf("[direct network update] error updating network %s: %s", network.OrgVDCNetwork.Name, err)
	}

	err = createOrUpdateMetadata(d, network, "metadata")
	if err != nil {
		return diag.Errorf("[direct network update] error updating network metadata: %s", err)
	}

	return resourceVcdNetworkDirectRead(ctx, d, meta)
}

func getNetwork(d *schema.ResourceData, vcdClient *VCDClient, isDataSource bool, wanted string) (*govcd.OrgVDCNetwork, error) {

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	var network *govcd.OrgVDCNetwork
	if isDataSource {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_network_"+wanted)
		}
		filter, hasFilter := d.GetOk("filter")

		if hasFilter {
			network, err = getNetworkByFilter(vdc, filter, wanted)
			if err != nil {
				return nil, err
			}
			return network, nil
		}
	}

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, fmt.Errorf("[get network] no identifier found for network")
	}
	network, err = vdc.GetOrgVdcNetworkByNameOrId(identifier, false)
	if err != nil {
		return nil, fmt.Errorf("[get network] error getting network %s: %s", identifier, err)
	}

	return network, nil
}

// resourceVcdNetworkDirectImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_network_direct.my-network
// Example import path (_the_id_string_): org.vdc.my-network
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdNetworkDirectImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[direct network import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[direct network import] unable to find VDC %s: %s ", vdcName, err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[direct network import] error retrieving network %s: %s", networkName, err)
	}
	parentNetwork := network.OrgVDCNetwork.Configuration.ParentNetwork
	if parentNetwork == nil || parentNetwork.Name == "" {
		return nil, fmt.Errorf("[direct network import] no parent network found for %s", network.OrgVDCNetwork.Name)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "external_network", parentNetwork.Name)
	d.SetId(network.OrgVDCNetwork.ID)
	return []*schema.ResourceData{d}, nil
}
