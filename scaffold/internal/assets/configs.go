
package assets

import "github.com/go-teal/teal/pkg/processing"

var ProjectAssets = map[string] processing.Asset{
	
	"staging.stg_airports":stagingStgAirportsAsset,
	
	"staging.stg_crew_assignments":stagingStgCrewAssignmentsAsset,
	
	"staging.stg_employees":stagingStgEmployeesAsset,
	
	"staging.stg_flights":stagingStgFlightsAsset,
	
	"staging.stg_routes":stagingStgRoutesAsset,
	
	"dds.dim_airports":ddsDimAirportsAsset,
	
	"dds.dim_employees":ddsDimEmployeesAsset,
	
	"dds.dim_routes":ddsDimRoutesAsset,
	
	"dds.fact_crew_assignments":ddsFactCrewAssignmentsAsset,
	
	"dds.fact_flights":ddsFactFlightsAsset,
	
	"mart.mart_airport_statistics":martMartAirportStatisticsAsset,
	
	"mart.mart_crew_utilization":martMartCrewUtilizationAsset,
	
	"mart.mart_flight_performance":martMartFlightPerformanceAsset,
	
}

var DAG = [][]string{
	
	{
		
		"staging.stg_airports",
		
		"staging.stg_crew_assignments",
		
		"staging.stg_flights",
		
		"staging.stg_employees",
		
		"staging.stg_routes",
		
	},
	
	{
		
		"dds.dim_airports",
		
		"dds.dim_employees",
		
		"dds.dim_routes",
		
	},
	
	{
		
		"dds.fact_flights",
		
	},
	
	{
		
		"dds.fact_crew_assignments",
		
		"mart.mart_airport_statistics",
		
		"mart.mart_flight_performance",
		
	},
	
	{
		
		"mart.mart_crew_utilization",
		
	},
	
}
