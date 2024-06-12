package viettelidc

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdIpAllocation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdIpAllocationCreate,
		ReadContext:   resourceVcdIpAllocationRead,
		UpdateContext: resourceVcdIpAllocationUpdate,
		DeleteContext: resourceVcdIpAllocationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdIpAllocationImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Org ID for IP Allocation",
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "IP Space ID for IP Allocation",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "Type of allocation. One of `FLOATING_IP``, `IP_PREFIX`",
				ValidateFunc: validation.StringInSlice([]string{"FLOATING_IP", "IP_PREFIX"}, false),
			},
			"usage_state": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Can be set to 'USED_MANUAL' to mark the IP Allocation for manual use",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Custom description can only be set when usage_state is set to 'USED_MANUAL'",
			},
			"prefix_length": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Computed:      true,
				Description:   "Required if 'type' is IP_PREFIX and no custom 'value` is provided",
				ConflictsWith: []string{"value"},
			},
			"value": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "IP address or CIDR to use. (VCD 10.4.2+)",
				ConflictsWith: []string{"prefix_length"},
				ForceNew:      true, // Once a particular IP or Prefix is allocated - its changes are ignored by the API

				// API supports allocation of IP ranges (e.g. 10.10.10.1-10.10.10.3), but this
				// results in multiple separate allocations with separate IDs which goes against
				// Terraform principle. We have 'quantity' field disabled for the same reason.
				// Users can define multiple instances of this resource to allocate IP ranges
				ValidateFunc: validation.StringDoesNotMatch(
					regexp.MustCompile("-"), // having a hyphen '-' in the value means that it is an IP range
					"This resource does not support allocating IP ranges due to Terraform resources map to single entity. "+
						"Please use multiple resource instances to allocate multiple IP addresses",
				),
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address or CIDR",
			},
			"used_by_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of entity that is using this allocation",
			},
			"allocation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Allocation date in ISO 8601 format (e.g. 2023-06-07T09:57:58.721Z)",
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address part",
			},

			// The resource supports 'quantity' parameter, however it goes against Terraform concept
			// of - "one resource:one entity" mapping therefore it is left commented for now unless
			// a real use case appears and we have to hack around Terraform schema
			// "quantity": {
			// 	Type:        schema.TypeString,
			// 	Optional:    true,
			// 	Default:     "1",
			// 	ForceNew:    true,
			// 	Description: "Number of entries to allocate",
			// },
		},
	}
}

func resourceVcdIpAllocationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space IP Allocation creation initiated")

	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)
	org, err := vcdClient.GetOrgById(orgId)
	if err != nil {
		return diag.Errorf("error getting Org by ID: %s", err)
	}

	ipSpaceId := d.Get("ip_space_id").(string)
	ipSpace, err := vcdClient.GetIpSpaceById(ipSpaceId)
	if err != nil {
		return diag.Errorf("error getting IP Space by ID '%s': %s", ipSpaceId, err)
	}

	allocationConfig := types.IpSpaceIpAllocationRequest{
		Type:     d.Get("type").(string),
		Quantity: addrOf(1),
	}

	prefixLength := d.Get("prefix_length").(string)
	if prefixLength != "" {
		intPrefixLength, _ := strconv.Atoi(prefixLength)
		allocationConfig.PrefixLength = &intPrefixLength
	}

	if d.Get("value").(string) != "" {
		// API throws an exception if this is specified for older versions
		if vcdClient.Client.APIVCDMaxVersionIs("< 37.2") {
			return diag.Errorf("'value' can only be specified on VCD 10.4.2+")
		}

		allocationConfig.Value = d.Get("value").(string)
		allocationConfig.Quantity = nil // Quantity field must not be set when 'value' is specified
	}

	allocation, err := ipSpace.AllocateIp(orgId, org.Org.Name, &allocationConfig)
	if err != nil {
		return diag.Errorf("error allocating IP: %s", err)
	}

	d.SetId(allocation[0].ID)
	dSet(d, "ip_address", allocation[0].Value)

	// Perform manual reservation if there is a request for USED_MANUAL (it always needs a separate
	// API call)
	// * UNUSED - the allocated IP is current not being used in the system.
	// * USED - the allocated IP is currently in use in the system. An allocated IP address or IP
	// Prefix is considered used if it is being used in network services such as NAT rule or in Org
	// VDC network definition.
	// * USED_MANUAL - manual usage reservation with custom description

	// If user specified
	usageState := d.Get("usage_state").(string)
	if usageState == "USED_MANUAL" {
		// Retrieve IP Allocation object
		ipSpaceAllocation, err := org.GetIpSpaceAllocationById(ipSpaceId, allocation[0].ID)
		if err != nil {
			return diag.Errorf("error retrieving IP Space IP Allocation after request: %s", err)
		}

		ipAllocationUpdate := &types.IpSpaceIpAllocation{
			ID:          allocation[0].ID,
			Type:        d.Get("type").(string),
			UsageState:  usageState,
			Description: d.Get("description").(string),
			Value:       ipSpaceAllocation.IpSpaceIpAllocation.Value,
		}

		_, err = ipSpaceAllocation.Update(ipAllocationUpdate)
		if err != nil {
			return diag.Errorf("error updating IP Space IP Allocation after creation: %s", err)
		}
	}

	return resourceVcdIpAllocationRead(ctx, d, meta)
}

func resourceVcdIpAllocationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space IP Allocation update initiated")

	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)
	org, err := vcdClient.GetOrgById(orgId)
	if err != nil {
		return diag.Errorf("error getting Org by ID: %s", err)
	}

	ipAllocation, err := org.GetIpSpaceAllocationById(ipSpaceId, d.Id())
	if err != nil {
		return diag.Errorf("error retrieving IP Allocation: %s", err)
	}

	if d.HasChange("usage_state") || d.HasChange("description") {
		ipAllocation.IpSpaceIpAllocation.UsageState = d.Get("usage_state").(string)
		ipAllocation.IpSpaceIpAllocation.Description = d.Get("description").(string)
		_, err = ipAllocation.Update(ipAllocation.IpSpaceIpAllocation)
		if err != nil {
			return diag.Errorf("error updating IP Space IP Allocation: %s", err)
		}
	}

	return resourceVcdIpAllocationRead(ctx, d, meta)
}

func resourceVcdIpAllocationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space IP Allocation read initiated")

	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)

	org, err := vcdClient.GetOrgById(orgId)
	if err != nil {
		return diag.Errorf("error getting Org by id: %s", err)
	}

	ipAllocation, err := org.GetIpSpaceAllocationById(ipSpaceId, d.Id())
	if err != nil {
		return diag.Errorf("error getting IP Space IP Allocation: %s", err)
	}

	dSet(d, "description", ipAllocation.IpSpaceIpAllocation.Description)
	dSet(d, "type", ipAllocation.IpSpaceIpAllocation.Type)
	if ipAllocation.IpSpaceIpAllocation.OrgRef != nil {
		dSet(d, "org_id", ipAllocation.IpSpaceIpAllocation.OrgRef.ID)
	}
	if ipAllocation.IpSpaceIpAllocation.UsedByRef != nil {
		dSet(d, "used_by_id", ipAllocation.IpSpaceIpAllocation.UsedByRef.ID)
	}
	dSet(d, "allocation_date", ipAllocation.IpSpaceIpAllocation.AllocationDate)
	dSet(d, "usage_state", ipAllocation.IpSpaceIpAllocation.UsageState)
	dSet(d, "ip_address", ipAllocation.IpSpaceIpAllocation.Value)

	// When IP Prefix is allocated, the returned value is in CIDR format (e.g. 192.168.1.0/24), and
	// although it can be split using Terraform native functions, we're adding a convenience layer for
	// users by splitting this address into IP and prefix length
	if ipAllocation.IpSpaceIpAllocation.Type == "IP_PREFIX" {
		splitCidr := strings.Split(ipAllocation.IpSpaceIpAllocation.Value, "/")
		if len(splitCidr) == 2 {
			dSet(d, "ip", splitCidr[0])
			dSet(d, "prefix_length", splitCidr[1])
		} else {
			util.Logger.Printf("[TRACE] resourceVcdIpAllocationRead Unable to store split CIDR (value %s, split length %d)\n",
				ipAllocation.IpSpaceIpAllocation.Value, len(ipAllocation.IpSpaceIpAllocation.Value))
		}
	}

	if ipAllocation.IpSpaceIpAllocation.Type == "FLOATING_IP" {
		dSet(d, "ip", ipAllocation.IpSpaceIpAllocation.Value)
	}

	return nil
}

func resourceVcdIpAllocationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] IP Space IP Allocation deletion initiated")

	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)
	ipSpaceId := d.Get("ip_space_id").(string)

	org, err := vcdClient.GetOrgById(orgId)
	if err != nil {
		return diag.Errorf("error getting Org by id: %s", err)
	}

	ipAllocation, err := org.GetIpSpaceAllocationById(ipSpaceId, d.Id())
	if err != nil {
		return diag.Errorf("error getting IP Space IP Allocation: %s", err)
	}

	err = ipAllocation.Delete()
	if err != nil {
		return diag.Errorf("error deleting IP Space IP Allocation: %s", err)
	}

	return nil
}

func resourceVcdIpAllocationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] IP Allocation import initiated")

	resourceURI := strings.SplitN(d.Id(), ImportSeparator, 4)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.ip-space-name.ip-allocation-type.ip-allocation-ip")
	}

	orgName, ipSpaceName, ipAllocationType, ipAllocationIp := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org '%s': %s", orgName, err)
	}

	ipSpace, err := vcdClient.GetIpSpaceByName(ipSpaceName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Space %s: %s ", ipSpaceName, err)
	}

	ipAllocation, err := org.GetIpSpaceAllocationByTypeAndValue(ipSpace.IpSpace.ID, ipAllocationType, ipAllocationIp, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving IP Allocation: %s", err)
	}

	// Only setting Org because VDC is a deprecated field. `owner_id` is set by resourceVcdNsxtEdgeGatewayRead by itself
	dSet(d, "org_id", org.Org.ID)
	dSet(d, "ip_space_id", ipSpace.IpSpace.ID)
	dSet(d, "type", ipSpace.IpSpace.Type)

	d.SetId(ipAllocation.IpSpaceIpAllocation.ID)

	return []*schema.ResourceData{d}, nil
}
