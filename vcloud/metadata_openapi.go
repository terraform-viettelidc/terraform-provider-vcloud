package vcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"reflect"
	"strconv"
)

// openApiMetadataEntryDatasourceSchema returns the schema associated to the OpenAPI metadata_entry for a given data source.
// The description will refer to the object type given as input.
func openApiMetadataEntryDatasourceSchema(resourceType string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: fmt.Sprintf("Metadata entries from the given %s", resourceType),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ID of the metadata entry",
				},
				"key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Key of this metadata entry",
				},
				"value": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Value of this metadata entry",
				},
				"type": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s'", types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry),
				},
				"readonly": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "True if the metadata entry is read only",
				},
				"domain": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Only meaningful for providers. Allows them to share entries with their tenants. One of: `TENANT`, `PROVIDER`",
				},
				"namespace": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Namespace of the metadata entry",
				},
				"persistent": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: "Persistent metadata entries can be copied over on some entity operation",
				},
			},
		},
	}
}

// openApiMetadataEntryResourceSchema returns the schema associated to the OpenAPI metadata_entry for a given resource.
// The description will refer to the object name given as input.
func openApiMetadataEntryResourceSchema(resourceType string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: fmt.Sprintf("Metadata entries for the given %s", resourceType),
		MaxItems:    50, // As per the documentation
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "ID of the metadata entry",
				},
				"key": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Key of this metadata entry. Required if the metadata entry is not empty",
				},
				"value": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Value of this metadata entry. Required if the metadata entry is not empty",
				},
				"type": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      types.OpenApiMetadataStringEntry,
					Description:  fmt.Sprintf("Type of this metadata entry. One of: '%s', '%s', '%s'", types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry),
					ValidateFunc: validation.StringInSlice([]string{types.OpenApiMetadataStringEntry, types.OpenApiMetadataNumberEntry, types.OpenApiMetadataBooleanEntry}, false),
				},
				"readonly": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "True if the metadata entry is read only",
				},
				"domain": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "TENANT",
					Description:  "Only meaningful for providers. Allows them to share entries with their tenants. Currently, accepted values are: `TENANT`, `PROVIDER`",
					ValidateFunc: validation.StringInSlice([]string{"TENANT", "PROVIDER"}, false),
				},
				"namespace": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Namespace of the metadata entry",
				},
				"persistent": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Persistent metadata entries can be copied over on some entity operation",
				},
			},
		},
	}
}

// openApiMetadataCompatible allows to consider all structs that implement OpenAPI metadata handling to be the same type
type openApiMetadataCompatible interface {
	GetMetadata() ([]*govcd.OpenApiMetadataEntry, error)
	GetMetadataByKey(domain, namespace, key string) (*govcd.OpenApiMetadataEntry, error)
	GetMetadataById(id string) (*govcd.OpenApiMetadataEntry, error)
	AddMetadata(metadataEntry types.OpenApiMetadataEntry) (*govcd.OpenApiMetadataEntry, error)
}

// createOrUpdateOpenApiMetadataEntryInVcd creates or updates OpenAPI metadata entries in VCD for the given resource, only if the attribute
// metadata_entry has been set or updated in the state.
func createOrUpdateOpenApiMetadataEntryInVcd(d *schema.ResourceData, resource openApiMetadataCompatible) error {
	if !d.HasChange("metadata_entry") {
		return nil
	}

	oldRaw, newRaw := d.GetChange("metadata_entry")
	metadataToAdd, metadataToUpdate, metadataToDelete, err := getOpenApiMetadataOperations(oldRaw.(*schema.Set).List(), newRaw.(*schema.Set).List())
	if err != nil {
		return fmt.Errorf("could not calculate the needed metadata operations: %s", err)
	}

	for _, entry := range metadataToDelete {
		toDelete, err := resource.GetMetadataByKey(entry.KeyValue.Domain, entry.KeyValue.Namespace, entry.KeyValue.Key) // Refreshes ETags
		if err != nil {
			return fmt.Errorf("error reading metadata with namespace '%s' and key '%s': %s", entry.KeyValue.Namespace, entry.KeyValue.Key, err)
		}
		err = toDelete.Delete()
		if err != nil {
			return fmt.Errorf("error deleting metadata with namespace '%s' and key '%s': %s", entry.KeyValue.Namespace, entry.KeyValue.Key, err)
		}
	}

	for _, entry := range metadataToUpdate {
		toUpdate, err := resource.GetMetadataByKey(entry.KeyValue.Domain, entry.KeyValue.Namespace, entry.KeyValue.Key) // Refreshes ETags
		if err != nil {
			return fmt.Errorf("error reading metadata with namespace '%s' and key '%s': %s", entry.KeyValue.Namespace, entry.KeyValue.Key, err)
		}
		err = toUpdate.Update(entry.KeyValue.Value.Value, entry.IsPersistent)
		if err != nil {
			return fmt.Errorf("error updating metadata with namespace '%s' and key '%s': %s", entry.KeyValue.Namespace, entry.KeyValue.Key, err)
		}
	}

	for _, metadataEntry := range metadataToAdd {
		_, err := resource.AddMetadata(metadataEntry)
		if err != nil {
			return fmt.Errorf("error adding metadata entry: %s", err)
		}
	}
	return nil
}

