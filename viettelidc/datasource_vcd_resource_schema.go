package viettelidc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdResourceSchema() *schema.Resource {
	Attribute := schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "Attributes of the resource",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "attribute name",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "attribute type",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "attribute description",
				},
				"required": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Is the attribute required",
				},
				"computed": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Is the attribute computed",
				},
				"optional": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Is the attribute optional",
				},
				"sensitive": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Is the attribute sensitive",
				},
			},
		},
	}
	return &schema.Resource{
		ReadContext: datasourceVcdResourceSchemaRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name of the structure",
			},
			"resource_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Which resource we should list",
			},
			"attributes": &Attribute,
			"block_attributes": {
				Computed: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Block",
						},
						"nesting_mode": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "How the block is nested",
						},
						"attributes": &Attribute,
					},
				},
				Description: "The compound attributes for this resource",
			},
		},
	}
}

func datasourceVcdResourceSchemaRead(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {

	resourceType := d.Get("resource_type").(string)

	d.SetId(d.Get("name").(string))

	resource, ok := globalResourceMap[resourceType]
	if !ok {
		return diag.FromErr(fmt.Errorf("unhandled resource %s", resourceType))
	}

	attr := resource.CoreConfigSchema().Attributes
	block := resource.CoreConfigSchema().BlockTypes
	var data []map[string]interface{}
	var blockData []map[string]interface{}
	for name, a := range attr {
		var elem = map[string]interface{}{
			"name":        name,
			"type":        a.Type.FriendlyName(),
			"description": a.Description,
			"required":    a.Required,
			"optional":    a.Optional,
			"computed":    a.Computed,
			"sensitive":   a.Sensitive,
		}
		data = append(data, elem)
	}

	// A block element is a nested structure containing information about compound types.
	for name, b := range block {

		var mapElem = map[string]interface{}{
			"name":         name,
			"nesting_mode": b.Nesting.String(),
		}
		var blockAttributes []map[string]interface{}
		bAttr := b.Attributes
		if len(bAttr) > 0 {
			for b1Name, aa := range bAttr {
				var aElem = map[string]interface{}{
					"name":        b1Name,
					"type":        aa.Type.FriendlyName(),
					"description": aa.Description,
					"required":    aa.Required,
					"optional":    aa.Optional,
					"computed":    aa.Computed,
					"sensitive":   aa.Sensitive,
				}
				blockAttributes = append(blockAttributes, aElem)
			}
		} else {
			for aName, b1 := range b.BlockTypes {
				for b1Name, aa := range b1.Attributes {
					var aElem = map[string]interface{}{
						"name":        aName + " " + b1Name,
						"type":        aa.Type.FriendlyName(),
						"description": aa.Description,
						"required":    aa.Required,
						"optional":    aa.Optional,
						"computed":    aa.Computed,
						"sensitive":   aa.Sensitive,
					}
					blockAttributes = append(blockAttributes, aElem)
				}
			}
		}
		mapElem["attributes"] = blockAttributes
		blockData = append(blockData, mapElem)
	}
	if len(blockData) > 0 {
		err := d.Set("block_attributes", blockData)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err := d.Set("attributes", data)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}
