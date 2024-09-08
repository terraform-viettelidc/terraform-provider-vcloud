package vcloud

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtIpSecVpnTunnel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtIpSecVpnTunnelCreate,
		ReadContext:   resourceVcdNsxtIpSecVpnTunnelRead,
		UpdateContext: resourceVcdNsxtIpSecVpnTunnelUpdate,
		DeleteContext: resourceVcdNsxtIpSecVpnTunnelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtIpSecVpnTunnelImport,
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
				Description: "Edge gateway name in which IP Sec VPN configuration is located",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Enables or disables this configuration (default true)",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of IP Sec VPN Tunnel",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description IP Sec VPN Tunnel",
			},
			"pre_shared_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Pre-Shared Key (PSK)",
			},
			"authentication_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "PSK",
				Description: "One of 'PSK' (default), 'CERTIFICATE'",
			},
			"certificate_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Optional certificate ID to use for authentication",
				RequiredWith: []string{"ca_certificate_id"},
			},
			"ca_certificate_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Optional CA certificate ID to use for authentication",
				RequiredWith: []string{"certificate_id"},
			},
			"local_ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IPv4 Address for the endpoint. This has to be a sub-allocated IP on the Edge Gateway.",
			},
			"local_networks": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "Set of local networks in CIDR format. At least one value is required",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"remote_ip_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Public IPv4 Address of the remote device terminating the VPN connection",
			},
			"remote_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Custom remote ID of the peer site. 'remote_ip_address' is used by default",
			},
			"remote_networks": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Set of remote networks in CIDR format. Leaving it empty is interpreted as 0.0.0.0/0",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Sets whether logging for the tunnel is enabled or not. (default - false)",
			},
			"security_profile_customization": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Security profile customization",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ike_version": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IKE version one of IKE_V1, IKE_V2, IKE_FLEX",
							ValidateFunc: validation.StringInSlice([]string{"IKE_V1", "IKE_V2", "IKE_FLEX"}, false),
						},
						"ike_encryption_algorithms": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "Encryption algorithms. One of SHA1, SHA2_256, SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_digest_algorithms": {
							Type:     schema.TypeSet,
							Optional: true,
							Description: "Secure hashing algorithms to use during the IKE negotiation. One of SHA1, " +
								"SHA2_256, SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_dh_groups": {
							Type:     schema.TypeSet,
							Required: true,
							Description: "Diffie-Hellman groups to be used if Perfect Forward Secrecy is enabled. One " +
								"of GROUP2, GROUP5, GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_sa_lifetime": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Security Association life time (in seconds). It is number of seconds " +
								"before the IPsec tunnel needs to reestablish",
						},

						"tunnel_pfs_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Perfect Forward Secrecy Enabled or Disabled. Default (enabled)",
						},

						"tunnel_df_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "COPY",
							Description:  "Policy for handling defragmentation bit. One of COPY, CLEAR",
							ValidateFunc: validation.StringInSlice([]string{"COPY", "CLEAR"}, false),
						},

						"tunnel_encryption_algorithms": {
							Type:     schema.TypeSet,
							Required: true,
							Description: "Encryption algorithms to use in IPSec tunnel establishment. One of AES_128, " +
								"AES_256, AES_GCM_128, AES_GCM_192, AES_GCM_256, NO_ENCRYPTION_AUTH_AES_GMAC_128, " +
								"NO_ENCRYPTION_AUTH_AES_GMAC_192, NO_ENCRYPTION_AUTH_AES_GMAC_256, NO_ENCRYPTION",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_digest_algorithms": {
							Type:     schema.TypeSet,
							Optional: true,
							Description: "Digest algorithms to be used for message digest. One of SHA1, SHA2_256, " +
								"SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_dh_groups": {
							Type:     schema.TypeSet,
							Required: true,
							Description: "Diffie-Hellman groups to be used is PFS is enabled. One of GROUP2, GROUP5, " +
								"GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_sa_lifetime": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Security Association life time (in seconds)",
						},
						"dpd_probe_internal": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Value in seconds of dead probe detection interval. Minimum is 3 seconds and " +
								"the maximum is 60 seconds",
						},
					},
				},
			},
			// Computed attributes from here
			"security_profile": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "Security type which is use for IPsec VPN Tunnel. It will be 'DEFAULT' if nothing is " +
					"customized and 'CUSTOM' if some changes are applied",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Overall IPsec VPN Tunnel Status",
			},
			"ike_service_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status for the actual IKE Session for the given tunnel",
			},
			"ike_fail_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Provides more details of failure if the IKE service is not UP",
			},
		},
	}
}

func resourceVcdNsxtIpSecVpnTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel create] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel create] error retrieving Edge Gateway: %s", err)
	}

	ipSecVpnConfig, err := getNsxtIpSecVpnTunnelType(d)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel create] error getting NSX-T IPsec VPN Tunnel configuration type: %s", err)
	}

	createdIpSecVpnConfig, err := nsxtEdge.CreateIpSecVpnTunnel(ipSecVpnConfig)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel create] error creating NSX-T IPsec VPN Tunnel configuration: %s", err)
	}
	// IPSec VPN Tunnel is already created - storing ID
	d.SetId(createdIpSecVpnConfig.NsxtIpSecVpn.ID)

	// Check if Tunnel Profile has custom settings and apply them
	tunnelProfileConfig := getNsxtIpSecVpnProfileTunnelConfigurationType(d)

	// If `security_profile_customization` is defined, update the Ipsec VPN tunnel with the new profile
	_, customSecurityProfile := d.GetOk("security_profile_customization")
	if customSecurityProfile {
		_, err = createdIpSecVpnConfig.UpdateTunnelConnectionProperties(tunnelProfileConfig)
		if err != nil {
			return diag.Errorf("[nsx-t ipsec vpn tunnel create] error setting VPN Tunnel Profile: %s", err)
		}
	}

	return resourceVcdNsxtIpSecVpnTunnelRead(ctx, d, meta)
}

func resourceVcdNsxtIpSecVpnTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel update] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel update] error retrieving Edge Gateway: %s", err)
	}

	ipSecVpnConfiguration, err := nsxtEdge.GetIpSecVpnTunnelById(d.Id())
	if err != nil {
		diag.Errorf("[nsx-t ipsec vpn tunnel update] error retrieving existing NSX-T IPsec VPN Tunnel configuration: %s", err)
	}

	ipSecVpnConfig, err := getNsxtIpSecVpnTunnelType(d)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel update] error getting NSX-T IPsec VPN Tunnel configuration type: %s", err)
	}
	// Inject ID for update
	ipSecVpnConfig.ID = d.Id()

	// If unset, it will set it back to default, so before updating the VPN config, we need to check if the `security_profile_customization`
	// field is defined in the configuration
	_, customSecurityProfile := d.GetOk("security_profile_customization")
	if customSecurityProfile {
		ipSecVpnConfig.SecurityType = "CUSTOM"
	} else {
		ipSecVpnConfig.SecurityType = "DEFAULT"
	}

	if customSecurityProfile {
		// Get Security Tunnel profile from Terraform state
		ipSecTunnelProfileConfig := getNsxtIpSecVpnProfileTunnelConfigurationType(d)

		// To set IPsec VPN Tunnel Connection Profile - it must be updated (HTTP PUT) with all the options configured
		_, err = ipSecVpnConfiguration.UpdateTunnelConnectionProperties(ipSecTunnelProfileConfig)
		if err != nil {
			return diag.Errorf("[nsx-t ipsec vpn tunnel update] error updating NSX-T IPsec VPN Tunnel Security Profile: %s", err)
		}
	}

	_, err = ipSecVpnConfiguration.Update(ipSecVpnConfig)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel update] error updating NSX-T IPsec VPN Tunnel configuration '%s': %s", ipSecVpnConfig.Name, err)
	}

	return resourceVcdNsxtIpSecVpnTunnelRead(ctx, d, meta)
}

func resourceVcdNsxtIpSecVpnTunnelRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	ipSecVpnConfig, err := nsxtEdge.GetIpSecVpnTunnelById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
		}
		return diag.Errorf("error retrieving NSX-T IPsec VPN Tunnel configuration: %s", err)
	}

	// Set general schema for configuration
	err = setNsxtIpSecVpnTunnelData(d, ipSecVpnConfig.NsxtIpSecVpn)
	if err != nil {
		return diag.Errorf("error storing NSX-T IPsec VPN Tunnel configuration to schema: %s", err)
	}

	// Tunnel Security Properties
	tunnelConnectionProperties, err := ipSecVpnConfig.GetTunnelConnectionProperties()
	if err != nil {
		return diag.Errorf("error reading NSX-T IPsec VPN Tunnel Security Customization: %s", err)
	}

	err = setNsxtIpSecVpnProfileTunnelConfigurationData(d, tunnelConnectionProperties)
	if err != nil {
		return diag.Errorf("error storing NSX-T IPsec VPN Tunnel Security Customization to schema: %s", err)
	}

	// Read tunnel status data from separate endpoint
	tunnelStatus, err := ipSecVpnConfig.GetStatus()
	if err != nil {
		return diag.Errorf("error reading NSX-T IPsec VPN Tunnel status: %s", err)
	}
	setNsxtIpSecVpnTunnelStatusData(d, tunnelStatus)

	return nil
}

func resourceVcdNsxtIpSecVpnTunnelDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	unlock, err := vcdClient.lockParentVdcGroupOrEdgeGateway(d)
	if err != nil {
		return diag.Errorf("[nsx-t ipsec vpn tunnel delete] %s", err)
	}

	defer unlock()

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	ipSecVpnConfig, err := nsxtEdge.GetIpSecVpnTunnelById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving NSX-T IPsec VPN Tunnel configuration for deletion: %s", err)
	}

	err = ipSecVpnConfig.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T IPsec VPN Tunnel configuration: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdNsxtIpSecVpnTunnelImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T IPsec VPN Tunnel Import started")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.ipsec_tunnel_name")
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, ipSecVpnTunnelIdentifier := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, errors.New("vcd_nsxt_ipsec_vpn_tunnel is only supported by NSX-T VDCs")
	}

	edgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	ipSecVpnTunnel, err := edgeGateway.GetIpSecVpnTunnelByName(ipSecVpnTunnelIdentifier)
	if govcd.ContainsNotFound(err) {
		ipSecVpnTunnel, err = edgeGateway.GetIpSecVpnTunnelById(ipSecVpnTunnelIdentifier)
	}

	listStr := ""
	// Error occurred and it is not ErrorEntityNotFound. This means - more than configuration found and it should be
	// dumped their IDs so that user can pick ID
	if err != nil && !govcd.ContainsNotFound(err) {
		allRules, err2 := edgeGateway.GetAllIpSecVpnTunnels(nil)
		if err2 != nil {
			return nil, fmt.Errorf("error getting list of all IPsec VPN Tunnels: %s", err)
		}

		listStr = "\n" + getIpSecVpnTunnelsList(ipSecVpnTunnelIdentifier, allRules)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find IPsec VPN Tunnels '%s': %s%s", ipSecVpnTunnelIdentifier, err, listStr)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(ipSecVpnTunnel.NsxtIpSecVpn.ID)

	return []*schema.ResourceData{d}, nil
}

