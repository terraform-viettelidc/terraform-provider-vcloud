package vcloud

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNetworkRoutedV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNetworkRoutedV2Read,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				Deprecated:    "Deprecated in favor of `edge_gateway_id`. Routed networks will inherit VDC from parent Edge Gateway.",
				ConflictsWith: []string{"edge_gateway_id"},
			},
			"edge_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Edge gateway name in which Routed network is located",
				ConflictsWith: []string{"vdc"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "A unique name for this network (optional if 'filter' is used)",
			},
			"filter": {
				Type:         schema.TypeList,
				MaxItems:     1,
				MinItems:     1,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "Criteria for retrieving a network by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"ip":         elementIp,
					},
				},
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of VDC or VDC Group",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network description",
			},
			"interface_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Interface type. One of 'INTERNAL', 'DISTRIBUTED', 'SUBINTERFACE', 'NON_DISTRIBUTED'",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP address",
			},
			"prefix_length": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Network prefix",
			},
			"dual_stack_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value if Dual-Stack mode is enabled",
			},
			"secondary_gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary gateway (can only be IPv6 and requires enabled Dual Stack mode)",
			},
			"secondary_prefix_length": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secondary prefix (can only be IPv6 and requires enabled Dual Stack mode)",
			},
			"secondary_static_ip_pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Secondary IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRangeComputed,
			},
			"dns1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns2": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix",
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRangeComputed,
			},
			"guest_vlan_allowed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if network allows guest VLAN tagging",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata assigned to this network. Key and value can be any string",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("Network"),
			"route_advertisement_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this network is advertised so that it can be routed out to the external networks.",
			},
		},
	}
}

var networkV2IpRangeComputed = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start address of the IP range",
		},
		"end_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "End address of the IP range",
		},
	},
}

func datasourceVcdNetworkRoutedV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network read v2] error retrieving Org: %s", err)
	}

	networkName := d.Get("name").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	if !nameOrFilterIsSet(d) {
		return diag.Errorf(noNameOrFilterError, "vcd_network_routed_v2")
	}

	// Try to search by filter if it exists
	var network *govcd.OpenApiOrgVdcNetwork
	filter, hasFilter := d.GetOk("filter")

	switch {
	// User supplied `filter` and also `edge_gateway_id` is present, search in the `vdc` (in data
	// source or inherited)
	case hasFilter && networkName == "" && edgeGatewayId != "":
		// Lookup Edge Gateway to know parent VDC or VDC Group (routed networks always exists in the
		// same VDC/VDC Group as Edge Gateway)
		anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(edgeGatewayId)
		if err != nil {
			return diag.Errorf("error retrieving Edge Gateway structure: %s", err)
		}
		parentVdcOrVdcGroupId := anyEdgeGateway.EdgeGateway.OwnerRef.ID

		if govcd.OwnerIsVdcGroup(parentVdcOrVdcGroupId) {
			return diag.Errorf("filters do not support VDC Groups yet")
		}

		vdc, err := org.GetVDCById(parentVdcOrVdcGroupId, false)
		if err != nil {
			return diag.Errorf("error retrieving parent VDC: %s", err)
		}

		network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "routed")
		if err != nil {
			return diag.FromErr(err)
		}
	// User supplied `filter` but `edge_gateway_id` is not present, therefore VDC must be looked up
	// by legacy means
	case hasFilter && networkName == "" && edgeGatewayId == "":
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf("error retrieving VDC: %s", err)
		}

		network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "routed")
		if err != nil {
			return diag.FromErr(err)
		}
	// TODO - XML Query based API does not support VDC Group networks (does not return them)
	// User supplied `filter` and `edge_gateway_id` (search scope can be detected - VDC or VDC Group)
	// case hasFilter && edgeGatewayId != "":
	// 	network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "routed")
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// User supplied `name` and also `edge_gateway_id`
	case edgeGatewayId != "" && networkName != "":
		// Lookup Edge Gateway to know parent VDC or VDC Group (routed networks always exists in the
		// same VDC/VDC Group as Edge Gateway)
		anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(edgeGatewayId)
		if err != nil {
			return diag.Errorf("error retrieving Edge Gateway structure: %s", err)
		}
		parentVdcOrVdcGroupId := anyEdgeGateway.EdgeGateway.OwnerRef.ID

		network, err = org.GetOpenApiOrgVdcNetworkByNameAndOwnerId(networkName, parentVdcOrVdcGroupId)
		if err != nil {
			return diag.Errorf("[routed network read v2] error getting Org VDC network: %s", err)
		}
	// Users supplied only `name` (VDC reference will be used from resource or inherited from provider)
	case networkName != "":
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf("error getting VDC: %s", err)
		}

		network, err = vdc.GetOpenApiOrgVdcNetworkByName(d.Get("name").(string))
		if err != nil {
			return diag.Errorf("[routed network read v2] error getting Org VDC network: %s", err)
		}
	default:
		return diag.Errorf("error - not all parameters specified for network lookup")
	}

	if !network.IsRouted() {
		return diag.Errorf("[routed network read v2] Org network with name '%s' found, but is not of type Routed (NAT_ROUTED) (type is '%s')",
			network.OpenApiOrgVdcNetwork.Name, network.GetType())
	}

	err = setOpenApiOrgVdcRoutedNetworkData(d, network.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[routed network read v2] error setting Org VDC network data: %s", err)
	}

	d.SetId(network.OpenApiOrgVdcNetwork.ID)

	// Metadata is not supported when the network is in a VDC Group
	if !govcd.OwnerIsVdcGroup(network.OpenApiOrgVdcNetwork.OwnerRef.ID) {
		diags = append(diags, updateMetadataInStateDeprecated(d, vcdClient, "vcd_network_routed_v2", network)...)
		if diags != nil && diags.HasError() {
			return diags
		}
	}

	// This must be checked at the end as updateMetadataInStateDeprecated can throw Warning diagnostics
	if len(diags) > 0 {
		return diags
	}

	return nil
}
