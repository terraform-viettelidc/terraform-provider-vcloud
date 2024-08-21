package viettelidc

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// natRuleTypeGetter sets a type for getDnatRuleType and getSnatRuleType so that both can be accepted
// as function parameters
type natRuleTypeGetter func(d *schema.ResourceData, edgeGateway govcd.EdgeGateway) (*types.EdgeNatRule, error)

// natRuleDataSetter sets a type for setDatRuleData and setSnatRuleData so that both can be accepted
// as function parameters
type natRuleDataSetter func(d *schema.ResourceData, natRule *types.EdgeNatRule, edgeGateway govcd.EdgeGateway) error

// natRuleCreate returns a schema.CreateFunc for both SNAT and DNAT rules
func natRuleCreate(natType string, setData natRuleDataSetter, getNatRule natRuleTypeGetter) schema.CreateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		vcdClient := meta.(*VCDClient)
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)

		edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		natRule, err := getNatRule(d, *edgeGateway)
		if err != nil {
			return fmt.Errorf("unable to make structure for API call: %s", err)
		}

		natRule.Action = natType

		createdNatRule, err := edgeGateway.CreateNsxvNatRule(natRule)
		if err != nil {
			return fmt.Errorf("error creating new NAT rule: %s", err)
		}

		d.SetId(createdNatRule.ID)
		return natRuleRead("id", natType, setData)(d, meta)
	}
}

// natRuleUpdate returns a schema.UpdateFunc for both SNAT and DNAT rules
func natRuleUpdate(natType string, setData natRuleDataSetter, getNatRule natRuleTypeGetter) schema.UpdateFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		vcdClient := meta.(*VCDClient)
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)

		edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		updateNatRule, err := getNatRule(d, *edgeGateway)
		if err != nil {
			return fmt.Errorf("unable to make structure for API call: %s", err)
		}
		updateNatRule.ID = d.Id()

		updateNatRule.Action = natType

		updatedNatRule, err := edgeGateway.UpdateNsxvNatRule(updateNatRule)
		if err != nil {
			return fmt.Errorf("unable to update NAT rule with ID %s: %s", d.Id(), err)
		}

		err = setData(d, updatedNatRule, *edgeGateway)
		if err != nil {
			return fmt.Errorf("error setting data: %s", err)
		}

		return natRuleRead("id", natType, setData)(d, meta)
	}
}

// natRuleRead returns a schema.ReadFunc for both SNAT and DNAT rules
// ifField: specifies field name which holds NAT rule ID for lookup. In data sources it is rule_id
// while in resources it is simply ID
// natType: 'snat' or 'dnat'
func natRuleRead(idField, natType string, setData natRuleDataSetter) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		vcdClient := meta.(*VCDClient)

		edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		// if default ID field 'id' is used, then rely on Terraform's d.Id(). Otherwise use the
		// string value
		var idValue string
		if idField == "id" {
			idValue = d.Id()
		} else {
			idValue = d.Get(idField).(string)
		}

		readNatRule, err := edgeGateway.GetNsxvNatRuleById(idValue)
		if err != nil {
			d.SetId("")
			return fmt.Errorf("unable to find NAT (%s) rule with ID '%s': %s", natType, idValue, err)
		}

		if strings.ToLower(readNatRule.Action) != natType {
			return fmt.Errorf("NAT rule with id (%s) is of type %s, but expected type %s",
				readNatRule.ID, readNatRule.Action, natType)
		}

		d.SetId(readNatRule.ID)
		return setData(d, readNatRule, *edgeGateway)
	}
}

// natRuleDelete returns a schema.DeleteFunc for both SNAT and DNAT rules
func natRuleDelete(natType string) schema.DeleteFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		vcdClient := meta.(*VCDClient)
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)

		edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		err = edgeGateway.DeleteNsxvNatRuleById(d.Id())
		if err != nil {
			return fmt.Errorf("error deleting NAT rule of type %s: %s", natType, err)
		}

		d.SetId("")
		return nil
	}
}

// natRuleImporter returns a schema.StateFunc for both SNAT and DNAT rules
func natRuleImport(natType string) schema.StateFunc {
	return func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		resourceURI := strings.Split(d.Id(), ImportSeparator)
		if len(resourceURI) != 4 {
			return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw.rule-id")
		}
		orgName, vdcName, edgeName, natRuleId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

		vcdClient := meta.(*VCDClient)
		edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
		if err != nil {
			return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		readNatRule, err := edgeGateway.GetNsxvNatRuleById(natRuleId)
		if err != nil {
			return []*schema.ResourceData{}, fmt.Errorf("unable to find NAT rule with id %s: %s",
				d.Id(), err)
		}

		if strings.ToLower(readNatRule.Action) != natType {
			return []*schema.ResourceData{}, fmt.Errorf("NAT rule with id %s is of type %s. "+
				"Expected type %s. Please use correct resource",
				readNatRule.ID, readNatRule.Action, natType)
		}

		dSet(d, "org", orgName)
		dSet(d, "vdc", vdcName)
		dSet(d, "edge_gateway", edgeName)

		d.SetId(readNatRule.ID)
		return []*schema.ResourceData{d}, nil
	}
}

// getvNicIndexFromNetworkNameType helps to find edge gateway vNic index number
// (needed for NAT rules) by network_name and network_type
func getvNicIndexFromNetworkNameType(networkName, networkType string, edgeGateway govcd.EdgeGateway) (*int, error) {
	var edgeGatewayNetworkType string
	switch networkType {
	case "ext":
		edgeGatewayNetworkType = types.EdgeGatewayVnicTypeUplink
	case "org":
		edgeGatewayNetworkType = types.EdgeGatewayVnicTypeInternal
	}

	vnicIndex, err := edgeGateway.GetVnicIndexByNetworkNameAndType(networkName, edgeGatewayNetworkType)
	// if `org` network of type `types.EdgeGatewayVnicTypeInternal` network was not found - try to
	// look for it in subinterface `types.EdgeGatewayVnicTypeSubinterface`
	if networkType == "org" && govcd.IsNotFound(err) {
		vnicIndex, err = edgeGateway.GetVnicIndexByNetworkNameAndType(networkName, types.EdgeGatewayVnicTypeSubinterface)
	}

	if err != nil {
		return nil, fmt.Errorf("can't find network '%s' of type '%s' attached to edge gateway "+
			"therefore unable to identify vNic: %s", networkName, networkType, err)
	}

	return vnicIndex, nil
}

// getNetworkNameTypeFromVnicIndex is a reverse function to getvNicIndexFromNetworkNameType and
// helps to find edge gateway attached network_name and network_type by vNic index number
func getNetworkNameTypeFromVnicIndex(index int, edgeGateway govcd.EdgeGateway) (string, string, error) {
	networkName, networkType, err := edgeGateway.GetNetworkNameAndTypeByVnicIndex(index)
	if err != nil {
		return "", "", fmt.Errorf("unable to determine network name and type: %s", err)
	}

	var resourceNetworkType string
	switch networkType {
	case "uplink":
		resourceNetworkType = "ext"
	case "internal":
		resourceNetworkType = "org"
	case "subinterface":
		resourceNetworkType = "org"
	}

	return networkName, resourceNetworkType, nil
}
