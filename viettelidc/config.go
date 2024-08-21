package viettelidc

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func init() {
	separator := os.Getenv("VCD_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	}
}

type Config struct {
	User                    string
	Password                string
	Token                   string // Token used instead of user and password
	ApiToken                string // User generated token used instead of user and password
	ApiTokenFile            string // File containing a user generated API token
	AllowApiTokenFile       bool   // Setting to suppress API Token File security warnings
	ServiceAccountTokenFile string // File containing the Service Account API token
	AllowSATokenFile        bool   // Setting to suppress Service Account Token File security warnings
	SysOrg                  string // Org used for authentication
	Org                     string // Default Org used for API operations
	Vdc                     string // Default (optional) VDC for API operations
	Href                    string
	MaxRetryTimeout         int
	InsecureFlag            bool

	// UseSamlAdfs specifies if SAML auth is used for authenticating VCD instead of local login.
	// The following conditions must be met so that authentication SAML authentication works:
	// * SAML IdP (Identity Provider) is Active Directory Federation Service (ADFS)
	// * Authentication endpoint "/adfs/services/trust/13/usernamemixed" must be enabled on ADFS
	// server
	UseSamlAdfs bool
	// CustomAdfsRptId allows to set custom Relaying Party Trust identifier. By default VCD Entity
	// ID is used as Relaying Party Trust identifier.
	CustomAdfsRptId string

	// IgnoredMetadata allows to configure a set of metadata entries that should be ignored by all the
	// API operations related to metadata.
	IgnoredMetadata []govcd.IgnoredMetadata
}

type VCDClient struct {
	*govcd.VCDClient
	SysOrg          string
	Org             string // name of default Org
	Vdc             string // name of default VDC
	MaxRetryTimeout int
	InsecureFlag    bool
}

// StringMap type is used to simplify reading resource definitions
type StringMap map[string]interface{}

const (
	// Most common error messages in the library

	// Used when a call to GetOrgAndVdc fails. The placeholder is for the error
	errorRetrievingOrgAndVdc = "error retrieving Org and VDC: %s"

	// Used when a call to GetOrgAndVdc fails. The placeholders are for vdc, org, and the error
	errorRetrievingVdcFromOrg = "error retrieving VDC %s from Org %s: %s"

	// Used when we can't get a valid edge gateway. The placeholder is for the error
	errorUnableToFindEdgeGateway = "unable to find edge gateway: %s"

	// Used when a task fails. The placeholder is for the error
	errorCompletingTask = "error completing tasks: %s"

	// Used when a call to GetAdminOrgFromResource fails. The placeholder is for the error
	errorRetrievingOrg = "error retrieving Org: %s"
)

// Cache values for vCD connection.
// When the Client() function is called with the same parameters, it will return
// a cached value instead of connecting again.
// This makes the Client() function both deterministic and fast.
//
// WARNING: Cached clients need to be evicted by calling cacheStorage.reset() after the rights of the associated
// logged user change. Otherwise, retrieving or manipulating objects that require the new rights could return 403
// forbidden errors. For example, adding read rights of a RDE Type to a specific user requires a cacheStorage.reset()
// afterwards, to force a re-authentication. If this is not done, the cached client won't be able to read this RDE Type.
type cachedConnection struct {
	initTime   time.Time
	connection *VCDClient
}

type cacheStorage struct {
	// conMap holds cached VDC authenticated connection
	conMap map[string]cachedConnection
	// cacheClientServedCount records how many times we have cached a connection
	cacheClientServedCount int
	sync.Mutex
}

// reset clears cache to force re-authentication
func (c *cacheStorage) reset() {
	c.Lock()
	defer c.Unlock()
	c.cacheClientServedCount = 0
	c.conMap = make(map[string]cachedConnection)
}

var (
	// Enables the caching of authenticated connections
	enableConnectionCache = os.Getenv("VCD_CACHE") != ""

	// Cached VDC authenticated connection
	cachedVCDClients = &cacheStorage{conMap: make(map[string]cachedConnection)}

	// Invalidates the cache after a given time (connection tokens usually expire after 20 to 30 minutes)
	maxConnectionValidity = 20 * time.Minute

	enableDebug = os.Getenv("GOVCD_DEBUG") != ""
	enableTrace = os.Getenv("GOVCD_TRACE") != ""

	// ImportSeparator is the separation string used for import operations
	// Can be changed using either "import_separator" property in Provider
	// or environment variable "VCD_IMPORT_SEPARATOR"
	ImportSeparator = "."
)

// Displays conditional messages
func debugPrintf(format string, args ...interface{}) {
	// When GOVCD_TRACE is enabled, we also display the function that generated the message
	if enableTrace {
		format = fmt.Sprintf("[%s] %s", filepath.Base(callFuncName()), format)
	}
	// The formatted message passed to this function is displayed only when GOVCD_DEBUG is enabled.
	if enableDebug {
		fmt.Printf(format, args...)
	}
}

// This is a global mutexKV for all resources
var vcdMutexKV = newMutexKV()

func (cli *VCDClient) lockVapp(d *schema.ResourceData) {
	vappName := d.Get("name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvLock(key)
}

func (cli *VCDClient) unLockVapp(d *schema.ResourceData) {
	vappName := d.Get("name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvUnlock(key)
}

// lockEdgeGateway locks an edge gateway resource
// id field is used as key
func (cli *VCDClient) lockEdgeGateway(d *schema.ResourceData) {
	edgeGatewayId := d.Id()
	if edgeGatewayId == "" {
		panic("edge gateway ID not found")
	}

	vcdMutexKV.kvLock(edgeGatewayId)
}

// unlockEdgeGateway unlocks an Edge Gateway resource
// id field is used as key
func (cli *VCDClient) unlockEdgeGateway(d *schema.ResourceData) {
	edgeGatewayId := d.Id()
	if edgeGatewayId == "" {
		panic("edge gateway ID not found")
	}

	vcdMutexKV.kvUnlock(edgeGatewayId)
}

// lockParentVappWithName locks using provided vappName.
// Parent means the resource belongs to the vApp being locked
func (cli *VCDClient) lockParentVappWithName(d *schema.ResourceData, vappName string) {
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvLock(key)
}

func (cli *VCDClient) unLockParentVappWithName(d *schema.ResourceData, vappName string) {
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvUnlock(key)
}

// function lockParentVapp locks using vapp_name name existing in resource parameters.
// Parent means the resource belongs to the vApp being locked
func (cli *VCDClient) lockParentVapp(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvLock(key)
}

func (cli *VCDClient) unLockParentVapp(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", cli.getOrgName(d), cli.getVdcName(d), vappName)
	vcdMutexKV.kvUnlock(key)
}

func (cli *VCDClient) lockVappWithName(org, vdc, vappName string) func() {
	if vappName == "" {
		panic("vApp name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s", org, vdc, vappName)
	vcdMutexKV.kvLock(key)

	return func() {
		vcdMutexKV.kvUnlock(key)
	}
}

// lockParentVm locks using vapp_name and vm_name names existing in resource parameters.
// Parent means the resource belongs to the VM being locked
//
//lint:ignore U1000 For future use
func (cli *VCDClient) lockParentVm(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	vmName := d.Get("vm_name").(string)
	if vmName == "" {
		panic("vmName name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s|vm:%s", cli.getOrgName(d), cli.getVdcName(d), vappName, vmName)
	vcdMutexKV.kvLock(key)
}

//lint:ignore U1000 For future use
func (cli *VCDClient) unLockParentVm(d *schema.ResourceData) {
	vappName := d.Get("vapp_name").(string)
	if vappName == "" {
		panic("vApp name not found")
	}
	vmName := d.Get("vm_name").(string)
	if vmName == "" {
		panic("vmName name not found")
	}
	key := fmt.Sprintf("org:%s|vdc:%s|vapp:%s|vm:%s", cli.getOrgName(d), cli.getVdcName(d), vappName, vmName)
	vcdMutexKV.kvUnlock(key)
}

// lockById locks on supplied ID field
func (cli *VCDClient) lockById(id string) {
	vcdMutexKV.kvLock(id)
}

// unlockById unlocks on supplied ID field
func (cli *VCDClient) unlockById(id string) {
	vcdMutexKV.kvUnlock(id)
}

// lockParentVdcGroup locks on VDC Group ID using 'vdc_group_id' field
func (cli *VCDClient) lockParentVdcGroup(d *schema.ResourceData) {
	vdcGroupId := d.Get("vdc_group_id").(string)
	if vdcGroupId == "" {
		panic("'vdc_group_id' is empty")
	}

	vcdMutexKV.kvLock(vdcGroupId)
}

// unlockParentVdcGroup unlocks on VDC Group ID using 'vdc_group_id' field
func (cli *VCDClient) unlockParentVdcGroup(d *schema.ResourceData) {
	vdcGroupId := d.Get("vdc_group_id").(string)
	if vdcGroupId == "" {
		panic("'vdc_group_id' is empty")
	}

	vcdMutexKV.kvUnlock(vdcGroupId)
}

// lockParentExternalNetwork locks on External Network using 'external_network_id' field
func (cli *VCDClient) lockParentExternalNetwork(d *schema.ResourceData) {
	externalNetworkId := d.Get("external_network_id").(string)
	if externalNetworkId == "" {
		panic("'external_network_id' is empty")
	}

	vcdMutexKV.kvLock(externalNetworkId)
}

// unlockParentVdcGroup unlocks on External Network using 'external_network_id' field
func (cli *VCDClient) unlockParentExternalNetwork(d *schema.ResourceData) {
	externalNetworkId := d.Get("external_network_id").(string)
	if externalNetworkId == "" {
		panic("'external_network_id' is empty")
	}

	vcdMutexKV.kvUnlock(externalNetworkId)
}

// lockIfOwnerIsVdcGroup locks VDC Group based on `owner_id` field (if it is a VDC Group)
func (cli *VCDClient) lockIfOwnerIsVdcGroup(d *schema.ResourceData) {
	vdcGroupId := d.Get("owner_id")
	vdcGroupIdValue := vdcGroupId.(string)
	if govcd.OwnerIsVdcGroup(vdcGroupIdValue) {
		vcdMutexKV.kvLock(vdcGroupIdValue)
	}
}

// unLockIfOwnerIsVdcGroup unlocks VDC Group based on `owner_id` field (if it is a VDC Group)
func (cli *VCDClient) unLockIfOwnerIsVdcGroup(d *schema.ResourceData) {
	vdcGroupId := d.Get("owner_id")
	vdcGroupIdValue := vdcGroupId.(string)
	if govcd.OwnerIsVdcGroup(vdcGroupIdValue) {
		vcdMutexKV.kvUnlock(vdcGroupIdValue)
	}
}

// function lockParentEdgeGtw locks using edge_gateway or edge_gateway_id name existing in resource parameters.
// Edge Gateway is used as a lock key. If only `name` is present in resource - it will find the Edge Gateway itself
func (cli *VCDClient) lockParentEdgeGtw(d *schema.ResourceData) {
	var edgeGtwIdValue string
	var edgeGtwNameValue string

	edgeGtwId := d.Get("edge_gateway_id")
	if edgeGtwId != nil {
		edgeGtwIdValue = edgeGtwId.(string)
	}

	edgeGtwName := d.Get("edge_gateway")
	if edgeGtwName != nil {
		edgeGtwNameValue = edgeGtwName.(string)
	}

	// If none are specified - panic
	if edgeGtwIdValue == "" && edgeGtwNameValue == "" {
		panic("edge gateway not found")
	}

	// Only Edge gateway name ('edge_gateway' field) was specified - need to lookup ID
	if edgeGtwNameValue != "" && edgeGtwIdValue == "" {
		egw, err := cli.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			panic(fmt.Sprintf("edge gateway '%s' not found: %s", edgeGtwNameValue, err))
		}

		edgeGtwIdValue = egw.EdgeGateway.ID
	}

	if edgeGtwIdValue == "" {
		panic("edge gateway ID not found")
	}

	vcdMutexKV.kvLock(edgeGtwIdValue)
}

func (cli *VCDClient) unLockParentEdgeGtw(d *schema.ResourceData) {
	var edgeGtwIdValue string
	var edgeGtwNameValue string

	edgeGtwId := d.Get("edge_gateway_id")
	if edgeGtwId != nil {
		edgeGtwIdValue = edgeGtwId.(string)
	}

	edgeGtwName := d.Get("edge_gateway")
	if edgeGtwName != nil {
		edgeGtwNameValue = edgeGtwName.(string)
	}

	// If none are specified - panic
	if edgeGtwIdValue == "" && edgeGtwNameValue == "" {
		panic("edge gateway not found")
	}

	// Only Edge gateway name ('edge_gateway' field) was specified - need to lookup ID
	if edgeGtwNameValue != "" && edgeGtwIdValue == "" {
		egw, err := cli.GetEdgeGatewayFromResource(d, "edge_gateway")
		if err != nil {
			panic(fmt.Sprintf("edge gateway '%s' not found: %s", edgeGtwNameValue, err))
		}

		edgeGtwIdValue = egw.EdgeGateway.ID
	}

	if edgeGtwIdValue == "" {
		panic("edge gateway ID not found")
	}

	vcdMutexKV.kvUnlock(edgeGtwIdValue)
}

// lockParentVdcGroupOrEdgeGateway handles lock of parent Edge Gateway or parent VDC group, depending
// if the parent Edge Gateway is in a VDC or a VDC group. Returns a function that contains the needed
// unlock function, so that it can be deferred and called after the work with the resource has been
// done.
func (cli *VCDClient) lockParentVdcGroupOrEdgeGateway(d *schema.ResourceData) (func(), error) {
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(cli, d)
	if err != nil {
		return nil, fmt.Errorf("error finding parent Edge Gateway: %s", err)
	}

	// Handling locks is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		cli.lockById(parentEdgeGatewayOwnerId)
		return func() {
			cli.unlockById(parentEdgeGatewayOwnerId)
		}, nil
	} else {
		cli.lockParentEdgeGtw(d)
		return func() {
			cli.unLockParentEdgeGtw(d)
		}, nil
	}
}

func (cli *VCDClient) lockParentOrgNetwork(d *schema.ResourceData) {
	orgNetworkId := d.Get("org_network_id").(string)
	vcdMutexKV.kvLock(orgNetworkId)
}

func (cli *VCDClient) unLockParentOrgNetwork(d *schema.ResourceData) {
	orgNetworkId := d.Get("org_network_id").(string)
	vcdMutexKV.kvUnlock(orgNetworkId)
}

func (cli *VCDClient) getOrgName(d *schema.ResourceData) string {
	orgName := d.Get("org").(string)
	if orgName == "" {
		orgName = cli.Org
	}
	return orgName
}

func (cli *VCDClient) getVdcName(d *schema.ResourceData) string {
	orgName := d.Get("vdc").(string)
	if orgName == "" {
		orgName = cli.Vdc
	}
	return orgName
}

// GetOrgAndVdc finds a pair of org and vdc using the names provided
// in the args. If the names are empty, it will use the default
// org and vdc names from the provider.
func (cli *VCDClient) GetOrgAndVdc(orgName, vdcName string) (org *govcd.Org, vdc *govcd.Vdc, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return nil, nil, fmt.Errorf("empty Org name provided")
	}
	if vdcName == "" {
		vdcName = cli.Vdc
	}
	if vdcName == "" {
		return nil, nil, fmt.Errorf("empty VDC name provided")
	}
	org, err = cli.VCDClient.GetOrgByName(orgName)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	if org.Org.Name == "" || org.Org.HREF == "" || org.Org.ID == "" {
		return nil, nil, fmt.Errorf("empty Org %s found ", orgName)
	}
	vdc, err = org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving VDC %s: %s", vdcName, err)
	}
	if vdc == nil || vdc.Vdc.ID == "" || vdc.Vdc.HREF == "" || vdc.Vdc.Name == "" {
		return nil, nil, fmt.Errorf("error retrieving VDC %s: not found", vdcName)
	}
	return org, vdc, err
}

// GetAdminOrg finds org using the names provided in the args.
// If the name is empty, it will use the default
// org name from the provider.
func (cli *VCDClient) GetAdminOrg(orgName string) (org *govcd.AdminOrg, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return nil, fmt.Errorf("empty Org name provided")
	}

	org, err = cli.VCDClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	if org.AdminOrg.Name == "" || org.AdminOrg.HREF == "" || org.AdminOrg.ID == "" {
		return nil, fmt.Errorf("empty org %s found", orgName)
	}
	return org, err
}

