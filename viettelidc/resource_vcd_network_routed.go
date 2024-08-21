package viettelidc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdNetworkRouted() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNetworkRoutedCreate,
		ReadContext:   resourceVcdNetworkRoutedRead,
		DeleteContext: resourceVcdNetworkDeleteLocked,
		UpdateContext: resourceVcdNetworkRoutedUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNetworkRoutedImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique name for the network",
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

			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the edge gateway",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description for the network",
			},

			"interface_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "internal",
				ForceNew:     true,
				Description:  "Which interface to use (one of `internal`, `subinterface`, `distributed`)",
				ValidateFunc: validation.StringInSlice([]string{"internal", "subinterface", "distributed"}, true),
				// Diff suppress function used to ease upgrade operations from versions where the interface was implicit
				DiffSuppressFunc: suppressNetworkUpgradedInterface(),
			},

			"netmask": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "255.255.255.0",
				Description:  "The netmask for the new network",
				ValidateFunc: validation.IsIPAddress,
			},

			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Description:  "The gateway of this network",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns1": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "First DNS server to use",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns2": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Second DNS server to use",
				ValidateFunc: validation.IsIPAddress,
			},

			"dns_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A FQDN for the virtual machines on this network",
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

			"dhcp_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"end_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"default_lease_time": {
							Type:        schema.TypeInt,
							Computed:    true, // vCD doesn't process this field as input. It sets the value to max_lease_time
							Description: "The default DHCP lease time to use",
						},

						"max_lease_time": {
							Type:        schema.TypeInt,
							Default:     7200,
							Optional:    true,
							Description: "The maximum DHCP lease time to use",
						},
					},
				},
				Set: resourceVcdNetworkRoutedDhcpPoolHash,
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The first address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},

						"end_address": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The final address in the IP Range",
							ValidateFunc: validation.IsIPAddress,
						},
					},
				},
				Set: resourceVcdNetworkStaticIpPoolHash,
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

func resourceVcdNetworkRoutedCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	if vdc.IsNsxt() {
		logForScreen("vcd_network_routed", "WARNING: please use 'vcd_network_routed_v2' for NSX-T VDCs")
	}

	edgeGatewayName := d.Get("edge_gateway").(string)
	edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	gatewayName := d.Get("gateway").(string)
	networkName := d.Get("name").(string)
	netMask := d.Get("netmask").(string)
	dns1 := d.Get("dns1").(string)
	dns2 := d.Get("dns2").(string)

	ipRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	orgVDCNetwork := &types.OrgVDCNetwork{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        networkName,
		Description: d.Get("description").(string),

		EdgeGateway: &types.Reference{
			HREF: edgeGateway.EdgeGateway.HREF,
		},
		Configuration: &types.NetworkConfiguration{
			FenceMode: "natRouted",
			IPScopes: &types.IPScopes{
				IPScope: []*types.IPScope{{
					IsInherited: false,
					Gateway:     gatewayName,
					Netmask:     netMask,
					DNS1:        dns1,
					DNS2:        dns2,
					DNSSuffix:   d.Get("dns_suffix").(string),
					IPRanges:    &ipRanges,
				}},
			},
			BackwardCompatibilityMode: true,
		},
		IsShared: d.Get("shared").(bool),
	}
	distributedAllowed := false
	if edgeGateway.EdgeGateway.Configuration.DistributedRoutingEnabled != nil {
		if *edgeGateway.EdgeGateway.Configuration.DistributedRoutingEnabled {
			distributedAllowed = true
		}
	}
	networkInterface := d.Get("interface_type").(string)
	trueValue := true
	switch strings.ToLower(networkInterface) {
	case "internal":
		// default: no configuration is needed
		orgVDCNetwork.Configuration.SubInterface = nil
	case "subinterface":
		orgVDCNetwork.Configuration.SubInterface = &trueValue
	case "distributed":
		if distributedAllowed {
			orgVDCNetwork.Configuration.DistributedInterface = &trueValue
		} else {
			return diag.Errorf("interface 'distributed' requested, but distributed routing is not enabled in edge gateway '%s'", edgeGateway.EdgeGateway.Name)
		}
	}

	err = vdc.CreateOrgVDCNetworkWait(orgVDCNetwork)
	if err != nil {
		return diag.Errorf("error: %s", err)
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, true)
	if err != nil {
		return diag.Errorf("error finding network: %s", err)
	}

	if dhcp, ok := d.GetOk("dhcp_pool"); ok {
		task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.(*schema.Set).List())
		if err != nil {
			return diag.Errorf("error adding DHCP pool: %s", err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf(errorCompletingTask, err)
		}
	}

	d.SetId(network.OrgVDCNetwork.ID)

	err = createOrUpdateMetadata(d, network, "metadata")
	if err != nil {
		return diag.Errorf("error adding metadata to routed network: %s", err)
	}

	return resourceVcdNetworkRoutedRead(ctx, d, meta)
}

func resourceVcdNetworkRoutedRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNetworkRoutedRead(ctx, d, meta, "resource")
}

func genericVcdNetworkRoutedRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	var diags diag.Diagnostics
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := getNetwork(d, vcdClient, origin == "datasource", "routed")
	if err != nil {
		if origin == "resource" {
			networkName := d.Get("name").(string)
			log.Printf("[DEBUG] Network %s no longer exists. Removing from tfstate", networkName)
			d.SetId("")
			return nil
		}
		return diag.Errorf("[routed network read] error retrieving Org VDC network  %s", err)
	}
	edgeGatewayName := d.Get("edge_gateway").(string)

	// When this function is called from the data source
	if edgeGatewayName == "" {
		edgeGatewayName, err = vdc.FindEdgeGatewayNameByNetwork(network.OrgVDCNetwork.Name)
		if err != nil {
			return diag.Errorf("[routed network read] no edge gateway connection found for network %s: %s", network.OrgVDCNetwork.Name, err)
		}
		dSet(d, "edge_gateway", edgeGatewayName)
	}
	edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
	if err != nil {
		log.Printf("[DEBUG] error retrieving edge gateway")
		return diag.Errorf("[routed network read] error retrieving edge gateway %s: %s", edgeGatewayName, err)
	}

	dSet(d, "name", network.OrgVDCNetwork.Name)
	dSet(d, "href", network.OrgVDCNetwork.HREF)
	dSet(d, "shared", network.OrgVDCNetwork.IsShared)
	if c := network.OrgVDCNetwork.Configuration; c != nil {
		if c.IPScopes != nil {
			dSet(d, "gateway", c.IPScopes.IPScope[0].Gateway)
			dSet(d, "netmask", c.IPScopes.IPScope[0].Netmask)
			dSet(d, "dns1", c.IPScopes.IPScope[0].DNS1)
			dSet(d, "dns2", c.IPScopes.IPScope[0].DNS2)
			dSet(d, "dns_suffix", c.IPScopes.IPScope[0].DNSSuffix)
		}
	}

	dhcp := getDhcpFromEdgeGateway(network.OrgVDCNetwork.HREF, edgeGateway)
	if len(dhcp) > 0 {
		newSet := &schema.Set{
			F: resourceVcdNetworkRoutedDhcpPoolHash,
		}
		for _, element := range dhcp {
			newSet.Add(element)
		}
		err := d.Set("dhcp_pool", newSet)
		if err != nil {
			return diag.Errorf("[routed network read] dhcp set: %s", err)
		}
	}

	staticIpPool := getStaticIpPool(network)
	if len(staticIpPool) > 0 {
		newSet := &schema.Set{
			F: resourceVcdNetworkStaticIpPoolHash,
		}
		for _, element := range staticIpPool {
			newSet.Add(element)
		}
		err := d.Set("static_ip_pool", newSet)
		if err != nil {
			return diag.Errorf("[routed network read] static_ip set: %s", err)
		}
	}

	if network.OrgVDCNetwork.Configuration.SubInterface == nil {
		dSet(d, "interface_type", "internal")
	} else {
		if *network.OrgVDCNetwork.Configuration.SubInterface {
			dSet(d, "interface_type", "subinterface")
		} else {
			if *network.OrgVDCNetwork.Configuration.DistributedInterface {
				dSet(d, "interface_type", "distributed")
			}
		}
	}
	dSet(d, "description", network.OrgVDCNetwork.Description)
	d.SetId(network.OrgVDCNetwork.ID)

	diags = append(diags, updateMetadataInStateDeprecated(d, vcdClient, "vcd_network_routed", network)...)
	if diags != nil && diags.HasError() {
		log.Printf("[DEBUG] Unable to set routed network metadata: %v", diags)
		return diags
	}

	// This must be checked at the end as updateMetadataInStateDeprecated can throw Warning diagnostics
	if len(diags) > 0 {
		return diags
	}
	return nil
}

func getStaticIpPool(network *govcd.OrgVDCNetwork) []map[string]interface{} {
	var staticIpPool []map[string]interface{}
	if network.OrgVDCNetwork.Configuration.IPScopes == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope) == 0 ||
		network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges == nil ||
		len(network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges.IPRange) == 0 {
		return staticIpPool
	}
	for _, sip := range network.OrgVDCNetwork.Configuration.IPScopes.IPScope {
		if sip.IsEnabled {
			for _, iprange := range sip.IPRanges.IPRange {
				staticIp := map[string]interface{}{
					"start_address": iprange.StartAddress,
					"end_address":   iprange.EndAddress,
				}
				staticIpPool = append(staticIpPool, staticIp)
			}
		}
	}

	return staticIpPool
}

// getDhcpFromEdgeGateway examines the edge gateway for a DHCP service
// that is registered to the given network HREF.
// Returns an array of string maps suitable to be passed to d.Set("dhcp_pool", value)
func getDhcpFromEdgeGateway(networkHref string, edgeGateway *govcd.EdgeGateway) []map[string]interface{} {

	var dhcpConfig []map[string]interface{}

	gwConf := edgeGateway.EdgeGateway.Configuration

	if gwConf == nil ||
		gwConf.EdgeGatewayServiceConfiguration == nil ||
		gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService == nil ||
		len(gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool) == 0 {
		return dhcpConfig
	}
	for _, dhcp := range gwConf.EdgeGatewayServiceConfiguration.GatewayDhcpService.Pool {
		// This check should avoid a crash when the network is not referenced. See Issue #500
		if dhcp.Network == nil {
			continue
		}
		if haveSameUuid(dhcp.Network.HREF, networkHref) {
			dhcpRec := map[string]interface{}{
				"start_address":      dhcp.LowIPAddress,
				"end_address":        dhcp.HighIPAddress,
				"max_lease_time":     dhcp.MaxLeaseTime,
				"default_lease_time": dhcp.DefaultLeaseTime,
			}
			dhcpConfig = append(dhcpConfig, dhcpRec)
		}
	}
	return dhcpConfig
}

func resourceVcdNetworkDeleteLocked(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	return resourceVcdNetworkDelete(ctx, d, meta)
}

func resourceVcdNetworkDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	network, err := vdc.GetOrgVdcNetworkByNameOrId(d.Id(), false)
	if err != nil {
		return diag.Errorf("[routed network delete] error retrieving Org VDC network: %s", err)
	}

	task, err := network.Delete()
	if err != nil {
		return diag.Errorf("error deleting network: %s", err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceVcdNetworkStaticIpPoolHash computes a hash for a Static IP pool
func resourceVcdNetworkStaticIpPoolHash(v interface{}) int {
	// Handle this function with care.
	// Changing the hash algorithm will trigger a plan update.
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	_, err := buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["start_address"].(string))))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	_, err = buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["end_address"].(string))))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	return hashcodeString(buf.String())
}

func genericResourceVcdNetworkDhcpPoolHash(v interface{}, networkType string) int {
	// Handle this function with care.
	// Changing the hash algorithm will trigger a plan update.
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	_, err := buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["start_address"].(string))))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	_, err = buf.WriteString(fmt.Sprintf("%s-",
		strings.ToLower(m["end_address"].(string))))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	_, err = buf.WriteString(fmt.Sprintf("%d-",
		m["max_lease_time"].(int)))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}

	switch networkType {
	case "isolated":
		_, err = buf.WriteString(fmt.Sprintf("%d-", m["default_lease_time"].(int)))
		if err != nil {
			util.Logger.Printf("[ERROR] error writing to string: %s", err)
		}
	case "routed":
		// do nothing
	default:
		panic(fmt.Sprintf("network type %s not supported", networkType))
	}
	return hashcodeString(buf.String())
}

