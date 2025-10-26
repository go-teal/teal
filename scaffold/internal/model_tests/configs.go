package modeltests

import (
	"github.com/go-teal/teal/pkg/processing"
)

var ProjectTests = map[string] processing.ModelTesting{

	"root.test_data_integrity":rootTestDataIntegritySimpleTestCase,

	"root.test_flight_delays":rootTestFlightDelaysSimpleTestCase,

	"dds.test_dim_airports_unique":ddsTestDimAirportsUniqueSimpleTestCase,

	"dds.test_dim_employees_unique":ddsTestDimEmployeesUniqueSimpleTestCase,

	"dds.test_dim_routes_unique":ddsTestDimRoutesUniqueSimpleTestCase,

}


