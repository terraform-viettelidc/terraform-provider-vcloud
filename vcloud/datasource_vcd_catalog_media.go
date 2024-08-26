package vcloud

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdCatalogMedia() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVcdMediaRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "catalog name where upload the Media file",
				Deprecated:   "use catalog_id instead, especially if using a shared catalog",
				ExactlyOneOf: []string{"catalog", "catalog_id"},
			},
			"catalog_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the catalog from where to retrieve the Media item",
				ExactlyOneOf: []string{"catalog", "catalog_id"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "media name (Optional when 'filter' is used)",
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog item metadata",
				Deprecated:  "Use metadata_entry instead",
			},
			"metadata_entry": metadataEntryDatasourceSchema("Media"),
			"is_iso": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is ISO",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name",
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is in a published catalog",
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Media storage in Bytes",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media status",
			},
			"storage_profile_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile name",
			},
			"download_to_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Will download the contents of the media item into the given file",
			},
			"filter": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a catalog media by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"date":       elementDate,
						"earliest":   elementEarliest,
						"latest":     elementLatest,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func dataSourceVcdMediaRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdMediaRead(d, meta, "datasource")
}
