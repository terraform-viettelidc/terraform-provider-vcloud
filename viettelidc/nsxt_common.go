package viettelidc

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway returns VDC or VDC group ID and NSX-T Edge Gateway type
func getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient *VCDClient, d *schema.ResourceData, actionMessage string) (string, *govcd.NsxtEdgeGateway, error) {
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return "", nil, fmt.Errorf("[%s] error retrieving Org: %s", actionMessage, err)
	}

	// Lookup Edge Gateway to know parent VDC or VDC Group
	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return "", nil, fmt.Errorf("[%s] error retrieving Edge Gateway structure: %s", actionMessage, err)
	}
	if anyEdgeGateway.IsNsxv() {
		return "", nil, fmt.Errorf("[%s] NSX-V edge gateway not supported", actionMessage)
	}

	nsxtEdgeGateway, err := anyEdgeGateway.GetNsxtEdgeGateway()
	if err != nil {
		return "", nil, fmt.Errorf("[%s] could not retrieve NSX-T Edge Gateway with ID '%s': %s", actionMessage, d.Id(), err)
	}

	return anyEdgeGateway.EdgeGateway.OwnerRef.ID, nsxtEdgeGateway, nil
}