// resourceVcdNetworkRoutedDhcpPoolHash computes a hash for a DHCP pool
func resourceVcdNetworkRoutedDhcpPoolHash(v interface{}) int {
	return genericResourceVcdNetworkDhcpPoolHash(v, "routed")
}

// resourceVcdNetworkRoutedImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_network_routed.my-network
// Example import path (_the_id_string_): org.vdc.my-network
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdNetworkRoutedImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[routed network import] resource name must be specified as org-name.vdc-name.network-name")
	}
	orgName, vdcName, networkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[routed network import] unable to find VDC %s: %s ", vdcName, err)
	}

	if vdc.IsNsxt() {
		return nil, fmt.Errorf("[routed network import] please use 'vcd_network_routed_v2' for NSX-T VDCs")
	}

	network, err := vdc.GetOrgVdcNetworkByName(networkName, false)
	if err != nil {
		return nil, fmt.Errorf("[routed network import] error retrieving network %s: %s", networkName, err)
	}

	edgeGatewayName, err := vdc.FindEdgeGatewayNameByNetwork(networkName)
	if err != nil {
		return nil, fmt.Errorf("[routed network] no edge gateway connection found for network %s: %s", network.OrgVDCNetwork.Name, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway", edgeGatewayName)
	d.SetId(network.OrgVDCNetwork.ID)

	return []*schema.ResourceData{d}, nil
}

func resourceVcdNetworkRoutedUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	networkName := d.Get("name").(string)
	description := d.Get("description").(string)
	dns1 := d.Get("dns1").(string)
	dns2 := d.Get("dns2").(string)
	dnsSuffix := d.Get("dns_suffix").(string)
	isShared := d.Get("shared").(bool)
	networkInterface := d.Get("interface_type").(string)

	identifier := d.Id()
	if identifier == "" {
		identifier = networkName
	}
	network, err := getNetwork(d, vcdClient, false, "routed")
	if err != nil {
		return diag.Errorf("[routed network update] error getting network %s: %s", identifier, err)
	}
	network.OrgVDCNetwork.Name = networkName
	network.OrgVDCNetwork.Description = description
	network.OrgVDCNetwork.IsShared = d.Get("shared").(bool)
	ipRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	trueValue := true
	switch strings.ToLower(networkInterface) {
	case "internal":
		// default: no configuration is needed
		network.OrgVDCNetwork.Configuration.SubInterface = nil
	case "subinterface":
		network.OrgVDCNetwork.Configuration.SubInterface = &trueValue
	case "distributed":
		network.OrgVDCNetwork.Configuration.DistributedInterface = &trueValue
	}
	network.OrgVDCNetwork.IsShared = isShared
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS1 = dns1
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNS2 = dns2
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].DNSSuffix = dnsSuffix
	network.OrgVDCNetwork.Configuration.IPScopes.IPScope[0].IPRanges = &ipRanges

	err = network.Update()
	if err != nil {
		return diag.Errorf("[routed network update] error updating network %s: %s", network.OrgVDCNetwork.Name, err)
	}
	if d.HasChange("dhcp_pool") {
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf(errorRetrievingOrgAndVdc, err)
		}
		edgeGatewayName := d.Get("edge_gateway").(string)
		edgeGateway, err := vdc.GetEdgeGatewayByName(edgeGatewayName, false)
		if err != nil {
			return diag.Errorf(errorUnableToFindEdgeGateway, err)
		}
		if dhcp, ok := d.GetOk("dhcp_pool"); ok {
			task, err := edgeGateway.AddDhcpPool(network.OrgVDCNetwork, dhcp.(*schema.Set).List())
			if err != nil {
				return diag.Errorf("error updating DHCP pool: %s", err)
			}

			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.Errorf(errorCompletingTask, err)
			}
		}
	}

	err = createOrUpdateMetadata(d, network, "metadata")
	if err != nil {
		return diag.Errorf("[routed network update] error updating network metadata: %s", err)
	}

	return resourceVcdNetworkRoutedRead(ctx, d, meta)
}
