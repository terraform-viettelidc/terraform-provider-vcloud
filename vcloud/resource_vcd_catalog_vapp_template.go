package vcloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

const maximumSynchronisationCheckDuration = 60 * time.Second

func resourceVcdCatalogVappTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdCatalogVappTemplateCreate,
		ReadContext:   resourceVcdCatalogVappTemplateRead,
		UpdateContext: resourceVcdCatalogVappTemplateUpdate,
		DeleteContext: resourceVcdCatalogVappTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdCatalogVappTemplateImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Catalog where to upload the OVA file",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the VDC to which the vApp Template belongs",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vApp Template name",
			},
			"capture_vapp": {
				Optional:      true,
				Type:          schema.TypeList,
				MaxItems:      1,
				Description:   "Provides configuration options for creating a vApp Template from existing vApp",
				ConflictsWith: []string{"ovf_url", "ova_path"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_id": {
							Optional:    true,
							Type:        schema.TypeString,
							Description: "Source vApp ID (can be a vApp ID or 'vapp_id' field of standalone VM 'vcd_vm')",
						},
						"overwrite_catalog_item_id": {
							Optional:    true,
							Type:        schema.TypeString,
							Description: "An existing catalog item ID to overwrite",
						},
						"customize_on_instantiate": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Marks if instantiating applies customization settings ('true'). Default is 'false` - create an identical copy.",
						},
						"copy_tpm_on_instantiate": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Defines if Trusted Platform Module should be copied (false) or created (true). Default 'false'. VCD 10.4.2+",
						},
					},
				},
			},
			"description": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true, // Due to a bug in VCD when using `ovf_url`, `description` is overridden by the target OVA's description.
				Description:   "Description of the vApp Template. Not to be used with `ovf_url` when target OVA has a description",
				ConflictsWith: []string{"ovf_url"}, // This is to avoid the bug mentioned above.
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of when the vApp Template was created",
			},
			"catalog_item_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog Item ID of this vApp template",
			},
			"vm_names": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "Set of VM names within the vApp template",
			},
			"ova_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "Absolute or relative path to OVA",
				ConflictsWith: []string{"ovf_url", "capture_vapp"},
			},
			"ovf_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"description", "ova_path", "capture_vapp"}, // This is to avoid the bug mentioned above.
				Description:   "URL of OVF file",
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Size of upload file piece size in megabytes",
			},
			"lease": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Defines lease parameters for this vApp template",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_lease_in_sec": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "How long the vApp template is available before being automatically deleted or marked as expired. 0 means never expires (or expires at the maximum limit provided by the parent Org)",
							ValidateFunc: validateIntLeaseSeconds(), // Lease can be either 0 or 3600+
						},
					},
				},
			},
			"metadata": {
				Type:          schema.TypeMap,
				Optional:      true,
				Computed:      true, // To be compatible with `metadata_entry`
				Description:   "Key and value pairs for the metadata of this vApp Template",
				Deprecated:    "Use metadata_entry instead",
				ConflictsWith: []string{"metadata_entry"},
			},
			"metadata_entry": metadataEntryResourceSchemaDeprecated("vApp Template"),
			"inherited_metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "A map that contains metadata that is automatically added by VCD (10.5.1+) and provides details on the origin of the VM",
			},
		},
	}
}

func resourceVcdCatalogVappTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	catalogId := d.Get("catalog_id").(string)
	catalog, err := vcdClient.Client.GetCatalogById(catalogId)
	if err != nil {
		log.Printf("[DEBUG] Error finding Catalog: %s", err)
		return diag.Errorf("error finding Catalog: %s", err)
	}

	var diagError diag.Diagnostics

	ovaPath := d.Get("ova_path").(string)
	ovfUrl := d.Get("ovf_url").(string)
	vappTemplateName := d.Get("name").(string)

	capturevAppTemplate := d.Get("capture_vapp").([]interface{})

	switch {
	case ovaPath != "":
		diagError = uploadOvaFromFilePath(d, catalog, vappTemplateName, "vcd_catalog_vapp_template")
	case ovfUrl != "":
		diagError = uploadFromUrl(d, catalog, vappTemplateName, "vcd_catalog_vapp_template")
	case len(capturevAppTemplate) == 1:
		templateCaptureSettings := capturevAppTemplate[0].(map[string]interface{})
		sourceId := templateCaptureSettings["source_id"].(string)
		overwriteCatalogItemId := templateCaptureSettings["overwrite_catalog_item_id"].(string)
		customizeOnInstantiate := templateCaptureSettings["customize_on_instantiate"].(bool)

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return diag.FromErr(err)
		}

		url := vcdClient.Client.VCDHREF
		vAppHref := fmt.Sprintf("%s://%s/api/vApp/vapp-%s", url.Scheme, url.Host, extractUuid(sourceId))
		vapp, err := org.GetVAppByHref(vAppHref)
		if err != nil {
			return diag.FromErr(err)
		}

		vAppCaptureParams := &types.CaptureVAppParams{
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Source: &types.Reference{
				HREF: vAppHref,
			},
			CustomizationSection: types.CaptureVAppParamsCustomizationSection{
				Info:                   "CustomizeOnInstantiate Settings",
				CustomizeOnInstantiate: customizeOnInstantiate,
			},
		}

		// It is possible to overwrite an existing item
		if overwriteCatalogItemId != "" {
			catalogItemUuid := extractUuid(overwriteCatalogItemId)
			if !govcd.IsUuid(catalogItemUuid) {
				return diag.Errorf("expected Catalog Item ID to contain UUID, got: %s", overwriteCatalogItemId)
			}

			overWriteItemHref := fmt.Sprintf("%s://%s/api/catalogItem/%s", url.Scheme, url.Host, catalogItemUuid)
			vAppCaptureParams.TargetCatalogItem = &types.Reference{
				HREF: overWriteItemHref,
			}
		}

		// TPM (Trusted Platform Module) setting 'CopyTpmOnInstantiate' is only available in VCD 10.4.2+
		copyTpmOnInstantiate := templateCaptureSettings["copy_tpm_on_instantiate"].(bool)
		if vcdClient.Client.APIVCDMaxVersionIs(">= 37.2") {
			vAppCaptureParams.CopyTpmOnInstantiate = &copyTpmOnInstantiate
		} else if copyTpmOnInstantiate { // Throw an error if user set to `true`, but VCD is too old
			return diag.Errorf("'copy_tpm_on_instantiate' is supported on VCD 10.4.2+")
		}

		parentVdc, err := vapp.GetParentVDC()
		if err != nil {
			return diag.Errorf("error retrieving parent VDC for vApp %s: %s", vapp.VApp.Name, err)
		}

		// Locking vApp as it becomes busy when an image is being created
		unlock := vcdClient.lockVappWithName(org.Org.Name, parentVdc.Vdc.Name, vapp.VApp.Name)
		defer unlock()

		createdTemplate, err := catalog.CaptureVappTemplate(vAppCaptureParams)
		if err != nil {
			return diag.FromErr(err)
		}

		// Explicitly rename created template to what is specified in `name` field because by
		// default it will have the name of overwritten template
		createdTemplate.VAppTemplate.Name = d.Get("name").(string)
		updatedTemplate, err := createdTemplate.Update()
		if err != nil {
			return diag.Errorf("error renaming template after overwrite: %s", err)
		}

		vappTemplateName = updatedTemplate.VAppTemplate.Name
	default:
		diagError = diag.Errorf("`ova_path`, `ovf_url` or `capture_vapp` must be set: %s", err)

	}

	if diagError != nil {
		return diagError
	}

	vAppTemplate, err := catalog.GetVAppTemplateByName(vappTemplateName)
	if err != nil {
		return diag.Errorf("error retrieving vApp Template %s: %s", vappTemplateName, err)
	}

	err = createOrUpdateMetadata(d, vAppTemplate, "metadata")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)
	log.Printf("[TRACE] Catalog vApp Template created: %s", vappTemplateName)

	err = vappTemplateLeaseUpdate(vcdClient, vAppTemplate, d)
	if err != nil {
		return diag.Errorf("error updating VApp template lease terms: %s", err)
	}
	return resourceVcdCatalogVappTemplateRead(ctx, d, meta)
}

func resourceVcdCatalogVappTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogVappTemplateRead(ctx, d, meta, "resource")
}

// genericVcdCatalogVappTemplateRead performs a Read operation for the vApp Template resource (origin="resource")
// and data source (origin="datasource").
func genericVcdCatalogVappTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	var diags diag.Diagnostics
	vcdClient := meta.(*VCDClient)
	vAppTemplate, err := findVAppTemplate(d, vcdClient, origin)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template: %s", err)
		return diag.Errorf("Unable to find vApp Template: %s", err)
	}

	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	dSet(d, "created", vAppTemplate.VAppTemplate.DateCreated)
	dSet(d, "description", vAppTemplate.VAppTemplate.Description)

	_, isCatalogIdSet := d.GetOk("catalog_id")
	if !isCatalogIdSet { // This can only happen in the data source.
		catalogName, err := vAppTemplate.GetCatalogName()
		if err != nil {
			return diag.Errorf("error retrieving the Catalog name to which the vApp Template '%s' belongs: %s", vAppTemplate.VAppTemplate.Name, err)
		}

		orgName, err := vcdClient.GetOrgNameFromResource(d)
		if err != nil {
			return diag.Errorf("no org name found in resource data: %s", err)
		}
		catalog, err := vcdClient.Client.GetCatalogByName(orgName, catalogName)
		if err != nil {
			return diag.Errorf("error retrieving Catalog from vApp Template with name %s: %s", vAppTemplate.VAppTemplate.Name, err)
		}
		dSet(d, "catalog_id", catalog.Catalog.ID)
	} else {
		vappTemplateRec, err := vAppTemplate.GetVappTemplateRecord()
		if err != nil {
			return diag.Errorf("error retrieving the vApp Template record for '%s': %s", vAppTemplate.VAppTemplate.Name, err)
		}
		dSet(d, "vdc_id", "urn:vcloud:vdc:"+extractUuid(vappTemplateRec.Vdc))
	}

	var vmNames []string
	if vAppTemplate.VAppTemplate.Children != nil {
		for _, vm := range vAppTemplate.VAppTemplate.Children.VM {
			vmNames = append(vmNames, vm.Name)
		}
	}
	err = d.Set("vm_names", vmNames)
	if err != nil {
		diag.Errorf("Unable to set attribute 'vm_names' for the vApp Template: %s", err)
	}
	leaseInfo, err := vAppTemplate.GetLease()
	if err != nil {
		return diag.Errorf("unable to get lease information: %s", err)
	}
	leaseData := []map[string]interface{}{
		{
			"storage_lease_in_sec": leaseInfo.StorageLeaseInSeconds,
		},
	}
	err = d.Set("lease", leaseData)
	if err != nil {
		return diag.Errorf("unable to set lease information in state: %s", err)
	}

	d.SetId(vAppTemplate.VAppTemplate.ID)

	catalogItemId, err := runWithRetry("retrieving catalog item ID",
		"error retrieving Catalog Item ID for vApp template",
		10*time.Second,
		func() error {
			return vAppTemplate.Refresh()
		},
		func() (any, error) {
			return vAppTemplate.GetCatalogItemId()
		})
	if err != nil {
		return diag.Errorf("error retrieving Catalog Item ID for vApp template: %s", err)
	}
	dSet(d, "catalog_item_id", catalogItemId)

	diags = append(diags, updateMetadataInStateDeprecated(d, vcdClient, "vcd_catalog_vapp_template", vAppTemplate)...)
	if diags != nil && diags.HasError() {
		return diags
	}

	// This must be checked at the end as updateMetadataInStateDeprecated can throw Warning diagnostics
	if len(diags) > 0 {
		return diags
	}
	return nil
}

func resourceVcdCatalogVappTemplateUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vAppTemplate, err := findVAppTemplate(d, vcdClient, "resource")

	if d.HasChange("description") || d.HasChange("name") {
		if err != nil {
			return diag.Errorf("Unable to find vApp Template: %s", err)
		}

		vAppTemplate.VAppTemplate.Description = d.Get("description").(string)
		vAppTemplate.VAppTemplate.Name = d.Get("name").(string)
		_, err = vAppTemplate.Update()
		if err != nil {
			return diag.Errorf("error updating vApp Template: %s", err)
		}
	}

	err = vappTemplateLeaseUpdate(vcdClient, vAppTemplate, d)
	if err != nil {
		return diag.Errorf("error updating VApp template lease terms: %s", err)
	}
	err = createOrUpdateMetadata(d, vAppTemplate, "metadata")
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdCatalogVappTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	catalogId := d.Get("catalog_id").(string)
	catalog, err := vcdClient.Client.GetCatalogById(catalogId)
	if err != nil {
		log.Printf("[DEBUG] Unable to find Catalog with ID %s", catalogId)
		return diag.Errorf("unable to find Catalog with ID %s", catalogId)
	}

	vAppTemplateName := d.Get("name").(string)
	vAppTemplate, err := catalog.GetVAppTemplateByName(vAppTemplateName)
	if err != nil {
		log.Printf("[DEBUG] Unable to find vApp Template with name %s", vAppTemplateName)
		return diag.Errorf("unable to find vApp Template with name %s", vAppTemplateName)
	}

	err = vAppTemplate.Delete()
	if err != nil {
		log.Printf("[DEBUG] Error removing vApp Template %s", err)
		return diag.Errorf("error removing vApp Template %s", err)
	}

	_, err = catalog.GetVAppTemplateByName(vAppTemplateName)
	if err == nil {
		return diag.Errorf("vApp Template %s still found after deletion", vAppTemplateName)
	}
	log.Printf("[TRACE] vApp Template delete completed: %s", vAppTemplateName)

	return nil
}

func vappTemplateLeaseUpdate(vcdClient *VCDClient, vAppTemplate *govcd.VAppTemplate, d *schema.ResourceData) error {
	var storageLease = vAppTemplate.VAppTemplate.LeaseSettingsSection.StorageLeaseInSeconds
	rawLeaseSection1, ok := d.GetOk("lease")
	if ok {
		// We have a lease block
		rawLeaseSection2 := rawLeaseSection1.([]interface{})
		leaseSection := rawLeaseSection2[0].(map[string]interface{})
		storageLease = leaseSection["storage_lease_in_sec"].(int)
	}

	if storageLease != vAppTemplate.VAppTemplate.LeaseSettingsSection.StorageLeaseInSeconds {
		err := vAppTemplate.RenewLease(storageLease)
		if err != nil {
			return fmt.Errorf("error updating VApp template lease terms: %s", err)
		}
	}
	return nil
}

// Imports a vApp Template into Terraform state
// This function task is to get the data from VCD and fill the resource data container
// Expects the d.ID() to be a path to the resource by using a Catalog like org_name.catalog_name.vapp_template_name
// or a VDC like org_name.vdc_name.vapp_template_name
//
// Example import path (id): myOrg1.myCatalog2.myvAppTemplate3
// Example import path (id): myOrg1.myVdc2.myvAppTemplate3
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdCatalogVappTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org.catalog_name.vapp_template_name")
	}
	orgName, catalogOrVdcName, vAppTemplateName := resourceURI[0], resourceURI[1], resourceURI[2]

	if orgName == "" {
		return nil, fmt.Errorf("import: empty Org name provided")
	}
	if catalogOrVdcName == "" {
		return nil, fmt.Errorf("import: empty Catalog or VDC name provided")
	}
	if vAppTemplateName == "" {
		return nil, fmt.Errorf("import: empty vApp Template name provided")
	}

	vcdClient := meta.(*VCDClient)

	catalog, err := vcdClient.Client.GetCatalogByName(orgName, catalogOrVdcName)
	var vdc *govcd.Vdc
	if err != nil {
		org, err := vcdClient.GetOrgByName(orgName)
		if err != nil {
			return nil, fmt.Errorf(errorRetrievingOrg, orgName)
		}
		vdc, err = org.GetVDCByName(catalogOrVdcName, false)
		if err != nil {
			return nil, govcd.ErrorEntityNotFound
		}
	}
	var vAppTemplate *govcd.VAppTemplate
	if vdc != nil {
		vAppTemplate, err = vdc.GetVAppTemplateByName(vAppTemplateName)
		dSet(d, "vdc_id", vdc.Vdc.ID)
	} else {
		vAppTemplate, err = catalog.GetVAppTemplateByName(vAppTemplateName)
		dSet(d, "catalog_id", catalog.Catalog.ID)
	}
	if err != nil {
		return nil, govcd.ErrorEntityNotFound
	}

	dSet(d, "org", orgName)
	dSet(d, "name", vAppTemplate.VAppTemplate.Name)
	d.SetId(vAppTemplate.VAppTemplate.ID)

	return []*schema.ResourceData{d}, nil
}

