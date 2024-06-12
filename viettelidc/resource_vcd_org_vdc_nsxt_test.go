//go:build nsxt || vdc || ALL || functional

package viettelidc

import "testing"

func TestAccVcdOrgVdcNsxt(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	allocationModel := "ReservationPool"

	var params = StringMap{
		"VdcName":                    t.Name(),
		"OrgName":                    testConfig.VCD.Org,
		"AllocationModel":            "ReservationPool",
		"ProviderVdc":                testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":                testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                  "1024",
		"Reserved":                   "1024",
		"Limit":                      "1024",
		"LimitIncreased":             "1100",
		"AllocatedIncreased":         "1100",
		"ProviderVdcStorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ProviderVdcStorageProfile2": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Tags":                       "vdc nsxt",
		"FuncName":                   t.Name(),
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         " ",
		"FlexElasticKey":                     " ",
		"FlexElasticValue":                   " ",
		"FlexElasticValueUpdate":             " ",
		"FlexMemoryOverheadKey":              " ",
		"FlexMemoryOverheadValue":            " ",
		"FlexMemoryOverheadValueUpdate":      " ",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
	}

	runOrgVdcTest(t, params, allocationModel)
	postTestChecks(t)
}
