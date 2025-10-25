
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_MART_MART_FLIGHT_PERFORMANCE = `


select
    -- Route information
    r.origin_airport,
    origin.airport_name as origin_airport_name,
    origin.city as origin_city,
    r.destination_airport,
    dest.airport_name as destination_airport_name,
    dest.city as destination_city,
    r.route_category,
    r.distance_km,
    
    -- Flight metrics
    count(distinct f.flight_id) as total_flights,
    count(distinct f.flight_number) as unique_flight_numbers,
    
    -- Delay metrics
    avg(f.departure_delay_minutes) as avg_departure_delay_minutes,
    avg(f.arrival_delay_minutes) as avg_arrival_delay_minutes,
    max(f.departure_delay_minutes) as max_departure_delay_minutes,
    max(f.arrival_delay_minutes) as max_arrival_delay_minutes,
    
    -- On-time performance (within 15 minutes)
    sum(case when f.departure_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_departure_pct,
    sum(case when f.arrival_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_arrival_pct,
    
    -- Flight duration analysis
    avg(f.actual_flight_duration_minutes) as avg_actual_duration_minutes,
    avg(f.actual_flight_duration_minutes - r.average_duration_minutes) as avg_duration_variance_minutes,
    
    -- Time period analysis
    f.flight_year,
    f.flight_month,
    case 
        when f.flight_day_of_week in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type
    
from dds.fact_flights f
join dds.dim_routes r on f.route_key = r.route_key
join dds.dim_airports origin on f.origin_airport_key = origin.airport_key
join dds.dim_airports dest on f.destination_airport_key = dest.airport_key
group by
    r.origin_airport,
    origin.airport_name,
    origin.city,
    r.destination_airport,
    dest.airport_name,
    dest.city,
    r.route_category,
    r.distance_km,
    f.flight_year,
    f.flight_month,
    f.flight_day_of_week
`
const SQL_MART_MART_FLIGHT_PERFORMANCE_CREATE_VIEW = `
create view mart.mart_flight_performance as (

select
    -- Route information
    r.origin_airport,
    origin.airport_name as origin_airport_name,
    origin.city as origin_city,
    r.destination_airport,
    dest.airport_name as destination_airport_name,
    dest.city as destination_city,
    r.route_category,
    r.distance_km,
    
    -- Flight metrics
    count(distinct f.flight_id) as total_flights,
    count(distinct f.flight_number) as unique_flight_numbers,
    
    -- Delay metrics
    avg(f.departure_delay_minutes) as avg_departure_delay_minutes,
    avg(f.arrival_delay_minutes) as avg_arrival_delay_minutes,
    max(f.departure_delay_minutes) as max_departure_delay_minutes,
    max(f.arrival_delay_minutes) as max_arrival_delay_minutes,
    
    -- On-time performance (within 15 minutes)
    sum(case when f.departure_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_departure_pct,
    sum(case when f.arrival_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_arrival_pct,
    
    -- Flight duration analysis
    avg(f.actual_flight_duration_minutes) as avg_actual_duration_minutes,
    avg(f.actual_flight_duration_minutes - r.average_duration_minutes) as avg_duration_variance_minutes,
    
    -- Time period analysis
    f.flight_year,
    f.flight_month,
    case 
        when f.flight_day_of_week in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type
    
from dds.fact_flights f
join dds.dim_routes r on f.route_key = r.route_key
join dds.dim_airports origin on f.origin_airport_key = origin.airport_key
join dds.dim_airports dest on f.destination_airport_key = dest.airport_key
group by
    r.origin_airport,
    origin.airport_name,
    origin.city,
    r.destination_airport,
    dest.airport_name,
    dest.city,
    r.route_category,
    r.distance_km,
    f.flight_year,
    f.flight_month,
    f.flight_day_of_week)
`
const SQL_MART_MART_FLIGHT_PERFORMANCE_DROP_VIEW = `
drop view mart.mart_flight_performance
`

var martMartFlightPerformanceModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"mart.mart_flight_performance",
	RawSQL: 			RAW_SQL_MART_MART_FLIGHT_PERFORMANCE,
	CreateViewSQL: 		SQL_MART_MART_FLIGHT_PERFORMANCE_CREATE_VIEW,
	DropViewSQL: 		SQL_MART_MART_FLIGHT_PERFORMANCE_DROP_VIEW,	
	Upstreams: []string {
		"dds.fact_flights",
		"dds.dim_routes",
		"dds.dim_airports",
	},
	Downstreams: []string {
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"mart_flight_performance",
		Stage: 				"mart",
		Description: 		`IyMgRmxpZ2h0IFBlcmZvcm1hbmNlIEFuYWx5dGljcwoKKipQdXJwb3NlKio6IFJvdXRlLWxldmVsIG9wZXJhdGlvbmFsIGV4Y2VsbGVuY2UgbW9uaXRvcmluZwoKKipBZ2dyZWdhdGlvbioqOiBSb3V0ZSBhbmQgYWlycG9ydC1wYWlyIHBlcmZvcm1hbmNlIG1ldHJpY3MKCioqUGVyZm9ybWFuY2UgQ2F0ZWdvcmllcyoqOgpgYGAK8J+foiBPbi1UaW1lOiDiiaQgMTUgbWluIGRlbGF5IChJbmR1c3RyeSBTdGFuZGFyZCkK8J+foSBNaW5vciBEZWxheTogMTUtNjAgbWluCvCflLQgTWFqb3IgRGVsYXk6ID4gNjAgbWluCmBgYAoKKipLZXkgUGVyZm9ybWFuY2UgSW5kaWNhdG9ycyoqOgotICoqT1RQIFJhdGUqKjogJSBmbGlnaHRzIG9uLXRpbWUKLSAqKkF2ZyBEZWxheSoqOiBNZWFuIGRlbGF5IG1pbnV0ZXMKLSAqKkRlbGF5IFZhcmlhbmNlKio6IENvbnNpc3RlbmN5IG1lYXN1cmUKLSAqKkNhbmNlbGxhdGlvbiBSYXRlKio6IFNlcnZpY2UgcmVsaWFiaWxpdHkKCioqUm91dGUgQW5hbHlzaXMgTWF0cml4Kio6CnwgTWV0cmljIHwgRXhjZWxsZW50IHwgR29vZCB8IE5lZWRzIEltcHJvdmVtZW50IHwKfC0tLS0tLS0tfC0tLS0tLS0tLS0tfC0tLS0tLXwtLS0tLS0tLS0tLS0tLS0tLS0tfAp8IE9UUCB8ID4gODUlIHwgNzUtODUlIHwgPCA3NSUgfAp8IEF2ZyBEZWxheSB8IDwgMTAgbWluIHwgMTAtMjAgbWluIHwgPiAyMCBtaW4gfAp8IENhbmNlbCBSYXRlIHwgPCAxJSB8IDEtMyUgfCA+IDMlIHwKCioqQWN0aW9uYWJsZSBJbnNpZ2h0cyoqOgotIPCfk4UgU2NoZWR1bGUgcGFkZGluZyByZXF1aXJlbWVudHMKLSDwn5urIFR1cm5hcm91bmQgdGltZSBvcHRpbWl6YXRpb24KLSDwn4yk77iPIFdlYXRoZXIgaW1wYWN0IHBhdHRlcm5zCi0g8J+UpyBNYWludGVuYW5jZSBzY2hlZHVsaW5nCgoqKkRlY2lzaW9uIFN1cHBvcnQqKjogUmVhbC10aW1lIGZvciBvcGVyYXRpb25zIGNlbnRlcgo=`,
		Connection: 		"default",
		Materialization: 	"view",
		IsDataFramed: 		true,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
		},
	},
}

var martMartFlightPerformanceAsset processing.Asset = processing.InitSQLModelAsset(martMartFlightPerformanceModelDescriptor)