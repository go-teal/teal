package modeltests

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_EMPLOYEES_UNIQUE = `


-- Test for duplicate employee keys
select 
    employee_key, 
    count(*) as duplicate_count 
from dds.dim_employees 
group by employee_key 
having count(*) > 1
`

const COUNT_TEST_SQL_DDS_TEST_DIM_EMPLOYEES_UNIQUE = `
select count(*) as test_count from 
(


-- Test for duplicate employee keys
select 
    employee_key, 
    count(*) as duplicate_count 
from dds.dim_employees 
group by employee_key 
having count(*) > 1
) having test_count > 0 limit 1
`


var ddsTestDimEmployeesUniqueTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_employees_unique",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_EMPLOYEES_UNIQUE,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_EMPLOYEES_UNIQUE,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_employees_unique",
		Stage: 				"dds",
		Description: 		`IyMg8J+UjSBFbXBsb3llZSBEaW1lbnNpb24gVW5pcXVlbmVzcyBUZXN0CgoqKlRlc3QgVHlwZSoqOiBEYXRhIFF1YWxpdHkgLSBQcmltYXJ5IEtleSBDb25zdHJhaW50CgoqKkJ1c2luZXNzIFJ1bGUqKjogRWFjaCBlbXBsb3llZSBtdXN0IGhhdmUgZXhhY3RseSBvbmUgZGltZW5zaW9uIHJlY29yZAoKKipTUUwgTG9naWMqKjoKYGBgc3FsCkdST1VQIEJZIGVtcGxveWVlX2tleSBIQVZJTkcgQ09VTlQoKikgPiAxCmBgYAoKKipGYWlsdXJlIENvbnNlcXVlbmNlcyoqOgotIPCfkaUgRG91YmxlLWNvdW50aW5nIGluIHV0aWxpemF0aW9uIG1ldHJpY3MKLSDwn5KwIEluY29ycmVjdCBzYWxhcnkgYWdncmVnYXRpb25zCi0g8J+TiiBXcm9uZyBoZWFkY291bnQgcmVwb3J0cwotIOKaoO+4jyBDb21wbGlhbmNlIHJlcG9ydGluZyBlcnJvcnMKCioqSW52ZXN0aWdhdGlvbiBTdGVwcyoqOgoxLiBDaGVjayBzb3VyY2UgZGF0YSBmb3IgZHVwbGljYXRlIGVtcGxveWVlX2lkcwoyLiBWZXJpZnkgU0hBMjU2IGhhc2ggZ2VuZXJhdGlvbiBsb2dpYwozLiBSZXZpZXcgaW5jcmVtZW50YWwgbG9hZCBwcm9jZXNzCgoqKlN1Y2Nlc3MgQ3JpdGVyaWEqKjogRW1wdHkgcmVzdWx0IHNldA==`,
		Connection: 		"default",
	},
}

var ddsTestDimEmployeesUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimEmployeesUniqueTestDescriptor)