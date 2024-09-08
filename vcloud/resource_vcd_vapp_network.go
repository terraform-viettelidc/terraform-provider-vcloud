package vcloud

//lint:file-ignore SA1019 ignore GetOkExists deprecated function error
import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdVappNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVappNetworkCreate,
		ReadContext:   resourceVappNetworkRead,
		UpdateContext: resourceVappNetworkUpdate,
		DeleteContext: resourceVappAndVappOrgNetworkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVappNetworkImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network name",
				// we can't change network name as this results in ID (HREF) change
			},
			"vapp_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp to use",
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
			"netmask": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Deprecated:   "Use prefix_length instead which supports both IPv4 and IPv6",
				Description:  "Netmask address for a subnet.",
				ExactlyOneOf: []string{"prefix_length", "netmask"},
			},
			"prefix_length": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "Prefix length for a subnet",
				ExactlyOneOf: []string{"netmask", "prefix_length"},
				ValidateFunc: IsIntAndAtLeast(0),
			},
			"gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Gateway of the network",
			},
			"dns1": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Primary DNS server",
			},
			"dns2": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Secondary DNS server",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "DNS suffix",
			},

			"guest_vlan_allowed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if Network allows guest VLAN tagging",
			},
			"org_network_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "org network name to which vapp network is connected",
			},
			"retain_ip_mac_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Specifies whether the network resources such as IP/MAC of router will be retained across deployments. Default is false.",
			},
			"dhcp_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"default_lease_time": {
							Type:     schema.TypeInt,
							Default:  3600,
							Optional: true,
						},

						"max_lease_time": {
							Type:     schema.TypeInt,
							Default:  7200,
							Optional: true,
						},

						"enabled": {
							Type:     schema.TypeBool,
							Default:  true,
							Optional: true,
						},
					},
				},
				Set: resourceVcdDhcpPoolHash,
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:     schema.TypeString,
							Required: true,
						},

						"end_address": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceVcdNetworkStaticIpPoolHash,
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

func resourceVappNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	vappNetworkName := d.Get("name").(string)
	vappNetworkSettings := &govcd.VappNetworkSettings{
		Name:               vappNetworkName,
		Description:        d.Get("description").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		SubnetPrefixLength: d.Get("prefix_length").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		RetainIpMacEnabled: addrOf(d.Get("retain_ip_mac_enabled").(bool)),
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	expandDhcpPool(d, vappNetworkSettings)

	var orgVdcNetwork *types.OrgVDCNetwork
	if networkId, ok := d.GetOk("org_network_name"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkId.(string), true)
		if err != nil {
			return diag.FromErr(err)
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}
	vAppNetworkConfig, err := vapp.CreateVappNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return diag.Errorf("error creating vApp network. %s", err)
	}

	vAppNetwork := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == vappNetworkName {
			vAppNetwork = networkConfig
		}
	}

	if vAppNetwork == (types.VAppNetworkConfiguration{}) {
		return diag.Errorf("didn't find vApp network: %s", vappNetworkName)
	}

	// Parsing UUID from 'https://HOST/api/admin/network/6ced8e2f-29dd-4201-9801-a02cb8bed821/action/reset' or similar
	networkId, err := govcd.GetUuidFromHref(vAppNetwork.Link.HREF, false)
	if err != nil {
		return diag.Errorf("unable to get network ID from HREF: %s", err)
	}
	d.SetId(normalizeId("urn:vcloud:network:", networkId))

	return resourceVappNetworkRead(ctx, d, meta)
}

func expandDhcpPool(d *schema.ResourceData, vappNetworkSettings *govcd.VappNetworkSettings) {
	if dhcp, ok := d.GetOk("dhcp_pool"); ok && len(dhcp.(*schema.Set).List()) > 0 {
		for _, item := range dhcp.(*schema.Set).List() {
			data := item.(map[string]interface{})
			vappNetworkSettings.DhcpSettings = &govcd.DhcpSettings{
				IsEnabled:        data["enabled"].(bool),
				DefaultLeaseTime: data["default_lease_time"].(int),
				MaxLeaseTime:     data["max_lease_time"].(int),
				IPRange: &types.IPRange{StartAddress: data["start_address"].(string),
					EndAddress: data["end_address"].(string)}}
		}
	}
}

func resourceVappNetworkRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVappNetworkRead(d, meta, "resource")
}

func genericVappNetworkRead(d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			log.Printf("vApp found. Removing vApp network from state file: %s", err)
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
	vappNetworkName := d.Get("name").(string)
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.Link != nil {
			networkId, err = govcd.GetUuidFromHref(networkConfig.Link.HREF, false)
			if err != nil {
				return diag.Errorf("unable to get network ID from HREF: %s", err)
			}
			// Check name as well to support old resource IDs that are names and datasources that have names provided by the user
			if extractUuid(d.Id()) == extractUuid(networkId) || networkConfig.NetworkName == vappNetworkName {
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
		return diag.Errorf("[VAPP network read] %s : %s", govcd.ErrorEntityNotFound, vappNetworkName)
	}

	// needs to set for datasource. Do not set always as keep back compatibility when ID was name.
	if d.Id() == "" {
		d.SetId(normalizeId("urn:vcloud:network:", networkId))
	}
	dSet(d, "description", vAppNetwork.Description)
	if config := vAppNetwork.Configuration; config != nil {
		if config.IPScopes != nil {
			dSet(d, "gateway", config.IPScopes.IPScope[0].Gateway)
			dSet(d, "netmask", config.IPScopes.IPScope[0].Netmask)
			dSet(d, "prefix_length", config.IPScopes.IPScope[0].SubnetPrefixLength)
			dSet(d, "dns1", config.IPScopes.IPScope[0].DNS1)
			dSet(d, "dns2", config.IPScopes.IPScope[0].DNS2)
			dSet(d, "dns_suffix", config.IPScopes.IPScope[0].DNSSuffix)
		}
		if config.Features != nil && config.Features.DhcpService != nil {
			transformed := schema.NewSet(resourceVcdDhcpPoolHash, []interface{}{})
			newValues := map[string]interface{}{
				"enabled":            config.Features.DhcpService.IsEnabled,
				"max_lease_time":     config.Features.DhcpService.MaxLeaseTime,
				"default_lease_time": config.Features.DhcpService.DefaultLeaseTime,
			}
			if config.Features.DhcpService.IPRange != nil {
				newValues["start_address"] = config.Features.DhcpService.IPRange.StartAddress
				// when only start address provided, API returns end address same as start address
				if config.Features.DhcpService.IPRange.StartAddress != config.Features.DhcpService.IPRange.EndAddress {
					newValues["end_address"] = config.Features.DhcpService.IPRange.EndAddress
				}
			}
			transformed.Add(newValues)
			err = d.Set("dhcp_pool", transformed)
			if err != nil {
				return diag.Errorf("[vApp network DHCP pool read] set issue: %s", err)
			}
		}

		if config.IPScopes != nil && config.IPScopes.IPScope[0].IPRanges != nil {
			staticIpRanges := schema.NewSet(resourceVcdNetworkStaticIpPoolHash, []interface{}{})
			for _, ipRange := range config.IPScopes.IPScope[0].IPRanges.IPRange {
				newValues := map[string]interface{}{
					"start_address": ipRange.StartAddress,
					"end_address":   ipRange.EndAddress,
				}
				staticIpRanges.Add(newValues)
			}
			err = d.Set("static_ip_pool", staticIpRanges)
			if err != nil {
				return diag.Errorf("[vApp network static pool read] set issue: %s", err)
			}
		}

		dSet(d, "guest_vlan_allowed", *config.GuestVlanAllowed)
		if config.ParentNetwork != nil {
			dSet(d, "org_network_name", config.ParentNetwork.Name)
		} else {
			dSet(d, "org_network_name", nil)
		}
		dSet(d, "retain_ip_mac_enabled", config.RetainNetInfoAcrossDeployments)
	}
	return nil
}

func resourceVappNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// reboot_vapp_on_removal does not have any effect on update therefore skipping update if only
	// this field was modified
	if !d.HasChangeExcept("reboot_vapp_on_removal") {
		return resourceVappNetworkRead(ctx, d, meta)
	}

	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return diag.Errorf("error finding vApp. %s", err)
	}

	staticIpRanges, err := expandIPRange(d.Get("static_ip_pool").(*schema.Set).List())
	if err != nil {
		return diag.FromErr(err)
	}

	vappNetworkSettings := &govcd.VappNetworkSettings{
		ID:                 d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Gateway:            d.Get("gateway").(string),
		NetMask:            d.Get("netmask").(string),
		SubnetPrefixLength: d.Get("prefix_length").(string),
		DNS1:               d.Get("dns1").(string),
		DNS2:               d.Get("dns2").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		StaticIPRanges:     staticIpRanges.IPRange,
		RetainIpMacEnabled: addrOf(d.Get("retain_ip_mac_enabled").(bool)),
	}

	if _, ok := d.GetOk("guest_vlan_allowed"); ok {
		convertedValue := d.Get("guest_vlan_allowed").(bool)
		vappNetworkSettings.GuestVLANAllowed = &convertedValue
	}

	expandDhcpPool(d, vappNetworkSettings)

	var orgVdcNetwork *types.OrgVDCNetwork
	if networkName, ok := d.GetOk("org_network_name"); ok {
		orgNetwork, err := vdc.GetOrgVdcNetworkByNameOrId(networkName.(string), true)
		if err != nil {
			return diag.FromErr(err)
		}
		orgVdcNetwork = orgNetwork.OrgVDCNetwork
	}

	_, err = vapp.UpdateNetwork(vappNetworkSettings, orgVdcNetwork)
	if err != nil {
		return diag.Errorf("error creating vApp network. %s", err)
	}
	return resourceVappNetworkRead(ctx, d, meta)
}