// getOpenApiMetadataOperations retrieves the metadata that needs to be added, to be updated and to be deleted depending
// on the old and new attribute values from Terraform state.
func getOpenApiMetadataOperations(oldMetadata []interface{}, newMetadata []interface{}) ([]types.OpenApiMetadataEntry, []types.OpenApiMetadataEntry, []types.OpenApiMetadataEntry, error) {
	oldMetadataEntries, err := getOpenApiMetadataEntryMap(oldMetadata)
	if err != nil {
		return nil, nil, nil, err
	}
	newMetadataEntries, err := getOpenApiMetadataEntryMap(newMetadata)
	if err != nil {
		return nil, nil, nil, err
	}

	var metadataToRemove []types.OpenApiMetadataEntry
	for oldNamespacedKey := range oldMetadataEntries {
		if _, ok := newMetadataEntries[oldNamespacedKey]; !ok {
			metadataToRemove = append(metadataToRemove, oldMetadataEntries[oldNamespacedKey])
		}
	}

	var metadataToCreate []types.OpenApiMetadataEntry
	metadataToUpdateMap := map[string]types.OpenApiMetadataEntry{}
	for newNamespacedKey, newEntry := range newMetadataEntries {
		if oldEntry, ok := oldMetadataEntries[newNamespacedKey]; ok {
			if reflect.DeepEqual(oldEntry, newEntry) {
				continue
			}
			// If a metadata property that is not "Value" or "IsPersistent" is changed, it needs to be recreated
			if oldEntry.IsReadOnly != newEntry.IsReadOnly || oldEntry.KeyValue.Namespace != newEntry.KeyValue.Namespace ||
				oldEntry.KeyValue.Domain != newEntry.KeyValue.Domain || oldEntry.KeyValue.Value.Type != newEntry.KeyValue.Value.Type {
				util.Logger.Printf("[DEBUG] entry with namespace '%s' and key '%s' is being deleted and re-created", oldEntry.KeyValue.Namespace, oldEntry.KeyValue.Key)
				metadataToRemove = append(metadataToRemove, oldMetadataEntries[newNamespacedKey])
				metadataToCreate = append(metadataToCreate, newMetadataEntries[newNamespacedKey])
			} else {
				// Only "Value" / "IsPersistent" is changed, it can be updated
				metadataToUpdateMap[newNamespacedKey] = newEntry
			}

		}
	}
	var metadataToUpdate []types.OpenApiMetadataEntry
	for _, v := range metadataToUpdateMap {
		metadataToUpdate = append(metadataToUpdate, v)
	}

	for newNamespacedKey, newEntry := range newMetadataEntries {
		_, alreadyExisting := oldMetadataEntries[newNamespacedKey]
		_, beingUpdated := metadataToUpdateMap[newNamespacedKey]
		if !alreadyExisting && !beingUpdated {
			metadataToCreate = append(metadataToCreate, newEntry)
		}
	}

	return metadataToCreate, metadataToUpdate, metadataToRemove, nil
}

