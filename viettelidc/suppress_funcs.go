package viettelidc

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// suppressWordToEmptyString is a DiffSuppressFunc which ignore the change from word to empty string "".
// This is useful when API returns some default value but it is not set (and not sent via API) in config.
func suppressWordToEmptyString(word string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		if old == word && new == "" {
			return true
		}
		return false
	}

}

// suppressNetworkUpgradedInterface is used to silence the changes in
// property "interface_type" in routed networks.
// In the old the version, the "internal" interface was implicit,
// while in the new one it is one of several.
// This function only considers the "internal" value, as the other interface types
// were not possible in the previous version
func suppressNetworkUpgradedInterface() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		if old == "" && new == "internal" {
			return true
		}
		return false
	}
}

// falseBoolSuppress suppresses change if value is set to false or is empty
func falseBoolSuppress() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		_, isTrue := d.GetOkExists(k)
		return !isTrue
	}
}

// suppressNewFalse always suppresses when new value is false
func suppressFalse() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return new == "false"
	}
}

// suppressCase is a schema.SchemaDiffSuppressFunc which ignore case changes
func suppressCase(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
