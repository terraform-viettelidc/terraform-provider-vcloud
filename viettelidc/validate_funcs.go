package viettelidc

//lint:file-ignore SA1019 ignore deprecated functions
import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// vmNetworkTypeValidator checks the valid network types for a VM
// It allows only 'org' and 'none' for standalone VMs, and adds 'vapp'
// for VMs inside a regular vApp
func vmNetworkTypeValidator(vmType typeOfVm) schema.SchemaValidateFunc {
	allowed := []string{"org", "none"}
	if vmType == vappVmType {
		allowed = append(allowed, "vapp")
	}
	return validation.StringInSlice(allowed, false)
}

// noopValueWarningValidator is a no-op validator which only emits warning string when fieldValue
// is set to the specified one
func noopValueWarningValidator(fieldValue interface{}, warningText string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		if fieldValue == i {
			warnings = append(warnings, fmt.Sprintf("%s\n\n", warningText))
		}

		return
	}
}

// anyValueWarningValidator is a validator which only emits always warning string
//
//lint:ignore U1000 For future use
func anyValueWarningValidator(warningText string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		warnings = append(warnings, fmt.Sprintf("%s\n\n", warningText))
		return
	}
}

// checkEmptyOrSingleIP validates if the field is set to empty or a valid IP address
func checkEmptyOrSingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if net.ParseIP(v) == nil && v != "" {
			es = append(es, fmt.Errorf(
				"expected %s to be empty or contain a valid IP, got: %s", k, v))
		}
		return
	}
}

// validateCase checks if a string is of caseType "upper" or "lower"
func validateCase(caseType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		switch caseType {
		case "upper":
			if strings.ToUpper(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be upper cased, got: %s", v))
			}
		case "lower":
			if strings.ToLower(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be lower cased, got: %s", v))
			}
		default:
			panic("unsupported validation type for validateCase() function")
		}
		return
	}
}

// validateBusType checks if bus type is in correct format for independent disk
func validateBusType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)

	if busTypes[strings.ToUpper(value)] == "" {
		errors = append(errors, fmt.Errorf("%q (%q) value isn't valid", k, value))
	}

	return
}

// validateBusSubType checks if bus subtype is in correct format for independent disk
func validateBusSubType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)

	if busSubTypes[strings.ToLower(value)] == "" {
		errors = append(errors, fmt.Errorf("%q (%q) value isn't valid", k, value))
	}

	return
}

// validateMultipleOf4 checks if value is a multiple of 4
func validateMultipleOf4() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		value, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if value%4 != 0 {
			es = append(es, fmt.Errorf("expected %s to be multiple of 4, got %d", k, value))
			return
		}

		return
	}
}

// validateIntLeaseSeconds validates amount of seconds for lease
// A value of 0 is accepted, as it means "never expires"
// Regular values must be > 3600
func validateIntLeaseSeconds() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		valid := i == 0 || v >= 3600
		if !valid {
			es = append(es, fmt.Errorf("expected %s to be either 0 or a number >= 3600 , got %d", k, v))
			return
		}

		return
	}
}

// IsIntAndAtLeast returns a SchemaValidateFunc which tests if the provided value string is convertable to int
// and is at least min (inclusive)
func IsIntAndAtLeast(min int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		value, err := strconv.Atoi(i.(string))
		if err != nil {
			errors = append(errors, fmt.Errorf("expected type of %s to be integer", k))
			return warnings, errors
		}

		if value < min {
			errors = append(errors, fmt.Errorf("expected %s to be at least (%d), got %d", k, min, value))
			return warnings, errors
		}

		return warnings, errors
	}
}

// IsFloatAndBetween returns a SchemaValidateFunc which tests if the provided value convertable to
// float64 and is between min and max (inclusive).
func IsFloatAndBetween(min, max float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		value, err := strconv.ParseFloat(i.(string), 64)
		if err != nil {
			es = append(es, fmt.Errorf("expected type of %s to be float64", k))
			return
		}

		if value < min || value > max {
			es = append(es, fmt.Errorf("expected %s to be in the range (%f - %f), got %f", k, min, max, value))
			return
		}

		return
	}
}

// validateAlphanumericWithUnderscoresAndHyphens returns a SchemaValidateFunc that tests whether the target attribute
// only contains alphanumeric characters, allowing also underscores and hyphens.
func validateAlphanumericWithUnderscoresAndHyphens() schema.SchemaValidateFunc {
	return validation.StringMatch(regexp.MustCompile(`(?i)^[a-z0-9_-]+$`), "only alphanumeric characters, underscores and hyphens allowed")
}

// allowTokenFileIfIsBoolAndTrue checks if the 'allow_token_file' is set to true, otherwise returns
// a descriptive error
func allowTokenFileIfIsBoolAndTrue() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		value := i.(bool)

		if !value {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "This field must be set to true",
					Detail: "The token file should be considered SENSITIVE INFORMATION. " +
						"If you acknowledge that, set 'allow_token_file' to 'true'.",
					AttributePath: path,
				},
			}
		}
		return nil
	}
}

// validateMetadataIgnoreResourceType validates whether the given attribute is a resource type
// that supports the `metadata_entry` argument.
func validateMetadataIgnoreResourceType() func(i interface{}, path cty.Path) diag.Diagnostics {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of '%v' to be string", i)
		}
		found := false
		for k := range resourceMetadataApiRelation {
			if v == k {
				found = true
				break
			}
		}
		if !found {
			return diag.Errorf("can't ignore metadata of resource type '%s'", i)
		}
		return nil
	}
}
