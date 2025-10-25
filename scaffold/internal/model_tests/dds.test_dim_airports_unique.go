package modeltests

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE = `
-- Test for duplicate airport keys
select 
    airport_key, 
    count(*) as duplicate_count 
from dds.dim_airports 
group by airport_key 
having count(*) > 1
`

const COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE = `
select count(*) as test_count from
(
-- Test for duplicate airport keys
select 
    airport_key, 
    count(*) as duplicate_count 
from dds.dim_airports 
group by airport_key 
having count(*) > 1
) having test_count > 0 limit 1
`


var ddsTestDimAirportsUniqueTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"dds.test_dim_airports_unique",
	RawSQL: 			RAW_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE,
	CountTestSQL: 		COUNT_TEST_SQL_DDS_TEST_DIM_AIRPORTS_UNIQUE,
	TestProfile: 		&configs.TestProfile {
		Name: 				"dds.test_dim_airports_unique",
		Stage: 				"dds",
		Connection: 		"default",
	},
}

var ddsTestDimAirportsUniqueSimpleTestCase = processing.InitSQLModelTesting(ddsTestDimAirportsUniqueTestDescriptor)