// GetOrg finds org using the names provided in the args.
// If the name is empty, it will use the default
// org name from the provider.
func (cli *VCDClient) GetOrg(orgName string) (org *govcd.Org, err error) {

	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return nil, fmt.Errorf("empty Org name provided")
	}

	org, err = cli.VCDClient.GetOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org %s: %s", orgName, err)
	}
	if org.Org.Name == "" || org.Org.HREF == "" || org.Org.ID == "" {
		return nil, fmt.Errorf("empty Org %s found", orgName)
	}
	return org, err
}

// GetOrgFromResource is the same as GetOrg, but using data from the resource, if available.
func (cli *VCDClient) GetOrgFromResource(d *schema.ResourceData) (org *govcd.Org, err error) {
	orgName := d.Get("org").(string)
	return cli.GetOrg(orgName)
}

// GetOrgAndVdcFromResource is the same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetOrgAndVdcFromResource(d *schema.ResourceData) (org *govcd.Org, vdc *govcd.Vdc, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	return cli.GetOrgAndVdc(orgName, vdcName)
}

// GetAdminOrgFromResource is the same as GetOrgAndVdc, but using data from the resource, if available.
func (cli *VCDClient) GetAdminOrgFromResource(d *schema.ResourceData) (org *govcd.AdminOrg, err error) {
	orgName := d.Get("org").(string)
	return cli.GetAdminOrg(orgName)
}

// GetEdgeGateway gets an NSX-V Edge Gateway when you don't need org or vdc for other purposes
func (cli *VCDClient) GetEdgeGateway(orgName, vdcName, edgeGwName string) (eg *govcd.EdgeGateway, err error) {

	if edgeGwName == "" {
		return nil, fmt.Errorf("empty Edge Gateway name provided")
	}
	_, vdc, err := cli.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org and VDC: %s", err)
	}
	eg, err = vdc.GetEdgeGatewayByName(edgeGwName, true)

	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf(fmt.Sprintf("(%s) [%s] ", edgeGwName, callFuncName())+errorUnableToFindEdgeGateway, err)
		}
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
	return eg, nil
}

// GetNsxtEdgeGateway gets an NSX-T Edge Gateway when you don't need Org or VDC for other purposes
func (cli *VCDClient) GetNsxtEdgeGateway(orgName, vdcName, edgeGwName string) (eg *govcd.NsxtEdgeGateway, err error) {

	if edgeGwName == "" {
		return nil, fmt.Errorf("empty NSX-T Edge Gateway name provided")
	}
	_, vdc, err := cli.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org and VDC: %s", err)
	}
	eg, err = vdc.GetNsxtEdgeGatewayByName(edgeGwName)

	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf(fmt.Sprintf("(%s) [%s] ", edgeGwName, callFuncName())+errorUnableToFindEdgeGateway, err)
		}
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
	return eg, nil
}

// GetNsxtEdgeGatewayById gets an NSX-T Edge Gateway when you don't need Org or VDC for other purposes
func (cli *VCDClient) GetNsxtEdgeGatewayById(orgName, edgeGwId string) (eg *govcd.NsxtEdgeGateway, err error) {
	if edgeGwId == "" {
		return nil, fmt.Errorf("empty NSX-T Edge Gateway ID provided")
	}

	org, err := cli.GetOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Org: %s", err)
	}
	eg, err = org.GetNsxtEdgeGatewayById(edgeGwId)

	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf(fmt.Sprintf("(%s) [%s] ", edgeGwId, callFuncName())+errorUnableToFindEdgeGateway, err)
		}
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}
	return eg, nil
}

// GetEdgeGatewayFromResource is the same as GetEdgeGateway, but using data from the resource, if available
// edgeGatewayFieldName is the name used in the resource. It is usually "edge_gateway"
// for all resources that *use* an edge gateway, and when the resource is vcd_edgegateway, it is "name"
func (cli *VCDClient) GetEdgeGatewayFromResource(d *schema.ResourceData, edgeGatewayFieldName string) (eg *govcd.EdgeGateway, err error) {
	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayName := d.Get(edgeGatewayFieldName).(string)
	egw, err := cli.GetEdgeGateway(orgName, vdcName, edgeGatewayName)
	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf("(%s) [%s] : %s", edgeGatewayName, callFuncName(), err)
		}
		return nil, err
	}
	return egw, nil
}