// resourceVappAndVappOrgNetworkDelete deletes a vApp network.
// Starting with VCD 10.4.1, a vApp network cannot be deleted from a powered on vApp. To avoid
// inconvenience (especially in `terraform destroy` scenarios), there is a
// 'reboot_vapp_on_removal=true' flag.
// When the flag is set and vApp status is not POWERED_OFF or RESOLVED, the vApp is powered off, the network is
// deleted and the vApp powered on (if it was not powered off before).
//
// Note. This function is used for both resource `vcd_vapp_network` and `vcd_vapp_org_network`
// because deletion of these networks is the same operation and maintaining two functions might
// become inconsistent. They can be split again, if required.
func resourceVappAndVappOrgNetworkDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentVapp(d)
	defer vcdClient.unLockParentVapp(d)

	// Should vApp be power cycled before deleting network? ('reboot_vapp_on_removal=true')
	rebootVAppOnRemoval := false
	vappStatusBeforeOperation := ""
	vAppRebootEnabled, isSet := d.GetOkExists("reboot_vapp_on_removal")
	if isSet && vAppRebootEnabled.(bool) {
		util.Logger.Printf("[TRACE] reboot_vapp_on_removal=true is enabled with parent vApp '%s",
			d.Get("vapp_name").(string))
		rebootVAppOnRemoval = true
	}

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByName(d.Get("vapp_name").(string), false)
	if err != nil {
		return diag.Errorf("error finding vApp: %s", err)
	}

	if rebootVAppOnRemoval {
		vappStatusBeforeOperation, err = vapp.GetStatus()
		if err != nil {
			return diag.Errorf("error getting vApp '%s' status before vApp network removal: %s",
				vapp.VApp.Name, err)
		}

		util.Logger.Printf("[TRACE] reboot_vapp_on_removal=true, vApp '%s' status before network removal is '%s'",
			vapp.VApp.Name, vappStatusBeforeOperation)
		if vappStatusBeforeOperation != "POWERED_OFF" && vappStatusBeforeOperation != "RESOLVED" {
			util.Logger.Println("[TRACE] reboot_vapp_on_removal=true, powering off vApp")
			task, err := vapp.Undeploy() // UI Button "Power Off" calls undeploy API endpoint
			if err != nil {
				return diag.Errorf("error Powering Off: %s", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return diag.Errorf("error completing vApp Power Off task: %s", err)
			}
		}
	}

	// Remove vApp network
	_, err = vapp.RemoveNetwork(d.Id())
	if err != nil {
		// VCD 10.4.1+ API returns error when removing network from powered off vApp
		// If this error occurs - we add a hint to use 'reboot_vapp_on_removal' flag
		if strings.Contains(err.Error(), "Stop the vApp and try again") {
			return diag.Errorf("error removing vApp network: %s \n\n"+
				"Parent vApp '%s' must be powered off in VCD 10.4.1+ to remove a vApp network. \n"+
				"You can use 'reboot_vapp_on_removal=true' flag to power off vApp before removing network.",
				vapp.VApp.Name, err)
		}

		return diag.Errorf("error removing vApp network: %s", err)
	}

	// If vApp was not powered off before and 'reboot_vapp_on_removal' flag was used - power it on
	// again. The reason we check for vappStatusBeforeOperation != "POWERED_ON" is that a vApp could
	// have had different states than "POWERED_OFF" and "POWERED_ON" (e.g. "PARTIALLY_POWERED_OFF"),
	// but we cannot restore exactly such state. So we restore "POWERED_ON" state.
	if rebootVAppOnRemoval && vappStatusBeforeOperation != "POWERED_OFF" && vappStatusBeforeOperation != "RESOLVED" {
		util.Logger.Printf("[TRACE] reboot_vapp_on_removal=true, restoring vApp '%s' power state, was '%s'",
			vapp.VApp.Name, vappStatusBeforeOperation)
		task, err := vapp.PowerOn()
		if err != nil {
			return diag.Errorf("error powering on vApp: %s", err)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return diag.Errorf("error completing vApp power on task: %s", err)
		}
	}

	d.SetId("")

	return nil
}

// resourceVcdVappNetworkImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_network.network_name
// Example import path (_the_id_string_): org-name.vdc-name.vapp-name.network-name
func resourceVcdVappNetworkImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("[vApp network import] resource name must be specified as org-name.vdc-name.vapp-name.network-name")
	}
	orgName, vdcName, vappName, networkName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] unable to find VDC %s: %s ", vdcName, err)
	}

	vapp, err := vdc.GetVAppByName(vappName, false)
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] error retrieving vapp %s: %s", vappName, err)
	}
	vAppNetworkConfig, err := vapp.GetNetworkConfig()
	if err != nil {
		return nil, fmt.Errorf("[vApp network import] error retrieving vApp network configuration %s: %s", networkName, err)
	}

	vappNetworkToImport := types.VAppNetworkConfiguration{}
	for _, networkConfig := range vAppNetworkConfig.NetworkConfig {
		if networkConfig.NetworkName == networkName {
			vappNetworkToImport = networkConfig
			break
		}
	}

	if vappNetworkToImport == (types.VAppNetworkConfiguration{}) {
		return nil, fmt.Errorf("didn't find vApp network: %s", networkName)
	}

	if isVappOrgNetwork(&vappNetworkToImport) {
		return nil, fmt.Errorf("found vApp org network, not vApp network: %s", networkName)
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
	dSet(d, "name", networkName)
	dSet(d, "vapp_name", vappName)

	return []*schema.ResourceData{d}, nil
}

func resourceVcdDhcpPoolHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	_, err := buf.WriteString(fmt.Sprintf("%t-", m["enabled"].(bool)))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	_, err = buf.WriteString(fmt.Sprintf("%d-", m["max_lease_time"].(int)))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	_, err = buf.WriteString(fmt.Sprintf("%d-", m["default_lease_time"].(int)))
	if err != nil {
		util.Logger.Printf("[ERROR] error writing to string: %s", err)
	}
	if m["start_address"] != nil {
		_, err = buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["start_address"].(string))))
		if err != nil {
			util.Logger.Printf("[ERROR] error writing to string: %s", err)
		}
	}
	if m["end_address"] != nil {
		_, err = buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["end_address"].(string))))
		if err != nil {
			util.Logger.Printf("[ERROR] error writing to string: %s", err)
		}
	}
	return hashcodeString(buf.String())
}

// Allows to identify if vApp Org network and not vApp network
func isVappOrgNetwork(networkConfig *types.VAppNetworkConfiguration) bool {
	return !govcd.IsVappNetwork(networkConfig.Configuration)
}
