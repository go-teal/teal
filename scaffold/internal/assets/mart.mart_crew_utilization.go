
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_MART_MART_CREW_UTILIZATION = `


with crew_flights as (
    select
        e.employee_id,
        e.full_name,
        e.position,
        e.base_airport,
        e.salary,
        e.years_of_service,
        ca.role_on_flight,
        ca.crew_category,
        ca.flight_date,
        f.flight_key,
        f.route_key,
        f.actual_flight_duration_minutes,
        r.route_category,
        r.distance_km
    from dds.fact_crew_assignments ca
    join dds.dim_employees e on ca.employee_key = e.employee_key
    join dds.fact_flights f on ca.flight_key = f.flight_key
    join dds.dim_routes r on f.route_key = r.route_key
)
select
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service,
    
    -- Flight counts
    count(distinct flight_key) as total_flights_assigned,
    count(distinct flight_date) as total_flight_days,
    
    -- Flight hours
    sum(actual_flight_duration_minutes) / 60.0 as total_flight_hours,
    avg(actual_flight_duration_minutes) / 60.0 as avg_flight_hours,
    
    -- Route analysis
    count(distinct route_key) as unique_routes_flown,
    sum(distance_km) as total_distance_km,
    
    -- Route category breakdown
    sum(case when route_category = 'Short-haul' then 1 else 0 end) as short_haul_flights,
    sum(case when route_category = 'Medium-haul' then 1 else 0 end) as medium_haul_flights,
    sum(case when route_category = 'Long-haul' then 1 else 0 end) as long_haul_flights,
    
    -- Role analysis
    mode() within group (order by role_on_flight) as most_common_role,
    mode() within group (order by crew_category) as crew_category,
    
    -- Efficiency metrics
    salary / nullif(sum(actual_flight_duration_minutes) / 60.0, 0) as cost_per_flight_hour,
    count(distinct flight_key)::float / nullif(count(distinct flight_date), 0) as avg_flights_per_day
    
from crew_flights
group by 
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service
order by 
    total_flight_hours desc
`
const SQL_MART_MART_CREW_UTILIZATION_CREATE_VIEW = `
create view mart.mart_crew_utilization as (

with crew_flights as (
    select
        e.employee_id,
        e.full_name,
        e.position,
        e.base_airport,
        e.salary,
        e.years_of_service,
        ca.role_on_flight,
        ca.crew_category,
        ca.flight_date,
        f.flight_key,
        f.route_key,
        f.actual_flight_duration_minutes,
        r.route_category,
        r.distance_km
    from dds.fact_crew_assignments ca
    join dds.dim_employees e on ca.employee_key = e.employee_key
    join dds.fact_flights f on ca.flight_key = f.flight_key
    join dds.dim_routes r on f.route_key = r.route_key
)
select
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service,
    
    -- Flight counts
    count(distinct flight_key) as total_flights_assigned,
    count(distinct flight_date) as total_flight_days,
    
    -- Flight hours
    sum(actual_flight_duration_minutes) / 60.0 as total_flight_hours,
    avg(actual_flight_duration_minutes) / 60.0 as avg_flight_hours,
    
    -- Route analysis
    count(distinct route_key) as unique_routes_flown,
    sum(distance_km) as total_distance_km,
    
    -- Route category breakdown
    sum(case when route_category = 'Short-haul' then 1 else 0 end) as short_haul_flights,
    sum(case when route_category = 'Medium-haul' then 1 else 0 end) as medium_haul_flights,
    sum(case when route_category = 'Long-haul' then 1 else 0 end) as long_haul_flights,
    
    -- Role analysis
    mode() within group (order by role_on_flight) as most_common_role,
    mode() within group (order by crew_category) as crew_category,
    
    -- Efficiency metrics
    salary / nullif(sum(actual_flight_duration_minutes) / 60.0, 0) as cost_per_flight_hour,
    count(distinct flight_key)::float / nullif(count(distinct flight_date), 0) as avg_flights_per_day
    
from crew_flights
group by 
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service
order by 
    total_flight_hours desc)
`
const SQL_MART_MART_CREW_UTILIZATION_DROP_VIEW = `
drop view mart.mart_crew_utilization
`

var martMartCrewUtilizationModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"mart.mart_crew_utilization",
	RawSQL: 			RAW_SQL_MART_MART_CREW_UTILIZATION,
	CreateViewSQL: 		SQL_MART_MART_CREW_UTILIZATION_CREATE_VIEW,
	DropViewSQL: 		SQL_MART_MART_CREW_UTILIZATION_DROP_VIEW,	
	Upstreams: []string {
		"dds.fact_crew_assignments",
		"dds.dim_employees",
		"dds.fact_flights",
		"dds.dim_routes",
	},
	Downstreams: []string {
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"mart_crew_utilization",
		Stage: 				"mart",
		Description: 		`IyMgQ3JldyBVdGlsaXphdGlvbiBBbmFseXRpY3MKCioqUHVycG9zZSoqOiBXb3JrZm9yY2UgcHJvZHVjdGl2aXR5IGFuZCBjb21wbGlhbmNlIG1vbml0b3JpbmcKCioqR3JhaW4qKjogRW1wbG95ZWUtbGV2ZWwgYWdncmVnYXRpb25zIHdpdGggcG9zaXRpb24gZ3JvdXBpbmcKCioqS2V5IENhbGN1bGF0aW9ucyoqOgpgYGBzcWwKdG90YWxfZmxpZ2h0X2hvdXJzID0gU1VNKGFjdHVhbF9kdXJhdGlvbl9taW51dGVzKSAvIDYwCnV0aWxpemF0aW9uX3JhdGUgPSBmbGlnaHRfaG91cnMgLyBhdmFpbGFibGVfaG91cnMKYXZnX2ZsaWdodF9ob3VycyA9IHRvdGFsX2hvdXJzIC8gZmxpZ2h0X2NvdW50CmBgYAoKKipSZWd1bGF0b3J5IENvbXBsaWFuY2UgTWV0cmljcyoqOgotIOKPsCAqKkR1dHkgVGltZSoqOiBUcmFjayBhZ2FpbnN0IGxlZ2FsIGxpbWl0cwotIPCfmLQgKipSZXN0IFBlcmlvZHMqKjogRW5zdXJlIG1pbmltdW0gYnJlYWtzCi0g8J+TiyAqKlF1YWxpZmljYXRpb24qKjogUG9zaXRpb24tc3BlY2lmaWMgaG91cnMKCioqT3BlcmF0aW9uYWwgSW5zaWdodHMqKjoKfCBQb3NpdGlvbiB8IEZvY3VzIEFyZWEgfCBLZXkgTWV0cmljIHwKfC0tLS0tLS0tLS18LS0tLS0tLS0tLS0tfC0tLS0tLS0tLS0tLXwKfCBQaWxvdCB8IEZsaWdodCBob3VycyB8IFNhZmV0eSBjb21wbGlhbmNlIHwKfCBDby1QaWxvdCB8IFRyYWluaW5nIGhvdXJzIHwgQ2VydGlmaWNhdGlvbiBwcm9ncmVzcyB8CnwgQXR0ZW5kYW50IHwgU2VydmljZSBob3VycyB8IEN1c3RvbWVyIGludGVyYWN0aW9uIHwKCioqTWFuYWdlbWVudCBBY3Rpb25zKio6CjEuICoqU2NoZWR1bGluZyoqOiBCYWxhbmNlIHdvcmtsb2FkIGRpc3RyaWJ1dGlvbgoyLiAqKlRyYWluaW5nKio6IElkZW50aWZ5IGRldmVsb3BtZW50IG5lZWRzCjMuICoqQ29tcGxpYW5jZSoqOiBQcmV2ZW50IHZpb2xhdGlvbnMKNC4gKipDb3N0Kio6IE9wdGltaXplIGNyZXcgZWZmaWNpZW5jeQoKKipEYXRhIEZyZXNobmVzcyoqOiBSZWFsLXRpbWUgdmlldwo=`,
		Connection: 		"default",
		Materialization: 	"view",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
		},
	},
}

var martMartCrewUtilizationAsset processing.Asset = processing.InitSQLModelAsset(martMartCrewUtilizationModelDescriptor)