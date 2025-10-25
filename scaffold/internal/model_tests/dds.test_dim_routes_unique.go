package modeltests

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_ROUTES_UNIQUE = `
-- Test for duplicate route keys
select 
    route_key, 
    count(*) as duplicate_count 
from dds.dim_routes 
group by route_key 
having count(*) > 1
`

const COUNT_TEST_SQL_DDS_TEST_DIM_ROUTES_UNIQUE = `
select count(*) as test_count from
(
-- Test for duplicate route keys
select 
    route_key, 
    count(*) as duplicate_count 
from dds.dim_routes 
group by route_key 
having count(*) > 1
) having test_count > 0 limit 1
`


var ddsTestDimRoutesUniqueTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_routes_unique",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_ROUTES_UNIQUE,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_ROUTES_UNIQUE,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_routes_unique",
		Stage: 				"dds",
		Connection: 		"default",
	},
}

var ddsTestDimRoutesUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimRoutesUniqueTestDescriptor)
