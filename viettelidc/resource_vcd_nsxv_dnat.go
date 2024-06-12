package viettelidc

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDnat() *schema.Resource {
	return &schema.Resource{
		Create: natRuleCreate("dnat", setDnatRuleData, getDnatRule),
		Read:   natRuleRead("id", "dnat", setDnatRuleData),
		Update: natRuleUpdate("dnat", setDnatRuleData, getDnatRule),
		Delete: natRuleDelete("dnat"),
		Importer: &schema.ResourceImporter{
			State: natRuleImport("dnat"),
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
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"network_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Org or external network name",
			},
			"network_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ext", "org"}, false),
				Description:  "Network type. One of 'ext', 'org'",
			},
			"rule_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Read only. Possible values 'user', 'internal_high'",
			},
			"rule_tag": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Optional. Allows to set custom rule tag",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "Whether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Whether logging should be enabled for this rule. Default 'false'",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "NAT rule description",
			},
			"original_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
				Description: "Original address or address range. This is the " +
					"the destination address for DNAT rules.",
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				ValidateFunc:     validateCase("lower"),
				Description:      "Protocol. Such as 'tcp', 'udp', 'icmp', 'any'",
			},
			"icmp_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validateCase("lower"),
				Description: "ICMP type. Only supported when protocol is ICMP. One of `any`, " +
					"`address-mask-request`, `address-mask-reply`, `destination-unreachable`, `echo-request`, " +
					"`echo-reply`, `parameter-problem`, `redirect`, `router-advertisement`, `router-solicitation`, " +
					"`source-quench`, `time-exceeded`, `timestamp-request`, `timestamp-reply`. Default `any`",
			},
			"original_port": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Original port. This is the destination port for DNAT rules",
			},
			"translated_address": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Description: "Translated address or address range",
			},
			"translated_port": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: suppressWordToEmptyString("any"),
				Description:      "Translated port",
			},
		},
	}
}

// getDnatRule is responsible for getting types.EdgeNatRule for DNAT rule from Terraform
// configuration
func getDnatRule(d *schema.ResourceData, edgeGateway govcd.EdgeGateway) (*types.EdgeNatRule, error) {
	networkName := d.Get("network_name").(string)
	networkType := d.Get("network_type").(string)

	vnicIndex, err := getvNicIndexFromNetworkNameType(networkName, networkType, edgeGateway)
	if err != nil {
		return nil, err
	}

	natRule := &types.EdgeNatRule{
		Enabled:           d.Get("enabled").(bool),
		LoggingEnabled:    d.Get("logging_enabled").(bool),
		Description:       d.Get("description").(string),
		Vnic:              vnicIndex,
		OriginalAddress:   d.Get("original_address").(string),
		Protocol:          d.Get("protocol").(string),
		IcmpType:          d.Get("icmp_type").(string),
		OriginalPort:      d.Get("original_port").(string),
		TranslatedAddress: d.Get("translated_address").(string),
		TranslatedPort:    d.Get("translated_port").(string),
	}

	if ruleTag, ok := d.GetOk("rule_tag"); ok {
		natRule.RuleTag = strconv.Itoa(ruleTag.(int))
	}

	return natRule, nil
}

// setDnatRuleData is responsible for setting DNAT rule data into the statefile
func setDnatRuleData(d *schema.ResourceData, natRule *types.EdgeNatRule, edgeGateway govcd.EdgeGateway) error {
	networkName, resourceNetworkType, err := getNetworkNameTypeFromVnicIndex(*natRule.Vnic, edgeGateway)
	if err != nil {
		return err
	}

	if natRule.RuleTag != "" {
		value, err := strconv.Atoi(natRule.RuleTag)
		if err != nil {
			return fmt.Errorf("could not convert ruletag (%s) from string to int: %s", natRule.RuleTag, err)
		}
		dSet(d, "rule_tag", value)
	}

	dSet(d, "network_type", resourceNetworkType)
	dSet(d, "network_name", networkName)
	dSet(d, "enabled", natRule.Enabled)
	dSet(d, "logging_enabled", natRule.LoggingEnabled)
	dSet(d, "description", natRule.Description)
	dSet(d, "original_address", natRule.OriginalAddress)
	dSet(d, "protocol", natRule.Protocol)
	dSet(d, "icmp_type", natRule.IcmpType)
	dSet(d, "original_port", natRule.OriginalPort)
	dSet(d, "translated_address", natRule.TranslatedAddress)
	dSet(d, "translated_port", natRule.TranslatedPort)
	dSet(d, "rule_type", natRule.RuleType)

	return nil
}
