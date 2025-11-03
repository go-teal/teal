

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_MART_MART_AIRPORT_STATISTICS = `
with airport_departures as (
    select
        f.origin_airport_key as airport_key,
        count(*) as departure_count,
        avg(f.departure_delay_minutes) as avg_departure_delay,
        sum(case when f.departure_delay_minutes > 15 then 1 else 0 end) as delayed_departures
    from dds.fact_flights f
    group by f.origin_airport_key
),
airport_arrivals as (
    select
        f.destination_airport_key as airport_key,
        count(*) as arrival_count,
        avg(f.arrival_delay_minutes) as avg_arrival_delay,
        sum(case when f.arrival_delay_minutes > 15 then 1 else 0 end) as delayed_arrivals
    from dds.fact_flights f
    group by f.destination_airport_key
),
airport_crew as (
    select
        sha256(base_airport::varchar) as airport_key,
        count(*) as based_crew_count,
        avg(salary) as avg_crew_salary,
        sum(salary) as total_crew_salary,
        count(distinct position) as unique_positions
    from dds.dim_employees
    group by base_airport
),
airport_routes as (
    select
        coalesce(o.airport_key, d.airport_key) as airport_key,
        coalesce(o.destinations_served, 0) as destinations_served,
        coalesce(d.origins_served, 0) as origins_served
    from (
        select
            origin_airport_key as airport_key,
            count(distinct destination_airport_key) as destinations_served
        from dds.dim_routes
        group by origin_airport_key
    ) o
    full outer join (
        select
            destination_airport_key as airport_key,
            count(distinct origin_airport_key) as origins_served
        from dds.dim_routes
        group by destination_airport_key
    ) d on o.airport_key = d.airport_key
)
select
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    
    -- Flight operations
    coalesce(dep.departure_count, 0) as total_departures,
    coalesce(arr.arrival_count, 0) as total_arrivals,
    coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) as total_movements,
    
    -- Delay metrics
    round(coalesce(dep.avg_departure_delay, 0), 2) as avg_departure_delay_minutes,
    round(coalesce(arr.avg_arrival_delay, 0), 2) as avg_arrival_delay_minutes,
    coalesce(dep.delayed_departures, 0) as delayed_departure_count,
    coalesce(arr.delayed_arrivals, 0) as delayed_arrival_count,
    
    -- On-time performance
    case 
        when dep.departure_count > 0 
        then round((1 - dep.delayed_departures::float / dep.departure_count) * 100, 2)
        else 100
    end as on_time_departure_pct,
    case 
        when arr.arrival_count > 0 
        then round((1 - arr.delayed_arrivals::float / arr.arrival_count) * 100, 2)
        else 100
    end as on_time_arrival_pct,
    
    -- Network connectivity
    coalesce(ar.destinations_served, 0) + coalesce(ar.origins_served, 0) as total_connections,
    
    -- Crew metrics
    coalesce(crew.based_crew_count, 0) as based_crew_count,
    round(coalesce(crew.avg_crew_salary, 0), 2) as avg_crew_salary,
    round(coalesce(crew.total_crew_salary, 0), 2) as total_crew_cost,
    coalesce(crew.unique_positions, 0) as unique_crew_positions,
    
    -- Hub classification
    case 
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 1000 then 'Major Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 500 then 'Regional Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 100 then 'Focus City'
        else 'Spoke'
    end as hub_type
    
from dds.dim_airports a
left join airport_departures dep on a.airport_key = dep.airport_key
left join airport_arrivals arr on a.airport_key = arr.airport_key
left join airport_crew crew on a.airport_key = crew.airport_key
left join airport_routes ar on a.airport_key = ar.airport_key
group by
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    dep.departure_count,
    arr.arrival_count,
    dep.avg_departure_delay,
    arr.avg_arrival_delay,
    dep.delayed_departures,
    arr.delayed_arrivals,
    ar.destinations_served,
    ar.origins_served,
    crew.based_crew_count,
    crew.avg_crew_salary,
    crew.total_crew_salary,
    crew.unique_positions
order by 
    total_movements desc
`




