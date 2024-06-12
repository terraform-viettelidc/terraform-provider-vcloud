package viettelidc

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdPortgroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourcePortgroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Tier-0 router.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Portgroup type. One of 'NETWORK', 'DV_PORTGROUP'",
				ValidateFunc: validation.StringInSlice([]string{types.ExternalNetworkBackingTypeNetwork, types.ExternalNetworkBackingDvPortgroup}, false),
			},
		},
	}
}

func datasourcePortgroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	portGroupType := d.Get("type").(string)
	portGroupName := d.Get("name").(string)

	var err error
	var pgs []*types.PortGroupRecordType

	switch portGroupType {
	// Standard vSwitch portgroup
	case types.ExternalNetworkBackingTypeNetwork:
		pgs, err = govcd.QueryNetworkPortGroup(vcdClient.VCDClient, portGroupName)
	// Distributed switch portgroup
	case types.ExternalNetworkBackingDvPortgroup:
		pgs, err = govcd.QueryDistributedPortGroup(vcdClient.VCDClient, portGroupName)
	default:
		return diag.Errorf("unrecognized portgroup_type: %s", portGroupType)
	}

	if err != nil {
		return diag.Errorf("error querying for portgroups '%s' of type '%s': %s", portGroupName, portGroupType, err)
	}

	if len(pgs) == 0 {
		return diag.Errorf("%s: expected to get exactly one portgroup with name '%s' of type '%s', got %d",
			govcd.ErrorEntityNotFound, portGroupName, portGroupType, len(pgs))
	}

	if len(pgs) > 1 {
		return diag.Errorf("expected to get exactly one portgroup with name '%s' of type '%s', got %d",
			portGroupName, portGroupType, len(pgs))
	}

	d.SetId(pgs[0].MoRef)
	return nil
}
