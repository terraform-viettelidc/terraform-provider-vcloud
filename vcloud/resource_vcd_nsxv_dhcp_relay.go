package vcloud

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var relayAgentResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"network_name": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "Org network which is to be used for relaying DHCP message to specified servers",
		},
		"gateway_ip_address": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeString,
			Description: "Optional gateway IP address of org network which is to be used for relaying DHCP message to specified servers",
		},
	},
}

func resourceVcdNsxvDhcpRelay() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvDhcpRelayCreate,
		Read:   resourceVcdNsxvDhcpRelayRead,
		Update: resourceVcdNsxvDhcpRelayUpdate,
		Delete: resourceVcdNsxvDhcpRelayDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvDhcpRelayImport,
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
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name for DHCP relay settings",
			},
			"ip_addresses": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP address of DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain_names": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP domain names of DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_sets": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP set names which consist DHCP servers",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"relay_agent": {
				Required: true,
				MinItems: 1,
				Type:     schema.TypeSet,
				Elem:     relayAgentResource,
			},
		},
	}
}

// resourceVcdNsxvDhcpRelayCreate sets up DHCP relay configuration as per supplied schema
// configuration
func resourceVcdNsxvDhcpRelayCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	dhcpRelayConfig, err := getDhcpRelayType(d, edgeGateway, vcdClient)
	if err != nil {
		return fmt.Errorf("could not process DHCP relay settings: %s", err)
	}

	_, err = edgeGateway.UpdateDhcpRelay(dhcpRelayConfig)
	if err != nil {
		return fmt.Errorf("unable to update DHCP relay settings for Edge Gateway %s: %s", edgeGateway.EdgeGateway.Name, err)
	}

	// This is not a real object but a settings property on Edge gateway - creating a fake composite
	// ID
	compositeId, err := getDhcpRelaySettingsId(edgeGateway)
	if err != nil {
		return fmt.Errorf("could not construct DHCP relay settings ID: %s", err)
	}

	d.SetId(compositeId)

	return resourceVcdNsxvDhcpRelayRead(d, meta)
}

// resourceVcdNsxvDhcpRelayUpdate is in fact exactly the same as create because there is no object,
// just settings to modify
func resourceVcdNsxvDhcpRelayUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdNsxvDhcpRelayCreate(d, meta)
}

func resourceVcdNsxvDhcpRelayRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdNsxvDhcpRelayRead(d, meta, "resource")
}

// genericVcdNsxvDhcpRelayRead reads DHCP relay configuration and persists to statefile
func genericVcdNsxvDhcpRelayRead(d *schema.ResourceData, meta interface{}, origin string) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	dhcpRelaySettings, err := edgeGateway.GetDhcpRelay()
	if err != nil {
		return fmt.Errorf("could not read DHCP relay settings: %s", err)
	}

	err = setDhcpRelayServerData(d, dhcpRelaySettings, edgeGateway, vdc)
	if err != nil {
		return fmt.Errorf("could not set DHCP relay server settings: %s", err)
	}

	err = setDhcpRelayAgentData(d, dhcpRelaySettings, edgeGateway, vdc)
	if err != nil {
		return fmt.Errorf("could not set DHCP relay agent settings: %s", err)
	}

	// This is not a real object but a settings property on Edge gateway - creating a fake composite
	// ID
	compositeId, err := getDhcpRelaySettingsId(edgeGateway)
	if err != nil {
		return fmt.Errorf("could not construct DHCP relay settings ID: %s", err)
	}

	d.SetId(compositeId)

	return nil
}

// resourceVcdNsxvDhcpRelayDelete removes DHCP relay configuration by triggering ResetDhcpRelay()
func resourceVcdNsxvDhcpRelayDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.ResetDhcpRelay()
	if err != nil {
		return fmt.Errorf("could not reset DHCP relay settings: %s", err)
	}

	return nil
}

// resourceVcdNsxvDhcpRelayImport imports DHCP relay configuration. Because DHCP relay is just a
// settings on edge gateway and not a separate object - the ID actually does not represent any
// object
func resourceVcdNsxvDhcpRelayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified in such way org-name.vdc-name.edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	compositeId, err := getDhcpRelaySettingsId(edgeGateway)
	if err != nil {
		return nil, fmt.Errorf("could not construct DHCP relay settings ID: %s", err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway", edgeName)
	d.SetId(compositeId)
	return []*schema.ResourceData{d}, nil
}

// getDhcpRelayType converts resource schema to *types.EdgeDhcpRelay
func getDhcpRelayType(d *schema.ResourceData, edge *govcd.EdgeGateway, vcdClient *VCDClient) (*types.EdgeDhcpRelay, error) {
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	dhcpRelayConfig := &types.EdgeDhcpRelay{}

	// Relay server part
	var (
		listOfIps         []string
		listOfDomainNames []string
		listOfIpSetNames  []string
		listOfIpSetIds    []string
	)

	if ipAddresses, ok := d.GetOk("ip_addresses"); ok {
		listOfIps = convertSchemaSetToSliceOfStrings(ipAddresses.(*schema.Set))
	}

	if domainNames, ok := d.GetOk("domain_names"); ok {
		listOfDomainNames = convertSchemaSetToSliceOfStrings(domainNames.(*schema.Set))
	}

	if ipSetNames, ok := d.GetOk("ip_sets"); ok {
		listOfIpSetNames = convertSchemaSetToSliceOfStrings(ipSetNames.(*schema.Set))
		listOfIpSetIds, err = ipSetNamesToIds(listOfIpSetNames, vdc, false)
		if err != nil {
			return nil, fmt.Errorf("could not lookup supplied IP set IDs by their names: %s", err)
		}
	}

	dhcpRelayServer := &types.EdgeDhcpRelayServer{
		IpAddress:        listOfIps,
		Fqdns:            listOfDomainNames,
		GroupingObjectId: listOfIpSetIds,
	}

	// Add DHCP relay server part to struct
	dhcpRelayConfig.RelayServer = dhcpRelayServer

	// Relay agent part
	relayAgent := d.Get("relay_agent")
	relayAgentsStruct, err := getDhcpRelayAgentsType(relayAgent.(*schema.Set), edge)
	if err != nil {
		return nil, fmt.Errorf("could not process relay agents: %s", err)
	}
	// Add all relay agent values to struct
	dhcpRelayConfig.RelayAgents = &types.EdgeDhcpRelayAgents{Agents: relayAgentsStruct}

	return dhcpRelayConfig, nil

}

// getDhcpRelayAgentsType converts relay_agent configuration blocks to []types.EdgeDhcpRelayAgent
func getDhcpRelayAgentsType(relayAgentsSet *schema.Set, edge *govcd.EdgeGateway) ([]types.EdgeDhcpRelayAgent, error) {
	relayAgentsSlice := relayAgentsSet.List()
	relayAgentsStruct := make([]types.EdgeDhcpRelayAgent, len(relayAgentsSlice))
	for index, relayAgent := range relayAgentsSlice {
		relayAgentMap := convertToStringMap(relayAgent.(map[string]interface{}))

		// Lookup vNic index by network name
		orgNetworkName := relayAgentMap["network_name"]
		vNicIndex, _, err := edge.GetAnyVnicIndexByNetworkName(orgNetworkName)
		if err != nil {
			return nil, fmt.Errorf("could not lookup edge gateway interface (vNic) index by network name for network %s: %s", orgNetworkName, err)
		}

		oneRelayAgent := types.EdgeDhcpRelayAgent{
			VnicIndex: vNicIndex,
		}

		if gatewayIp, isSet := relayAgentMap["gateway_ip_address"]; isSet {
			oneRelayAgent.GatewayInterfaceAddress = gatewayIp
		}

		relayAgentsStruct[index] = oneRelayAgent
	}

	return relayAgentsStruct, nil
}

// setDhcpRelayServerData sets DHCP relay server related fields into statefile
func setDhcpRelayServerData(d *schema.ResourceData, edgeRelay *types.EdgeDhcpRelay, edge *govcd.EdgeGateway, vdc *govcd.Vdc) error {
	relayServer := edgeRelay.RelayServer
	// If relay server has no config - just return it empty
	if relayServer == nil {
		return nil
	}

	relayServerIpAddressesSet := convertStringsToTypeSet(relayServer.IpAddress)
	err := d.Set("ip_addresses", relayServerIpAddressesSet)
	if err != nil {
		return fmt.Errorf("could not save ip_addresses to schema: %s", err)
	}

	relayServerDomainNamesSet := convertStringsToTypeSet(relayServer.Fqdns)
	err = d.Set("domain_names", relayServerDomainNamesSet)
	if err != nil {
		return fmt.Errorf("could not save domain_names to schema: %s", err)
	}
	ipSetNames, err := ipSetIdsToNames(relayServer.GroupingObjectId, vdc)
	if err != nil {
		return fmt.Errorf("could not find names for all IP set IDs: %s", err)
	}

	relayServerIpSetNamesSet := convertStringsToTypeSet(ipSetNames)
	err = d.Set("ip_sets", relayServerIpSetNamesSet)
	if err != nil {
		return fmt.Errorf("could not save ip_sets to schema: %s", err)
	}

	return nil
}

// setDhcpRelayAgentData sets DHCP relay agent related fields into statefile
func setDhcpRelayAgentData(d *schema.ResourceData, edgeRelay *types.EdgeDhcpRelay, edge *govcd.EdgeGateway, vdc *govcd.Vdc) error {
	relayAgents := edgeRelay.RelayAgents
	// If there are no relay agents - just return it empty
	if relayAgents == nil {
		return nil
	}

	relayAgentSlice := make([]interface{}, len(relayAgents.Agents))

	for index, agent := range relayAgents.Agents {
		relayAgentMap := make(map[string]interface{})
		if agent.VnicIndex == nil {
			return fmt.Errorf("DHCP relay agent configuration does not have vNic specified")
		}
		// Lookup org network name by edge gateway vNic index
		orgNetworkName, _, err := edge.GetNetworkNameAndTypeByVnicIndex(*agent.VnicIndex)
		if err != nil {
			return fmt.Errorf("could not find network name for edge gateway vNic %d: %s ", agent.VnicIndex, err)
		}

		relayAgentMap["network_name"] = orgNetworkName
		relayAgentMap["gateway_ip_address"] = agent.GatewayInterfaceAddress

		relayAgentSlice[index] = relayAgentMap
	}

	relayAgentSet := schema.NewSet(schema.HashResource(relayAgentResource), relayAgentSlice)
	err := d.Set("relay_agent", relayAgentSet)
	if err != nil {
		return fmt.Errorf("could not save relay_agent to schema: %s", err)
	}

	return nil
}

// getDhcpRelaySettingsId constructs a fake DHCP relay configuration ID which is needed for
// Terraform. The ID is in format "edgeGateway.ID:dhcpRelay"
// (eg.: "urn:vcloud:gateway:77ccbdcd-ac04-4111-bf08-8ac294a3185b:dhcpRelay"). Edge Gateway ID is
// left here just in case we ever want to refer this object somewhere but still be able to
// distinguish it from the real edge gateway resource.
func getDhcpRelaySettingsId(edge *govcd.EdgeGateway) (string, error) {
	if edge.EdgeGateway.ID == "" {
		return "", fmt.Errorf("edge gateway does not have ID populated")
	}

	id := edge.EdgeGateway.ID + ":dhcpRelay"
	return id, nil
}