const SQL_MART_MART_AIRPORT_STATISTICS_CREATE_VIEW = `
create view mart.mart_airport_statistics as (with airport_departures as (
    select
        f.origin_airport_key as airport_key,
        count(*) as departure_count,
        avg(f.departure_delay_minutes) as avg_departure_delay,
        sum(case when f.departure_delay_minutes > 15 then 1 else 0 end) as delayed_departures
    from dds.fact_flights f
    group by f.origin_airport_key
),
airport_arrivals as (
    select
        f.destination_airport_key as airport_key,
        count(*) as arrival_count,
        avg(f.arrival_delay_minutes) as avg_arrival_delay,
        sum(case when f.arrival_delay_minutes > 15 then 1 else 0 end) as delayed_arrivals
    from dds.fact_flights f
    group by f.destination_airport_key
),
airport_crew as (
    select
        sha256(base_airport::varchar) as airport_key,
        count(*) as based_crew_count,
        avg(salary) as avg_crew_salary,
        sum(salary) as total_crew_salary,
        count(distinct position) as unique_positions
    from dds.dim_employees
    group by base_airport
),
airport_routes as (
    select
        coalesce(o.airport_key, d.airport_key) as airport_key,
        coalesce(o.destinations_served, 0) as destinations_served,
        coalesce(d.origins_served, 0) as origins_served
    from (
        select
            origin_airport_key as airport_key,
            count(distinct destination_airport_key) as destinations_served
        from dds.dim_routes
        group by origin_airport_key
    ) o
    full outer join (
        select
            destination_airport_key as airport_key,
            count(distinct origin_airport_key) as origins_served
        from dds.dim_routes
        group by destination_airport_key
    ) d on o.airport_key = d.airport_key
)
select
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    
    -- Flight operations
    coalesce(dep.departure_count, 0) as total_departures,
    coalesce(arr.arrival_count, 0) as total_arrivals,
    coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) as total_movements,
    
    -- Delay metrics
    round(coalesce(dep.avg_departure_delay, 0), 2) as avg_departure_delay_minutes,
    round(coalesce(arr.avg_arrival_delay, 0), 2) as avg_arrival_delay_minutes,
    coalesce(dep.delayed_departures, 0) as delayed_departure_count,
    coalesce(arr.delayed_arrivals, 0) as delayed_arrival_count,
    
    -- On-time performance
    case 
        when dep.departure_count > 0 
        then round((1 - dep.delayed_departures::float / dep.departure_count) * 100, 2)
        else 100
    end as on_time_departure_pct,
    case 
        when arr.arrival_count > 0 
        then round((1 - arr.delayed_arrivals::float / arr.arrival_count) * 100, 2)
        else 100
    end as on_time_arrival_pct,
    
    -- Network connectivity
    coalesce(ar.destinations_served, 0) + coalesce(ar.origins_served, 0) as total_connections,
    
    -- Crew metrics
    coalesce(crew.based_crew_count, 0) as based_crew_count,
    round(coalesce(crew.avg_crew_salary, 0), 2) as avg_crew_salary,
    round(coalesce(crew.total_crew_salary, 0), 2) as total_crew_cost,
    coalesce(crew.unique_positions, 0) as unique_crew_positions,
    
    -- Hub classification
    case 
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 1000 then 'Major Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 500 then 'Regional Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 100 then 'Focus City'
        else 'Spoke'
    end as hub_type
    
from dds.dim_airports a
left join airport_departures dep on a.airport_key = dep.airport_key
left join airport_arrivals arr on a.airport_key = arr.airport_key
left join airport_crew crew on a.airport_key = crew.airport_key
left join airport_routes ar on a.airport_key = ar.airport_key
group by
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    dep.departure_count,
    arr.arrival_count,
    dep.avg_departure_delay,
    arr.avg_arrival_delay,
    dep.delayed_departures,
    arr.delayed_arrivals,
    ar.destinations_served,
    ar.origins_served,
    crew.based_crew_count,
    crew.avg_crew_salary,
    crew.total_crew_salary,
    crew.unique_positions
order by 
    total_movements desc)
`
const SQL_MART_MART_AIRPORT_STATISTICS_DROP_VIEW = `
drop view mart.mart_airport_statistics
`


var martMartAirportStatisticsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"mart.mart_airport_statistics",
	RawSQL: 			RAW_SQL_MART_MART_AIRPORT_STATISTICS,


	CreateViewSQL: 		SQL_MART_MART_AIRPORT_STATISTICS_CREATE_VIEW,
	DropViewSQL: 		SQL_MART_MART_AIRPORT_STATISTICS_DROP_VIEW,

	Upstreams: []string {

		"dds.fact_flights",

		"dds.dim_employees",

		"dds.dim_routes",

		"dds.dim_airports",

	},
	Downstreams: []string {

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"mart_airport_statistics",
		Description: 		`IyMgRXhlY3V0aXZlIEFpcnBvcnQgUGVyZm9ybWFuY2UgRGFzaGJvYXJkCgoqKlB1cnBvc2UqKjogMzYwLWRlZ3JlZSB2aWV3IG9mIGFpcnBvcnQgaHViIG9wZXJhdGlvbnMKCioqQWdncmVnYXRpb24gTGV2ZWwqKjogQWlycG9ydC1sZXZlbCBtZXRyaWNzCgoqKkNvcmUgTWV0cmljcyoqOgp8IE1ldHJpYyB8IERlc2NyaXB0aW9uIHwgQnVzaW5lc3MgSW1wYWN0IHwKfC0tLS0tLS0tfC0tLS0tLS0tLS0tLS18LS0tLS0tLS0tLS0tLS0tLS18CnwgVG90YWwgRmxpZ2h0cyB8IERlcGFydHVyZXMgKyBBcnJpdmFscyB8IENhcGFjaXR5IHV0aWxpemF0aW9uIHwKfCBBdmcgRGVsYXkgfCBNZWFuIGRlbGF5IG1pbnV0ZXMgfCBTZXJ2aWNlIHF1YWxpdHkgfAp8IE9UUCAlIHwgT24tdGltZSBwZXJmb3JtYW5jZSB8IEN1c3RvbWVyIHNhdGlzZmFjdGlvbiB8CnwgQ3JldyBDb3VudCB8IEJhc2VkIGVtcGxveWVlcyB8IFJlc291cmNlIGFsbG9jYXRpb24gfAoKKipTdHJhdGVnaWMgSW5zaWdodHMqKjoKLSDwn4+iIEh1YiBlZmZpY2llbmN5IHJhbmtpbmcKLSDwn5OIIEdyb3d0aCBvcHBvcnR1bml0eSBpZGVudGlmaWNhdGlvbgotIOKaoO+4jyBCb3R0bGVuZWNrIGRldGVjdGlvbgotIPCfkaUgV29ya2ZvcmNlIGRpc3RyaWJ1dGlvbgoKKipVc2UgQ2FzZXMqKjoKMS4gKipDYXBhY2l0eSBQbGFubmluZyoqOiBJZGVudGlmeSBjb25nZXN0ZWQgaHVicwoyLiAqKlNlcnZpY2UgSW1wcm92ZW1lbnQqKjogVGFyZ2V0IGRlbGF5IHJlZHVjdGlvbgozLiAqKlJlc291cmNlIEFsbG9jYXRpb24qKjogQmFsYW5jZSBjcmV3IGRlcGxveW1lbnQKNC4gKipOZXR3b3JrIE9wdGltaXphdGlvbioqOiBIdWIgdnMgc3Bva2UgZGVjaXNpb25zCgoqKlJlZnJlc2gqKjogVmlldyAocmVhbC10aW1lIGZyb20gZmFjdHMp`,
		Stage: 				"mart",
		Connection: 		"default",
		Materialization: 	"view",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

		},
	},
}

var martMartAirportStatisticsAsset processing.Asset = processing.InitSQLModelAsset(martMartAirportStatisticsModelDescriptor)