// getOpenApiMetadataEntryMap converts the input metadata attribute from Terraform state to a map composed by metadata
// namespaced keys (this is, namespace and key separated by '%%%') and their values.
func getOpenApiMetadataEntryMap(metadataAttribute []interface{}) (map[string]types.OpenApiMetadataEntry, error) {
	metadataMap := map[string]types.OpenApiMetadataEntry{}
	for _, rawItem := range metadataAttribute {
		metadataEntry := rawItem.(map[string]interface{})

		namespace := ""
		if _, ok := metadataEntry["namespace"]; ok {
			namespace = metadataEntry["namespace"].(string)
		}

		value, err := convertOpenApiMetadataValue(metadataEntry["type"].(string), metadataEntry["value"].(string))
		if err != nil {
			return nil, fmt.Errorf("error parsing the 'value' attribute '%s' from state: %s", metadataEntry["value"].(string), err)
		}

		// In OpenAPI, metadata is namespaced, hence it is possible to have same keys but in different namespaces.
		// For that reason, we use "namespace+key" to unequivocally identify the metadata entries.
		namespacedKey := fmt.Sprintf("%s%s", namespace, metadataEntry["key"].(string))
		if _, ok := metadataMap[namespacedKey]; ok {
			return nil, fmt.Errorf("metadata entry with namespace '%s' and key '%s' already exists", namespace, metadataEntry["key"])
		}

		metadataMap[namespacedKey] = types.OpenApiMetadataEntry{
			IsReadOnly:   metadataEntry["readonly"].(bool),   // It is always populated as it has a default value
			IsPersistent: metadataEntry["persistent"].(bool), // It is always populated as it has a default value
			KeyValue: types.OpenApiMetadataKeyValue{
				Domain: metadataEntry["domain"].(string), // It is always populated as it has a default value
				Key:    metadataEntry["key"].(string),    // It is always populated as it is required
				Value: types.OpenApiMetadataTypedValue{
					Value: value,
					Type:  metadataEntry["type"].(string), // It is always populated as it has a default value
				},
				Namespace: namespace,
			},
		}
	}
	return metadataMap, nil
}

// updateOpenApiMetadataInState updates metadata_entry in the Terraform state for the given receiver object.
// This can be done as both are Computed, for compatibility reasons.
func updateOpenApiMetadataInState(d *schema.ResourceData, vcdClient *VCDClient, resourceType string, receiverObject openApiMetadataCompatible) diag.Diagnostics {
	diags := checkIgnoredMetadataConflicts(d, vcdClient, resourceType)
	if diags != nil && diags.HasError() {
		return diags
	}

	allMetadata, err := receiverObject.GetMetadata()
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	metadata := make([]interface{}, len(allMetadata))
	for i, metadataEntryFromVcd := range allMetadata {
		// We need to set the correct type, otherwise saving the state will fail
		value := ""
		switch metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Type {
		case types.OpenApiMetadataBooleanEntry:
			value = fmt.Sprintf("%t", metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Value.(bool))
		case types.OpenApiMetadataNumberEntry:
			value = fmt.Sprintf("%.0f", metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Value.(float64))
		case types.OpenApiMetadataStringEntry:
			value = metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Value.(string)
		default:
			return append(diags, diag.Errorf("not supported metadata type %s", metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Type)...)
		}

		metadataEntry := map[string]interface{}{
			"id":         metadataEntryFromVcd.MetadataEntry.ID,
			"key":        metadataEntryFromVcd.MetadataEntry.KeyValue.Key,
			"readonly":   metadataEntryFromVcd.MetadataEntry.IsReadOnly,
			"domain":     metadataEntryFromVcd.MetadataEntry.KeyValue.Domain,
			"namespace":  metadataEntryFromVcd.MetadataEntry.KeyValue.Namespace,
			"type":       metadataEntryFromVcd.MetadataEntry.KeyValue.Value.Type,
			"value":      value,
			"persistent": metadataEntryFromVcd.MetadataEntry.IsPersistent,
		}
		metadata[i] = metadataEntry
	}

	err = d.Set("metadata_entry", metadata)
	return append(diags, diag.FromErr(err)...)
}

// convertOpenApiMetadataValue converts a metadata value from plain string to a correct typed value that can be sent
// in OpenAPI payloads.
func convertOpenApiMetadataValue(valueType, value string) (interface{}, error) {
	var convertedValue interface{}
	var err error
	switch valueType {
	case types.OpenApiMetadataStringEntry:
		convertedValue = value
	case types.OpenApiMetadataNumberEntry:
		convertedValue, err = strconv.ParseFloat(value, 64)
	case types.OpenApiMetadataBooleanEntry:
		convertedValue, err = strconv.ParseBool(value)
	default:
		return nil, fmt.Errorf("unrecognized metadata type %s", valueType)
	}
	if err != nil {
		return nil, fmt.Errorf("error parsing the value '%v': %s", value, err)
	}
	return convertedValue, nil
}
