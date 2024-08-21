//go:build functional || auth || ALL

package viettelidc

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

const testTokenFile = "test-token.json"

// TestAccAuth aims to test out all possible ways of `provider` section configuration and allow
// authentication. It tests:
// * local system username and password auth with default org and vdc
// * local system username and password auth with default org
// * local system username and password auth
// * Saml username and password auth (if testConfig.Provider.UseSamlAdfs is true)
// * token based authentication
// * token based authentication priority over username and password
// * token based authentication with auth_type=token
// * auth_type=saml_adfs,EmptySysOrg (if testConfig.Provider.SamlUser and
// testConfig.Provider.SamlPassword are provided)
// Note. Because this test does not use regular templateFill function - it will not generate binary
// tests, but there should be no need for them as well.
func TestAccAuth(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// Reset connection cache just to be sure that we are not reusing any
	cachedVCDClients.reset()

	// All other acceptance tests work by relying on environment variables being set in function
	// `getConfigStruct` to configure authentication method. However, because this test wants to test
	// combinations of accepted `provider` block auth configurations we are setting it as string
	// directly in `provider` and any environment variables set need to be unset during the run of
	// this test and restored afterwards.
	envVars := newEnvVarHelper()
	envVars.saveVcdVars()
	t.Logf("Clearing all VCD env variables")
	envVars.unsetVcdVars()
	defer func() {
		t.Logf("Restoring all VCD env variables")
		envVars.restoreVcdVars()
	}()

	type authTestCase struct {
		name        string
		configText  string
		skip        bool // To make subtests always show names
		skipReason  string
		expectError *regexp.Regexp
	}
	type authTests []authTestCase

	testCases := authTests{}

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPasswordWithDefaultOrgAndVdc",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "` + testConfig.Provider.Password + `"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				vdc                  = "` + testConfig.VCD.Vdc + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:        "InvalidSystemUserAndPasswordWithDefaultOrgAndVdc",
		skip:        testConfig.Provider.UseSamlAdfs,
		skipReason:  "testConfig.Provider.UseSamlAdfs must be false",
		expectError: regexp.MustCompile("401"),
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "INVALID-PASSWORD"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				vdc                  = "` + testConfig.VCD.Vdc + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPasswordWithDefaultOrg",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "` + testConfig.Provider.Password + `"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPassword,AuthType=integrated",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "` + testConfig.Provider.Password + `"
				auth_type            = "integrated"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})
	testCases = append(testCases, authTestCase{
		name:        "InvalidPassword,AuthType=integrated",
		skip:        testConfig.Provider.UseSamlAdfs,
		skipReason:  "testConfig.Provider.UseSamlAdfs must be false",
		expectError: regexp.MustCompile("401"),
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "INVALID-PASSWORD"
				auth_type            = "integrated"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:        "InvalidSystemUserAndPassword,AuthType=integrated",
		skip:        testConfig.Provider.UseSamlAdfs,
		skipReason:  "testConfig.Provider.UseSamlAdfs must be false",
		expectError: regexp.MustCompile("401"),
		configText: `
			provider "vcd" {
				user                 = "INVALID-USER"
				password             = "INVALID-PASSWORD"
				auth_type            = "integrated"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SamlSystemUserAndPassword,AuthType=saml_adfs",
		skip:       !testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be true",
		configText: `
			provider "vcd" {
				user                 = "` + testConfig.Provider.User + `"
				password             = "` + testConfig.Provider.Password + `"
				auth_type            = "saml_adfs"
				saml_adfs_rpt_id     = "` + testConfig.Provider.CustomAdfsRptId + `"
				sysorg               = "` + testConfig.Provider.SysOrg + `" 
				org                  = "` + testConfig.VCD.Org + `"
				vdc                  = "` + testConfig.VCD.Vdc + `"
				url                  = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "SystemUserAndPasswordWithoutSysOrg",
		configText: `
		provider "vcd" {
		  user                 = "` + testConfig.Provider.User + `"
		  password             = "` + testConfig.Provider.Password + `"
		  org                  = "` + testConfig.Provider.SysOrg + `" 
		  url                  = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	// To test token auth we must gain it at first
	tempConn := createTemporaryVCDConnection(false)

	testCases = append(testCases, authTestCase{
		name: "TokenAuth",
		configText: `
		provider "vcd" {
			token                = "` + tempConn.Client.VCDToken + `"
			auth_type            = "token"
			sysorg               = "` + testConfig.Provider.SysOrg + `" 
			org                  = "` + testConfig.VCD.Org + `"
			vdc                  = "` + testConfig.VCD.Vdc + `"
			url                  = "` + testConfig.Provider.Url + `"
			allow_unverified_ssl = true
		  }
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenAuthOnly,AuthType=token",
		configText: `
		provider "vcd" {
			token                = "` + tempConn.Client.VCDToken + `"
			auth_type            = "token"
			sysorg               = "` + testConfig.Provider.SysOrg + `" 
			org                  = "` + testConfig.VCD.Org + `"
			vdc                  = "` + testConfig.VCD.Vdc + `"
			url                  = "` + testConfig.Provider.Url + `"
			allow_unverified_ssl = true
		  }
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenPriorityOverUserAndPassword",
		configText: `
		provider "vcd" {
		  user                 = "invalidUser"
		  password             = "invalidPassword"
		  token                = "` + tempConn.Client.VCDToken + `"
		  auth_type            = "token"
		  sysorg               = "` + testConfig.Provider.SysOrg + `" 
		  org                  = "` + testConfig.VCD.Org + `"
		  vdc                  = "` + testConfig.VCD.Vdc + `"
		  url                  = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenWithUserAndPassword,AuthType=token",
		configText: `
		provider "vcd" {
		  auth_type            = "token"
		  user                 = "invalidUser"
		  password             = "invalidPassword"
		  token                = "` + tempConn.Client.VCDToken + `"
		  sysorg               = "` + testConfig.Provider.SysOrg + `" 
		  org                  = "` + testConfig.VCD.Org + `"
		  vdc                  = "` + testConfig.VCD.Vdc + `"
		  url                  = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	// auth_type=saml_adfs is only run if credentials were provided
	testCases = append(testCases, authTestCase{
		name:       "EmptySysOrg,AuthType=saml_adfs",
		skip:       testConfig.Provider.SamlUser == "" || testConfig.Provider.SamlPassword == "",
		skipReason: "testConfig.Provider.SamlUser and testConfig.Provider.SamlPassword must be set",
		configText: `
			provider "vcd" {
			  auth_type            = "saml_adfs"
			  user                 = "` + testConfig.Provider.SamlUser + `"
			  password             = "` + testConfig.Provider.SamlPassword + `"
			  saml_adfs_rpt_id     = "` + testConfig.Provider.SamlCustomRptId + `"
			  org                  = "` + testConfig.VCD.Org + `"
			  vdc                  = "` + testConfig.VCD.Vdc + `"
			  url                  = "` + testConfig.Provider.Url + `"
			  allow_unverified_ssl = true
			}
		  `,
	})

	// Conditional test on API tokens. This subtest will run only if an API token is defined
	// in an environment variable
	// Note: since this test has a manual input, there is no skip for VCD version. This test will fail if
	// run on VCD < 10.3.1
	apiToken := os.Getenv("TEST_VCD_API_TOKEN")
	if apiToken != "" {
		testOrg := os.Getenv("TEST_VCD_ORG")
		// If sysOrg is not defined in an environment variable, the API token must be one created for the
		// organization stated in testConfig.VCD.Org
		if testOrg == "" {
			testOrg = testConfig.VCD.Org
		}
		testCases = append(testCases, authTestCase{
			name: "ApiToken,AuthType=api_token",
			configText: `
			provider "vcd" {
			  user                 = "invalidUser"
		      password             = "invalidPassword"
		      api_token            = "` + apiToken + `"
		      auth_type            = "api_token"
		      org                  = "` + testOrg + `"
		      url                  = "` + testConfig.Provider.Url + `"
		      allow_unverified_ssl = true
			}
		  `,
		})
	}

	// Conditional test on API tokens. This subtest will run only if an API token is defined
	// in an environment variable
	// Note: since this test has a manual input, there is no skip for VCD version. This test will fail if
	// run on VCD < 10.4.0
	apiTokenFile := os.Getenv("TEST_VCD_SA_TOKEN_FILE")
	if apiTokenFile != "" {
		testOrg := os.Getenv("TEST_VCD_ORG")
		// If sysOrg is not defined in an environment variable, the API token must be one created for the
		// organization stated in testConfig.VCD.Org
		if testOrg == "" {
			testOrg = testConfig.VCD.Org
		}
		testCases = append(testCases, authTestCase{
			name: "ServiceAccountTokenFile,AuthType=service_account_token_file",
			configText: `
			provider "vcd" {
			  user                       = "invalidUser"
		      password                   = "invalidPassword"
			  api_token_file             = "` + apiTokenFile + `"
		      auth_type                  = "api_token_file"
		      org                        = "` + testOrg + `"
		      url                        = "` + testConfig.Provider.Url + `"
		      allow_unverified_ssl       = true
			}
		  `,
		})

		// Testing sending an invalid Service Account token
		createTestTokenFile(t)
		testCases = append(testCases, authTestCase{
			name:        "ApiTokenFile,AuthType=invalid_api_token_file",
			expectError: regexp.MustCompile(".*(Invalid refresh token|server_error)"),
			configText: `
				provider "vcd" {
					auth_type                  = "api_token_file" 
					api_token_file             = "` + testTokenFile + `"
					sysorg                     = "` + testConfig.Provider.SysOrg + `" 
					org                        = "` + testConfig.VCD.Org + `"
					vdc                        = "` + testConfig.VCD.Vdc + `"
					url                        = "` + testConfig.Provider.Url + `"
					allow_unverified_ssl       = true
				}
		`,
		})
	}

	saTokenFile := os.Getenv("TEST_VCD_SA_TOKEN_FILE")
	if saTokenFile != "" {
		testOrg := os.Getenv("TEST_VCD_ORG")
		// If sysOrg is not defined in an environment variable, the API token must be one created for the
		// organization stated in testConfig.VCD.Org
		if testOrg == "" {
			testOrg = testConfig.VCD.Org
		}
		testCases = append(testCases, authTestCase{
			name: "ServiceAccountTokenFile,AuthType=service_account_token_file",
			configText: `
			provider "vcd" {
			  user                       = "invalidUser"
		      password                   = "invalidPassword"
			  service_account_token_file = "` + saTokenFile + `"
		      auth_type                  = "service_account_token_file"
		      org                        = "` + testOrg + `"
		      url                        = "` + testConfig.Provider.Url + `"
		      allow_unverified_ssl       = true
			}
		  `,
		})

		// Testing sending an invalid Service Account token
		createTestTokenFile(t)
		testCases = append(testCases, authTestCase{
			name:        "ServiceAccountTokenFile,AuthType=invalid_service_account_token_file",
			expectError: regexp.MustCompile(".*(Invalid refresh token|server_error)"),
			configText: `
				provider "vcd" {
					auth_type                  = "service_account_token_file" 
					service_account_token_file = "` + testTokenFile + `"
					sysorg                     = "` + testConfig.Provider.SysOrg + `" 
					org                        = "` + testConfig.VCD.Org + `"
					vdc                        = "` + testConfig.VCD.Vdc + `"
					url                        = "` + testConfig.Provider.Url + `"
					allow_unverified_ssl       = true
				}
		`,
		})
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip("Skipping: " + test.skipReason)
			}
			runAuthTest(t, test.configText, test.expectError)
		})
	}

	// Clear connection cache to force other tests use their own mechanism
	cleanTestTokenFile(t)
	cachedVCDClients.reset()
	postTestChecks(t)
}

func runAuthTest(t *testing.T, configText string, expectError *regexp.Regexp) {

	dataSource := `
	data "vcd_org" "auth" {
		name = "` + testConfig.VCD.Org + `"
	}
	`

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectError: expectError,
				Config:      configText + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_org.auth", "id"),
				),
			},
		},
	})
}

// createTestTokenFile creates a test token file with a dummy token for
// API token authentication
func createTestTokenFile(t *testing.T) {
	// Check if token file exists
	if !fileExists(testTokenFile) {
		testData := types.ApiTokenRefresh{
			RefreshToken: "testToken",
		}
		data, err := json.Marshal(testData)
		if err != nil {
			t.Skipf("error creating test token file: %s", err)
		}

		err = os.WriteFile(testTokenFile, data, 0600)
		if err != nil {
			t.Skipf("error creating test token file: %s", err)
		}
	}
}

// cleanTestTokenFile is ueed to remove the test token file after the test
func cleanTestTokenFile(t *testing.T) {
	if fileExists(testTokenFile) {
		err := os.Remove(testTokenFile)
		if err != nil {
			t.Skipf("error removing test token file: %s", err)
		}
	}
}