// Finds a vApp Template with the information given in the resource data. If it's a data source it uses a filtering
// mechanism, if it's a resource it just gets the information.
func findVAppTemplate(d *schema.ResourceData, vcdClient *VCDClient, origin string) (*govcd.VAppTemplate, error) {
	log.Printf("[TRACE] vApp template search initiated")

	identifier := d.Id()
	// Check if identifier is still in deprecated style `catalogName:mediaName`
	if origin == "datasource" && identifier == "" || strings.Count(identifier, ":") <= 1 {
		identifier = d.Get("name").(string)
	}

	// Get the catalog only if its ID is set, as in data source we can search with VDC ID instead.
	var catalog *govcd.Catalog
	var err error
	var vdc *govcd.Vdc
	catalogId, isSearchedByCatalog := d.GetOk("catalog_id")
	if isSearchedByCatalog {
		catalog, err = vcdClient.Client.GetCatalogById(catalogId.(string))
		if err != nil {
			log.Printf("[DEBUG] Unable to find Catalog.")
			return nil, fmt.Errorf("unable to find Catalog: %s", err)
		}
	} else {
		// If we search the catalog by VDC, we assume access to the Org
		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return nil, fmt.Errorf(errorRetrievingOrg, err)
		}
		vdc, err = adminOrg.GetVDCById(d.Get("vdc_id").(string), false)
		if err != nil {
			log.Printf("[DEBUG] Unable to find VDC.")
			return nil, fmt.Errorf("unable to find VDC: %s", err)
		}
	}

	var vAppTemplate *govcd.VAppTemplate
	if origin == "datasource" {
		if !nameOrFilterIsSet(d) {
			return nil, fmt.Errorf(noNameOrFilterError, "vcd_catalog_vapp_template")
		}

		filter, hasFilter := d.GetOk("filter")

		if hasFilter {
			if isSearchedByCatalog {
				vAppTemplate, err = getVappTemplateByCatalogAndFilter(catalog, filter, vcdClient.Client.IsSysAdmin)
			} else {
				vAppTemplate, err = getVappTemplateByVdcAndFilter(vdc, filter, vcdClient.Client.IsSysAdmin)
			}
			// A race condition can happen between the getVAppTemplate call above and QuerySynchronizedVAppTemplateById below.
			// as we can have a vApp template that is not synchronized. The sync can happen by the time we
			// call QuerySynchronizedVAppTemplateById, but the ID will have changed, hence it will fail with a NotFoundError.
			if err != nil {
				return nil, err
			}
			err := checkSynchronisedVappTemplate(vcdClient, vAppTemplate)
			if err != nil {
				return nil, err
			}
			d.SetId(vAppTemplate.VAppTemplate.ID)
			return vAppTemplate, nil
		}
	}
	// No filter: we continue with single item  GET

	if isSearchedByCatalog {
		// In a resource, this is the only possibility
		vAppTemplate, err = catalog.GetVAppTemplateByNameOrId(identifier, false)
	} else {
		vAppTemplate, err = vdc.GetVAppTemplateByNameOrId(identifier, false)
	}
	// A race condition can happen between the GetVAppTemplate call above and QuerySynchronizedVAppTemplateById below.
	// as we can have a vApp template that is not synchronized. The sync can happen by the time we
	// call QuerySynchronizedVAppTemplateById, but the ID will have changed, hence it will fail with a NotFoundError.

	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] Unable to find vApp Template %s. Removing from tfstate", identifier)
		d.SetId("")
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to find vApp Template %s: %s", identifier, err)
	}
	if origin == "datasource" {
		// This checks that the vApp Template is synchronized in the catalog
		err := checkSynchronisedVappTemplate(vcdClient, vAppTemplate)
		if err != nil {
			return nil, err
		}
	}
	d.SetId(vAppTemplate.VAppTemplate.ID)
	log.Printf("[TRACE] vApp Template read completed: %#v", vAppTemplate.VAppTemplate)
	return vAppTemplate, nil
}

