package vcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdRdeTypeBehavior() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeTypeBehaviorCreate,
		ReadContext:   resourceVcdRdeTypeBehaviorRead,
		UpdateContext: resourceVcdRdeTypeBehaviorUpdate,
		DeleteContext: resourceVcdRdeTypeBehaviorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeTypeBehaviorImport,
		},
		Schema: map[string]*schema.Schema{
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the RDE Type that owns the Behavior override",
			},
			"rde_interface_behavior_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the original RDE Interface Behavior to override",
			},
			"execution": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Execution map of the Behavior that overrides the original",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the contract of the overridden Behavior",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the overridden Behavior",
			},
			"ref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Behavior invocation reference to be used for polymorphic behavior invocations",
			},
		},
	}
}

func resourceVcdRdeTypeBehaviorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdRdeTypeBehaviorCreateOrUpdate(ctx, d, meta, "create")
}

func resourceVcdRdeTypeBehaviorCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Type Behavior %s] could not retrieve the RDE Type with ID '%s': %s", operation, rdeTypeId, err)
	}
	payload := types.Behavior{
		ID:        d.Get("rde_interface_behavior_id").(string),
		Ref:       d.Get("ref").(string),
		Execution: d.Get("execution").(map[string]interface{}),
	}
	if desc, ok := d.GetOk("description"); ok {
		payload.Description = desc.(string)
	}
	_, err = rdeType.UpdateBehaviorOverride(payload)
	if err != nil {
		return diag.Errorf("[RDE Type Behavior %s] could not %s the Behavior in the RDE Type with ID '%s': %s", operation, operation, rdeTypeId, err)
	}
	return genericVcdRdeTypeBehaviorRead(ctx, d, meta, "resource")
}

func resourceVcdRdeTypeBehaviorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeTypeBehaviorRead(ctx, d, meta, "resource")
}

func genericVcdRdeTypeBehaviorRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Type Behavior read] could not read the Behavior of RDE Type with ID '%s': %s", rdeTypeId, err)
	}

	var behaviorId string
	if origin == "datasource" {
		behaviorId = d.Get("behavior_id").(string)
	} else {
		behaviorId = d.Get("rde_interface_behavior_id").(string)
	}

	behavior, err := rdeType.GetBehaviorById(behaviorId)
	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] RDE Type Behavior no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[RDE Type Behavior read] could not read the Behavior '%s' of RDE Type '%s': %s", behaviorId, rdeTypeId, err)
	}

	dSet(d, "name", behavior.Name)
	dSet(d, "ref", behavior.Ref)
	dSet(d, "description", behavior.Description)
	err = d.Set("execution", behavior.Execution)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(behavior.ID)

	return nil
}

func resourceVcdRdeTypeBehaviorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdRdeTypeBehaviorCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcdRdeTypeBehaviorDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("[RDE Type Behavior delete] could not read the Behavior of RDE Type with ID '%s': %s", rdeTypeId, err)
	}
	err = rdeType.DeleteBehaviorOverride(d.Id())
	if err != nil {
		return diag.Errorf("[RDE Type Behavior delete] could not delete the Behavior '%s' of RDE Type with ID '%s': %s", d.Id(), rdeTypeId, err)
	}
	return nil
}

// resourceVcdRdeTypeBehaviorImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde_type_behavior.behavior1
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0.myBehavior
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeTypeBehaviorImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	log.Printf("[DEBUG] importing vcd_rde_type_behavior resource with provided id %s", d.Id())
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	var rdeType *govcd.DefinedEntityType
	var behaviorName string
	var err error
	switch len(resourceURI) {
	case 4: // ie: VCD_IMPORT_SEPARATOR="_" vendor_nss_1.2.3_name
		rdeType, err = vcdClient.GetRdeType(resourceURI[0], resourceURI[1], resourceURI[2])
		behaviorName = resourceURI[3]
	case 6: // ie: vendor.nss.1.2.3.name
		rdeType, err = vcdClient.GetRdeType(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]))
		behaviorName = resourceURI[5]
	default:
		return nil, fmt.Errorf("the import ID should be specified like 'rdeTypeVendor.rdeTypeNss.rdeTypeVersion.behaviorName'")
	}
	if err != nil {
		return nil, fmt.Errorf("could not find any RDE Type with the provided ID '%s': %s", d.Id(), err)
	}

	behavior, err := rdeType.GetBehaviorByName(behaviorName)
	if err != nil {
		return nil, fmt.Errorf("could not find any Behavior with the name '%s' from the given ID '%s': %s", behaviorName, d.Id(), err)
	}

	d.SetId(behavior.ID)
	dSet(d, "rde_type_id", rdeType.DefinedEntityType.ID)
	dSet(d, "rde_interface_behavior_id", behavior.Ref) // Ref contains the original Interface Behavior ID
	return []*schema.ResourceData{d}, nil
}
