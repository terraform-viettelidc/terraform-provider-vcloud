package vcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdSolutionAddonInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionAddonInstanceCreate,
		ReadContext:   resourceVcdSolutionAddonInstanceRead,
		UpdateContext: resourceVcdSolutionAddonInstanceUpdate,
		DeleteContext: resourceVcdSolutionAddonInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionAddonInstanceImport,
		},

		Schema: map[string]*schema.Schema{
			"add_on_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Solution Add-On ID",
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "Defines if EULA is accepted. `false` will return an error and print EULA",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Solution Add-On Instance Name",
			},
			"validate_only_required_inputs": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Defines if all or only required inputs should be validated",
			},
			"input": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "Key value map of Solution Add-On Instance",
			},
			"delete_input": { // These will only be applicable to "delete" operation
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key value map for deletion of Solution Add-On Instance",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Description: "Parent RDE state",
				Computed:    true,
			},
		},
	}
}

func resourceVcdSolutionAddonInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOn, err := vcdClient.GetSolutionAddonById(d.Get("add_on_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	if addOn.SolutionAddOnEntity.Eula != "" && !d.Get("accept_eula").(bool) {
		return diag.Errorf("cannot create Solution Add-On Instance without accepting EULA.\n\n%s\n\n: %s", addOn.SolutionAddOnEntity.Eula, err)
	}

	// making a copy of `input` map because mutating it causes terraform state errors
	input := d.Get("input")
	inputMap := input.(map[string]interface{})
	inputCopy := make(map[string]interface{})

	for k, v := range inputMap {
		// keys for all user inputs must be prefixed with `input-` for keys, however they are
		// defined without this prefix in Add-On schema itself.
		inputCopy[fmt.Sprintf("input-%s", k)] = v
	}

	// injecting "name" field that does not fall under regular inputs with "input-" prefix
	inputCopy["name"] = d.Get("name").(string)

	// Solution Add-On schema has typed fields and they must be converted to particular types based
	// on definition of schema as supplying boolean value as string will cause failure on
	// instantiation.
	convertedInputs, err := addOn.ConvertInputTypes(inputCopy)
	if err != nil {
		return diag.Errorf("error checking field types: %s", err)
	}

	// Validation will print field information and missing fields as described in the Solution
	// Add-On manifest. Due to RDEs being very sensitive to input - user has to provide all field
	// values instead of only required ones in the schema.
	validateOnlyRequiredInputs := d.Get("validate_only_required_inputs").(bool)
	err = addOn.ValidateInputs(convertedInputs, validateOnlyRequiredInputs, false)
	if err != nil {
		return diag.Errorf("dynamic creation input field validation error: %s", err)
	}

	addOnInstance, _, err := addOn.CreateSolutionAddOnInstance(convertedInputs)
	if err != nil {
		return diag.Errorf("error creating Solution Add-On ('%s') Instance: %s",
			addOn.DefinedEntity.DefinedEntity.Name, err)
	}

	d.SetId(addOnInstance.RdeId())

	return resourceVcdSolutionAddonInstanceRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// delete_input are only effective for deletion time therefore they must be updateable.
	// In reality this is a noop, but Terraform has to
	// validate_only_required_inputs might need to be changed on demand
	if !d.HasChangesExcept("delete_input", "validate_only_required_inputs") {
		return nil
	}

	// There is no real update, but an update function must be present so that user can change
	// 'delete_input' for deletion. They are not set in the resource, but delete uses them.

	return resourceVcdSolutionAddonInstanceRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance by ID: %s", err)
	}

	dSet(d, "add_on_id", addOnInstance.SolutionAddOnInstance.Prototype)
	dSet(d, "name", addOnInstance.SolutionAddOnInstance.Name)
	dSet(d, "rde_state", addOnInstance.DefinedEntity.State())

	// an existing Solution Add-On Instance cannot exist without accepting EULA
	dSet(d, "accept_eula", true)

	// Retrieve creation input fields
	// 'delete_input' values cannot be read from Solution Add-On Instance as they are specified only
	// when deleting the Add-On Instance.
	inputValues, err := addOnInstance.ReadCreationInputValues(true)
	if err != nil {
		return diag.Errorf("error reading Input values from Solution Add-On Instance: %s", err)
	}

	err = d.Set("input", inputValues)
	if err != nil {
		return diag.Errorf("error storing 'input' values: %s", err)
	}

	return nil
}

func resourceVcdSolutionAddonInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	addOn, err := vcdClient.GetSolutionAddonById(d.Get("add_on_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance by ID: %s", err)
	}

	deleteInput := d.Get("delete_input").(map[string]interface{})
	deleteCopy := make(map[string]interface{})

	for k, v := range deleteInput {
		// keys for all user inputs must be prefixed with `input-` for keys, however they are
		// defined without this prefix in Add-On schema itself.
		deleteCopy[fmt.Sprintf("input-%s", k)] = v
	}

	// injecting "name" field that does not fall under regular inputs
	deleteCopy["name"] = d.Get("name").(string)

	// Solution Add-On schema has types fields and they must be converted to particular types based
	// on definition of schema
	convertedInputs, err := addOn.ConvertInputTypes(deleteCopy)
	if err != nil {
		return diag.Errorf("error checking field types: %s", err)
	}

	// Validation will print field information and missing fields as described in the Add-On
	// manifest. Due to RDEs being very sensitive to input - user has to provide all field values
	// instead of only required ones in the schema.
	validateOnlyRequiredInputs := d.Get("validate_only_required_inputs").(bool)
	err = addOn.ValidateInputs(convertedInputs, validateOnlyRequiredInputs, true)
	if err != nil {
		return diag.Errorf("dynamic deletion field validation error: %s", err)
	}

	_, err = addOnInstance.Delete(convertedInputs)
	if err != nil {
		return diag.Errorf("error removing Solution Add-On Instance: %s", err)
	}

	return nil

}

func resourceVcdSolutionAddonInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	addOnInstance, err := vcdClient.GetSolutionAddonInstanceByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error finding Solution Add-On Instance by Name '%s': %s", d.Id(), err)
	}

	d.SetId(addOnInstance.RdeId())

	return []*schema.ResourceData{d}, nil
}