// GetNsxtEdgeGatewayFromResourceById helps to retrieve NSX-T Edge Gateway when Org org VDC are not
// needed. It performs a query By ID.
func (cli *VCDClient) GetNsxtEdgeGatewayFromResourceById(d *schema.ResourceData, edgeGatewayFieldName string) (eg *govcd.NsxtEdgeGateway, err error) {
	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get(edgeGatewayFieldName).(string)
	egw, err := cli.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if os.Getenv("GOVCD_DEBUG") != "" {
			return nil, fmt.Errorf("(%s) [%s] : %s", edgeGatewayId, callFuncName(), err)
		}
		return nil, err
	}
	return egw, nil
}

// GetOrgNameFromResource returns the Org name if set at resource level. If not, tries to get it from provider level.
// It errors if none is provided.
func (cli *VCDClient) GetOrgNameFromResource(d *schema.ResourceData) (string, error) {
	orgName := d.Get("org").(string)
	return cli.GetOrgName(orgName)
}

// GetOrgName returns the parameter orgName if provided. If not tried to get it from provider.
func (cli *VCDClient) GetOrgName(orgName string) (string, error) {
	if orgName == "" {
		orgName = cli.Org
	}
	if orgName == "" {
		return "", fmt.Errorf("empty Org name provided")
	}

	return orgName, nil
}

// TODO Look into refactoring this into a method of *Config
func ProviderAuthenticate(client *govcd.VCDClient, user, password, token, org, apiToken, apiTokenFile, saTokenFile string) error {
	var err error
	if saTokenFile != "" {
		return client.SetServiceAccountApiToken(org, saTokenFile)
	}
	if apiTokenFile != "" {
		_, err := client.SetApiTokenFromFile(org, apiTokenFile)
		if err != nil {
			return err
		}
		return nil
	}
	if apiToken != "" {
		return client.SetToken(org, govcd.ApiTokenHeader, apiToken)
	}
	if token != "" {
		if len(token) > 32 {
			err = client.SetToken(org, govcd.BearerTokenHeader, token)
		} else {
			err = client.SetToken(org, govcd.AuthorizationHeader, token)
		}
		if err != nil {
			return fmt.Errorf("error during token-based authentication: %s", err)
		}
		return nil
	}

	return client.Authenticate(user, password, org)
}

func (c *Config) Client() (*VCDClient, error) {
	rawData := c.User + "#" +
		c.Password + "#" +
		c.Token + "#" +
		c.ApiToken + "#" +
		c.ApiTokenFile + "#" +
		c.ServiceAccountTokenFile + "#" +
		c.SysOrg + "#" +
		c.Vdc + "#" +
		c.Href
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(rawData)))

	// The cached connection is served only if the variable VCD_CACHE is set
	cachedVCDClients.Lock()
	client, ok := cachedVCDClients.conMap[checksum]
	cachedVCDClients.Unlock()
	if ok && enableConnectionCache {
		cachedVCDClients.Lock()
		cachedVCDClients.cacheClientServedCount += 1
		cachedVCDClients.Unlock()
		// debugPrintf("[%s] cached connection served %d times (size:%d)\n",
		elapsed := time.Since(client.initTime)
		if elapsed > maxConnectionValidity {
			debugPrintf("cached connection invalidated after %2.0f minutes \n", maxConnectionValidity.Minutes())
			cachedVCDClients.Lock()
			delete(cachedVCDClients.conMap, checksum)
			cachedVCDClients.Unlock()
		} else {
			return client.connection, nil
		}
	}

	authUrl, err := url.ParseRequestURI(c.Href)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while retrieving URL: %s", err)
	}

	userAgent := buildUserAgent(BuildVersion, c.SysOrg)

	vcdClient := &VCDClient{
		VCDClient: govcd.NewVCDClient(*authUrl, c.InsecureFlag,
			govcd.WithMaxRetryTimeout(c.MaxRetryTimeout),
			govcd.WithSamlAdfs(c.UseSamlAdfs, c.CustomAdfsRptId),
			govcd.WithHttpUserAgent(userAgent),
			govcd.WithIgnoredMetadata(c.IgnoredMetadata),
		),
		SysOrg:          c.SysOrg,
		Org:             c.Org,
		Vdc:             c.Vdc,
		MaxRetryTimeout: c.MaxRetryTimeout,
		InsecureFlag:    c.InsecureFlag}

	err = ProviderAuthenticate(vcdClient.VCDClient, c.User, c.Password, c.Token, c.SysOrg, c.ApiToken, c.ApiTokenFile, c.ServiceAccountTokenFile)
	if err != nil {
		return nil, fmt.Errorf("something went wrong during authentication: %s", err)
	}
	cachedVCDClients.Lock()
	cachedVCDClients.conMap[checksum] = cachedConnection{initTime: time.Now(), connection: vcdClient}
	cachedVCDClients.Unlock()

	return vcdClient, nil
}

