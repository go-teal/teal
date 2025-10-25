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
		Connection: 		"default",
	},
}

var ddsTestDimEmployeesUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimEmployeesUniqueTestDescriptor)
