package vcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtEdgeGatewayDns() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdNsxtEdgegatewayDnsRead,
		CreateContext: resourceVcdNsxtEdgegatewayDnsCreate,
		UpdateContext: resourceVcdNsxtEdgegatewayDnsUpdate,
		DeleteContext: resourceVcdNsxtEdgegatewayDnsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgegatewayDnsImport,
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
				Description: "Edge gateway ID for DNS configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Status of the DNS Forwarder. Defaults to `true`",
			},
			"listener_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "IP on which the DNS forwarder listens." +
					"Can be modified only if the Edge Gateway has a dedicated external network.",
				ValidateFunc: validation.IsIPAddress,
			},
			"snat_rule_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The value is `true` if a SNAT rule exists for the DNS forwarder.",
			},
			"snat_rule_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "The external IP address of the SNAT rule. " +
					"Can be modified only if the Edge Gateway's external network is using IP spaces. (VCD 10.5.0+)",
			},
			"default_forwarder_zone": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The default forwarder zone.",
				Elem:        defaultForwarderZone,
			},
			"conditional_forwarder_zone": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    5,
				Description: "Conditional forwarder zones",
				Elem:        conditionalForwarderZone,
			},
		},
	}
}

var defaultForwarderZone = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    3,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	},
}

var conditionalForwarderZone = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Required:    true,
			MaxItems:    3,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
		"domain_names": {
			Type:        schema.TypeSet,
			Required:    true,
			Description: "Set of domain names on which conditional forwarding is based.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func resourceVcdNsxtEdgegatewayDnsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDnsCreateUpdate(ctx, d, meta, "create")
}

func resourceVcdNsxtEdgegatewayDnsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtEdgegatewayDnsCreateUpdate(ctx, d, meta, "update")
}

func resourceVcdNsxtEdgegatewayDnsCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[edge gateway dns %s] %s", origin, err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[edge gateway dns %s] error retrieving Edge Gateway: %s", origin, err)
	}

	dns, err := nsxtEdge.GetDnsConfig()
	if err != nil {
		return diag.Errorf("[edge gateway dns %s] error getting current DNS configuration: %s", origin, err)
	}

	dnsConfig, err := getNsxtEdgeGatewayDnsConfig(d, vcdClient)
	if err != nil {
		return diag.Errorf("[edge gateway dns %s] error getting DNS configuration from schema: %s", origin, err)
	}

	updatedDns, err := dns.Update(dnsConfig)
	if err != nil {
		if strings.Contains(err.Error(), "or the target entity is invalid") {
			if err2 := doesNotWorkWithDistributedOnlyEdgeGateway("vcd_nsxt_edgegateway_dns", vcdClient, nsxtEdge); err2 != nil {
				return diag.Errorf(err.Error() + "\n\n" + err2.Error())
			}
		}
		return diag.Errorf("[edge gateway dns %s] error updating DNS configuration: %s", origin, err)
	}

	d.SetId(updatedDns.EdgeGatewayId)

	return resourceVcdNsxtEdgegatewayDnsRead(ctx, d, meta)
}

func resourceVcdNsxtEdgegatewayDnsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxtEdgegatewayDnsRead(ctx, d, meta, "resource")
}

func genericVcdNsxtEdgegatewayDnsRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)

	var nsxtEdge *govcd.NsxtEdgeGateway
	var err error
	edgeGatewayId := d.Get("edge_gateway_id").(string)
	if d.Id() == "" && edgeGatewayId == "" {
		return diag.Errorf("id wasn't provided for Edge Gateway DNS")
	}
	if edgeGatewayId == "" {
		nsxtEdge, err = vcdClient.GetNsxtEdgeGatewayById(orgName, d.Id())
	} else {
		nsxtEdge, err = vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			// When parent Edge Gateway is not found - this resource is also not found and should be
			// removed from state
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	dns, err := nsxtEdge.GetDnsConfig()
	if err != nil {
		return diag.Errorf("[edge gateway dns read] error retrieving NSX-T Edge Gateway DNS config: %s", err)
	}
	if dns.NsxtEdgeGatewayDns.DefaultForwarderZone == nil && origin == "resource" {
		d.SetId("")
		return nil
	}
	d.SetId(dns.EdgeGatewayId)
	dSet(d, "edge_gateway_id", dns.EdgeGatewayId)

	err = setNsxtEdgeGatewayDnsConfig(d, dns.NsxtEdgeGatewayDns)
	if err != nil {
		return diag.Errorf("[edge gateway dns read] error storing state: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDnsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[edge gateway dns delete] %s", err)
	}
	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[edge gateway dns delete] error retrieving Edge Gateway: %s", err)
	}

	dns, err := nsxtEdge.GetDnsConfig()
	if err != nil {
		return diag.Errorf("[edge gateway dns delete] error retrieving DNS Configuration: %s", err)
	}

	err = dns.Delete()
	if err != nil {
		return diag.Errorf("[edge gateway dns delete] error deleting DNS Configuration: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgegatewayDnsImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway DNS import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name or org-name.vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("this resource is only supported on NSX-T backed Edge Gateways")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edge.EdgeGateway.ID)
	// Storing Edge Gateway ID and Read will retrieve all other data
	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtEdgeGatewayDnsConfig(d *schema.ResourceData, vcdClient *VCDClient) (*types.NsxtEdgeGatewayDns, error) {
	enabled := d.Get("enabled").(bool)
	listenerIp := d.Get("listener_ip").(string)

	// SNAT Rule IP address field was introduced in API version 38.0
	if _, ok := d.GetOk("snat_rule_ip_address"); ok {
		if vcdClient.Client.APIVCDMaxVersionIs("<38.0") {
			return nil, fmt.Errorf("snat rule ip address is supported on vcd version 10.5.0 and newer")
		}
	}
	snatRuleIp := d.Get("snat_rule_ip_address").(string)

	defaultUpstreamServersSet := d.Get("default_forwarder_zone.0.upstream_servers").(*schema.Set)
	defaultUpstreamServers := convertSchemaSetToSliceOfStrings(defaultUpstreamServersSet)
	defaultForwarderZone := &types.NsxtDnsForwarderZoneConfig{
		ID:              d.Get("default_forwarder_zone.0.id").(string),
		DisplayName:     d.Get("default_forwarder_zone.0.name").(string),
		UpstreamServers: defaultUpstreamServers,
	}

	// Currently, if the conditional forwarder zones are updated, we don't preserve the IDs due to the nature of TypeSets and
	// just re-create the zones as it doesn't impact performance much and the values can't have any dependencies that would
	// prevent re-creation.
	// In the future, we might need to implement partial matching by domain_names, so that the ID stays the same
	// if domain_names are updated.
	conditionalForwarderZoneSet := d.Get("conditional_forwarder_zone").(*schema.Set)
	conditionalForwarderZones := make([]*types.NsxtDnsForwarderZoneConfig, len(conditionalForwarderZoneSet.List()))
	for zoneIndex, zone := range conditionalForwarderZoneSet.List() {
		zoneDefinition := zone.(map[string]any)
		upstreamServersSet := zoneDefinition["upstream_servers"].(*schema.Set)
		upstreamServers := convertSchemaSetToSliceOfStrings(upstreamServersSet)
		domainNameSet := zoneDefinition["domain_names"].(*schema.Set)
		domainNames := convertSchemaSetToSliceOfStrings(domainNameSet)
		zone := &types.NsxtDnsForwarderZoneConfig{
			ID:              zoneDefinition["id"].(string),
			DisplayName:     zoneDefinition["name"].(string),
			UpstreamServers: upstreamServers,
			DnsDomainNames:  domainNames,
		}
		conditionalForwarderZones[zoneIndex] = zone
	}

	dnsConfig := &types.NsxtEdgeGatewayDns{
		Enabled:                   enabled,
		ListenerIp:                listenerIp,
		SnatRuleExternalIpAddress: snatRuleIp,
		DefaultForwarderZone:      defaultForwarderZone,
		ConditionalForwarderZones: conditionalForwarderZones,
	}

	return dnsConfig, nil
}

func setNsxtEdgeGatewayDnsConfig(d *schema.ResourceData, dnsConfig *types.NsxtEdgeGatewayDns) error {
	dSet(d, "enabled", dnsConfig.Enabled)
	dSet(d, "listener_ip", dnsConfig.ListenerIp)
	dSet(d, "snat_rule_enabled", dnsConfig.SnatRuleEnabled)
	dSet(d, "snat_rule_ip_address", dnsConfig.SnatRuleExternalIpAddress)

	if dnsConfig.DefaultForwarderZone != nil {
		defaultForwarderZoneBlock := make([]interface{}, 1)
		defaultForwarderZone := make(map[string]interface{})
		defaultForwarderZone["id"] = dnsConfig.DefaultForwarderZone.ID
		defaultForwarderZone["name"] = dnsConfig.DefaultForwarderZone.DisplayName
		defaultForwarderZone["upstream_servers"] = convertStringsToTypeSet(dnsConfig.DefaultForwarderZone.UpstreamServers)
		defaultForwarderZoneBlock[0] = defaultForwarderZone
		err := d.Set("default_forwarder_zone", defaultForwarderZoneBlock)
		if err != nil {
			return fmt.Errorf("error storing 'default_forwarder_zone' into state: %s", err)
		}
	}

	if len(dnsConfig.ConditionalForwarderZones) != 0 {
		conditionalForwarderZoneInterface := make([]interface{}, len(dnsConfig.ConditionalForwarderZones))
		for index, zone := range dnsConfig.ConditionalForwarderZones {
			singleZone := make(map[string]interface{})
			singleZone["id"] = zone.ID
			singleZone["name"] = zone.DisplayName
			singleZone["domain_names"] = convertStringsToTypeSet(zone.DnsDomainNames)
			singleZone["upstream_servers"] = convertStringsToTypeSet(zone.UpstreamServers)

			conditionalForwarderZoneInterface[index] = singleZone
		}

		err := d.Set("conditional_forwarder_zone", conditionalForwarderZoneInterface)
		if err != nil {
			return fmt.Errorf("error storing 'conditional_forwarder_zone' into state: %s", err)
		}
	}

	return nil
}