func getNsxtIpSecVpnTunnelType(d *schema.ResourceData) (*types.NsxtIpSecVpnTunnel, error) {
	ipSecVpnConfig := &types.NsxtIpSecVpnTunnel{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Enabled:     d.Get("enabled").(bool),
		LocalEndpoint: types.NsxtIpSecVpnTunnelLocalEndpoint{
			LocalId:       d.Get("local_ip_address").(string),
			LocalAddress:  d.Get("local_ip_address").(string),
			LocalNetworks: convertSchemaSetToSliceOfStrings(d.Get("local_networks").(*schema.Set)),
		},
		RemoteEndpoint: types.NsxtIpSecVpnTunnelRemoteEndpoint{
			RemoteAddress:  d.Get("remote_ip_address").(string),
			RemoteNetworks: convertSchemaSetToSliceOfStrings(d.Get("remote_networks").(*schema.Set)),
		},
		PreSharedKey:       d.Get("pre_shared_key").(string),
		Logging:            d.Get("logging").(bool),
		AuthenticationMode: d.Get("authentication_mode").(string),
	}

	// If remote_id is not set, use remote_ip_address for backwards compatibility and this is the
	// default behavior of VCD, but user can explicitly specify 'remote_id' value
	//
	// Using d.GetRawConfig() instead of d.Get because d.Get would return a previous computed value
	// of "remote_id" field even if it is not specified in HCL
	if !d.GetRawConfig().GetAttr("remote_id").IsNull() {
		ipSecVpnConfig.RemoteEndpoint.RemoteId = d.GetRawConfig().GetAttr("remote_id").AsString()
	} else {
		ipSecVpnConfig.RemoteEndpoint.RemoteId = d.Get("remote_ip_address").(string)
	}

	// Handle Certificate authentication mode
	certificateId := d.Get("certificate_id").(string)
	if certificateId != "" {
		ipSecVpnConfig.CertificateRef = &types.OpenApiReference{
			ID: certificateId,
		}
	}
	caCertificateId := d.Get("ca_certificate_id").(string)
	if caCertificateId != "" {
		ipSecVpnConfig.CaCertificateRef = &types.OpenApiReference{
			ID: caCertificateId,
		}
	}

	return ipSecVpnConfig, nil
}

func setNsxtIpSecVpnTunnelData(d *schema.ResourceData, ipSecVpnConfig *types.NsxtIpSecVpnTunnel) error {
	dSet(d, "name", ipSecVpnConfig.Name)
	dSet(d, "description", ipSecVpnConfig.Description)
	dSet(d, "pre_shared_key", ipSecVpnConfig.PreSharedKey)
	dSet(d, "enabled", ipSecVpnConfig.Enabled)
	dSet(d, "local_ip_address", ipSecVpnConfig.LocalEndpoint.LocalAddress)
	dSet(d, "enabled", ipSecVpnConfig.Enabled)
	dSet(d, "logging", ipSecVpnConfig.Logging)
	dSet(d, "security_profile", ipSecVpnConfig.SecurityType)
	dSet(d, "remote_id", ipSecVpnConfig.RemoteEndpoint.RemoteId)
	dSet(d, "authentication_mode", ipSecVpnConfig.AuthenticationMode)

	localNetworksSet := convertStringsToTypeSet(ipSecVpnConfig.LocalEndpoint.LocalNetworks)
	err := d.Set("local_networks", localNetworksSet)
	if err != nil {
		return fmt.Errorf("error storing 'local_networks': %s", err)
	}

	dSet(d, "remote_ip_address", ipSecVpnConfig.RemoteEndpoint.RemoteAddress)
	remoteNetworksSet := convertStringsToTypeSet(ipSecVpnConfig.RemoteEndpoint.RemoteNetworks)
	err = d.Set("remote_networks", remoteNetworksSet)
	if err != nil {
		return fmt.Errorf("error storing 'remote_networks': %s", err)
	}

	if ipSecVpnConfig.CertificateRef != nil {
		dSet(d, "certificate_id", ipSecVpnConfig.CertificateRef.ID)
	}
	if ipSecVpnConfig.CaCertificateRef != nil {
		dSet(d, "ca_certificate_id", ipSecVpnConfig.CaCertificateRef.ID)
	}

	return nil
}