func checkSynchronisedVappTemplate(vcdClient *VCDClient, vAppTemplate *govcd.VAppTemplate) error {
	startCheck := time.Now()
	var err error
	timeout := false
	iterations := 0
	// This checks that the vApp Template is synchronized in the catalog
	for !timeout {
		iterations++
		_, err = vcdClient.QuerySynchronizedVAppTemplateById(vAppTemplate.VAppTemplate.ID)
		if err == nil {
			break
		}
		timeout = time.Since(startCheck) > maximumSynchronisationCheckDuration
		time.Sleep(500 * time.Millisecond)
	}
	util.Logger.Printf("[DEBUG] [checkSynchronisedVappTemplate] {timeout: %v - iterations %d} synchronisation query for %s completed after %s\n", timeout, iterations, vAppTemplate.VAppTemplate.Name, time.Since(startCheck))
	if err != nil {
		return fmt.Errorf("the found vApp Template %s (%s) is not usable: %s", vAppTemplate.VAppTemplate.Name, vAppTemplate.VAppTemplate.ID, err)
	}
	return nil
}

// uploadOvaFromFilePath uploads an OVA file specified in the resource to the given catalog
func uploadOvaFromFilePath(d *schema.ResourceData, catalog *govcd.Catalog, vappTemplate, resourceName string) diag.Diagnostics {
	uploadPieceSize := d.Get("upload_piece_size").(int)
	task, err := catalog.UploadOvf(d.Get("ova_path").(string), vappTemplate, d.Get("description").(string), int64(uploadPieceSize)*1024*1024) // Convert from megabytes to bytes
	if err != nil {
		log.Printf("[DEBUG] Error uploading file: %s", err)
		return diag.Errorf("error uploading file: %s", err)
	}

	return finishHandlingTask(d, *task.Task, vappTemplate, resourceName)
}

func uploadFromUrl(d *schema.ResourceData, catalog *govcd.Catalog, itemName, resourceName string) diag.Diagnostics {
	task, err := catalog.UploadOvfByLink(d.Get("ovf_url").(string), itemName, d.Get("description").(string))
	if err != nil {
		log.Printf("[DEBUG] Error uploading OVF from URL: %s", err)
		return diag.Errorf("error uploading OVF from URL: %s", err)
	}

	return finishHandlingTask(d, task, itemName, resourceName)
}

func finishHandlingTask(d *schema.ResourceData, task govcd.Task, itemName string, resourceName string) diag.Diagnostics {
	// This is a deprecated feature from vcd_catalog_item, to be removed with vcd_catalog_item
	if resourceName == "vcd_catalog_item" && d.Get("show_upload_progress").(bool) {
		for {
			progress, err := task.GetTaskProgress()
			if err != nil {
				log.Printf("VCD Error importing new catalog item: %s", err)
				return diag.Errorf("VCD Error importing new catalog item: %s", err)
			}
			logForScreen("vcd_catalog_item", fmt.Sprintf("vcd_catalog_item."+itemName+": VCD import catalog item progress "+progress+"%%\n"))
			if progress == "100" {
				break
			}
			time.Sleep(10 * time.Second)
		}
	}

	err := task.WaitTaskCompletion()
	if err != nil {
		return diag.Errorf("error waiting for task to complete: %+v", err)
	}
	return nil
}