// callFuncName returns the name of the function that called the current function. It is used for
// tracing
func callFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}

// buildUserAgent helps to construct HTTP User-Agent header
func buildUserAgent(version, sysOrg string) string {
	userAgent := fmt.Sprintf("terraform-provider-viettelidc/%s (%s/%s; isProvider:%t)",
		version, runtime.GOOS, runtime.GOARCH, strings.ToLower(sysOrg) == "system")

	return userAgent
}

// logForScreen writes to go-vcloud-director log with a tag that can be used to
// filter messages directed at the user.
// * origin is the name of the resource that originates the message
// * msg is the text that will end up in the logs
//
// To display its content at run time, you should run the command below in a separate
// terminal screen while `terraform apply` is running
//
//	tail -f go-vcloud-director.log | grep '\[SCREEN\]'
func logForScreen(origin, msg string) {
	util.Logger.Printf("[SCREEN] {%s} %s\n", origin, msg)
}

// dSet sets the value of a schema property, discarding the error
// Use only for scalar values (strings, booleans, and numbers)
func dSet(d *schema.ResourceData, key string, value interface{}) {
	if value != nil && !isScalar(value) {
		msg1 := "*** ERROR: only scalar values should be used for dSet()"
		msg2 := fmt.Sprintf("*** detected '%s' for key '%s' (called from %s)",
			reflect.TypeOf(value).Kind(), key, callFuncName())
		starLine := strings.Repeat("*", len(msg2))
		// This panic should never reach the final user.
		// Its purpose is to alert the developer that there was an improper use of `dSet`
		panic(fmt.Sprintf("\n%s\n%s\n%s\n%s\n", starLine, msg1, msg2, starLine))
	}
	err := d.Set(key, value)
	if err != nil {
		panic(fmt.Sprintf("error in %s - key '%s': %s ", callFuncName(), key, err))
	}
}

// isScalar returns true if its argument is not a composite object
// we want strings, numbers, booleans
func isScalar(t interface{}) bool {
	if t == nil {
		return true
	}
	typeOf := reflect.TypeOf(t)
	switch typeOf.Kind().String() {
	case "struct", "map", "array", "slice":
		return false
	}

	return true
}