func setNsxtIpSecVpnTunnelStatusData(d *schema.ResourceData, ipSecVpnStatus *types.NsxtIpSecVpnTunnelStatus) {
	dSet(d, "status", ipSecVpnStatus.TunnelStatus)
	dSet(d, "ike_service_status", ipSecVpnStatus.IkeStatus.IkeServiceStatus)
	dSet(d, "ike_fail_reason", ipSecVpnStatus.IkeStatus.FailReason)
}

func getNsxtIpSecVpnProfileTunnelConfigurationType(d *schema.ResourceData) *types.NsxtIpSecVpnTunnelSecurityProfile {
	tunnel, isSet := d.GetOk("security_profile_customization")

	if !isSet {
		return nil
	}
	tunnelSlice := tunnel.([]interface{})
	tunnelMap := tunnelSlice[0].(map[string]interface{})

	nsxtIpSecVpnTunnelProfile := &types.NsxtIpSecVpnTunnelSecurityProfile{
		SecurityType: "CUSTOM", // Security Type must become CUSTOM, because we are configuring profile
		IkeConfiguration: types.NsxtIpSecVpnTunnelProfileIkeConfiguration{
			IkeVersion:           tunnelMap["ike_version"].(string),
			EncryptionAlgorithms: convertSchemaSetToSliceOfStrings(tunnelMap["ike_encryption_algorithms"].(*schema.Set)),
			DigestAlgorithms:     convertSchemaSetToSliceOfStrings(tunnelMap["ike_digest_algorithms"].(*schema.Set)),
			DhGroups:             convertSchemaSetToSliceOfStrings(tunnelMap["ike_dh_groups"].(*schema.Set)),
			SaLifeTime:           addrOf(tunnelMap["ike_sa_lifetime"].(int)),
		},
		TunnelConfiguration: types.NsxtIpSecVpnTunnelProfileTunnelConfiguration{
			PerfectForwardSecrecyEnabled: tunnelMap["tunnel_pfs_enabled"].(bool),
			DfPolicy:                     tunnelMap["tunnel_df_policy"].(string),
			EncryptionAlgorithms:         convertSchemaSetToSliceOfStrings(tunnelMap["tunnel_encryption_algorithms"].(*schema.Set)),
			DigestAlgorithms:             convertSchemaSetToSliceOfStrings(tunnelMap["tunnel_digest_algorithms"].(*schema.Set)),
			DhGroups:                     convertSchemaSetToSliceOfStrings(tunnelMap["tunnel_dh_groups"].(*schema.Set)),
			SaLifeTime:                   addrOf(tunnelMap["tunnel_sa_lifetime"].(int)),
		},
		DpdConfiguration: types.NsxtIpSecVpnTunnelProfileDpdConfiguration{
			ProbeInterval: tunnelMap["dpd_probe_internal"].(int),
		},
	}

	return nsxtIpSecVpnTunnelProfile
}

func setNsxtIpSecVpnProfileTunnelConfigurationData(d *schema.ResourceData, tunnelConfig *types.NsxtIpSecVpnTunnelSecurityProfile) error {
	if tunnelConfig.SecurityType == "DEFAULT" {
		err := d.Set("security_profile_customization", nil)
		if err != nil {
			return fmt.Errorf("error resetting 'security_profile_customization' to empty: %s", err)
		}
		// Return early because there is nothing to store when DEFAULT profile is in use
		return nil
	}

	secProfileMap := make(map[string]interface{})
	secProfileMap["ike_version"] = tunnelConfig.IkeConfiguration.IkeVersion
	secProfileMap["ike_encryption_algorithms"] = convertStringsToTypeSet(tunnelConfig.IkeConfiguration.EncryptionAlgorithms)
	secProfileMap["ike_digest_algorithms"] = convertStringsToTypeSet(tunnelConfig.IkeConfiguration.DigestAlgorithms)
	secProfileMap["ike_dh_groups"] = convertStringsToTypeSet(tunnelConfig.IkeConfiguration.DhGroups)
	secProfileMap["ike_sa_lifetime"] = tunnelConfig.IkeConfiguration.SaLifeTime
	secProfileMap["tunnel_pfs_enabled"] = tunnelConfig.TunnelConfiguration.PerfectForwardSecrecyEnabled
	secProfileMap["tunnel_df_policy"] = tunnelConfig.TunnelConfiguration.DfPolicy
	secProfileMap["tunnel_encryption_algorithms"] = convertStringsToTypeSet(tunnelConfig.TunnelConfiguration.EncryptionAlgorithms)
	secProfileMap["tunnel_digest_algorithms"] = convertStringsToTypeSet(tunnelConfig.TunnelConfiguration.DigestAlgorithms)
	secProfileMap["tunnel_dh_groups"] = convertStringsToTypeSet(tunnelConfig.TunnelConfiguration.DhGroups)
	secProfileMap["tunnel_sa_lifetime"] = tunnelConfig.TunnelConfiguration.SaLifeTime
	secProfileMap["dpd_probe_internal"] = tunnelConfig.DpdConfiguration.ProbeInterval

	// wrap secProfileMap as first element into []interface{} to satisfy schema.TypeList requirement
	return d.Set("security_profile_customization", []interface{}{secProfileMap})
}

// getIpSecVpnTunnelsList is a helper for import. IPsec VPN tunnels don't enforce name uniqueness therefore it may
// be that user specifies a config with the same name. In that case IPsec VPN Tunnel details and their IDs are listed
// and then one will be able to import by using ID.
func getIpSecVpnTunnelsList(name string, allTunnels []*govcd.NsxtIpSecVpnTunnel) string {
	buf := new(bytes.Buffer)

	_, err := fmt.Fprintf(buf, "# The following IPsec VPN Tunnels with Name '%s' are available\n", name)
	if err != nil {
		logForScreen("vcd_vm_nsxt_ipsec_vpn_tunnel", fmt.Sprintf("error writing to buffer: %s", err))
	}
	_, err = fmt.Fprintf(buf, "# Please use ID instead of Name in import path to pick exact ipSecVpnTunnel\n")
	if err != nil {
		logForScreen("vcd_vm_nsxt_ipsec_vpn_tunnel", fmt.Sprintf("error writing to buffer: %s", err))
	}

	w := tabwriter.NewWriter(buf, 1, 1, 1, ' ', 0)
	_, err = fmt.Fprintln(w, "ID\tName\tLocal IP\tRemote IP")
	if err != nil {
		logForScreen("vcd_vm_nsxt_ipsec_vpn_tunnel", fmt.Sprintf("error writing to buffer: %s", err))
	}
	for _, ipSecVpnTunnel := range allTunnels {
		if ipSecVpnTunnel.NsxtIpSecVpn.Name != name {
			continue
		}

		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			ipSecVpnTunnel.NsxtIpSecVpn.ID, ipSecVpnTunnel.NsxtIpSecVpn.Name,
			ipSecVpnTunnel.NsxtIpSecVpn.LocalEndpoint.LocalAddress,
			ipSecVpnTunnel.NsxtIpSecVpn.RemoteEndpoint.RemoteAddress)
		if err != nil {
			logForScreen("vcd_vm_nsxt_ipsec_vpn_tunnel", fmt.Sprintf("error writing to buffer: %s", err))
		}
	}
	err = w.Flush()
	if err != nil {
		logForScreen("vcd_vm_nsxt_ipsec_vpn_tunnel", fmt.Sprintf("error flushing buffer: %s", err))
	}
	return buf.String()
}
