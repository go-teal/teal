

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_FACT_FLIGHTS = `
with flight_staging as (
    select 
        f.*,
        -- Join with route dimension to get route key
        r.route_key,
        r.origin_airport_key,
        r.destination_airport_key,
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance_km,
        r.average_duration_minutes as route_avg_duration,
        r.route_category
    from staging.stg_flights f
    inner join dds.dim_routes r 
        on f.route_id = r.route_id
    where f.status = 'COMPLETED'  -- Only finished flights
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number) as flight_key,
    flight_id,
    flight_number,
    route_id,
    route_key,
    origin_airport_key,
    destination_airport_key,
    route_category,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    date(scheduled_departure) as flight_date,
    extract(year from scheduled_departure) as flight_year,
    extract(month from scheduled_departure) as flight_month,
    extract(dow from scheduled_departure) as flight_day_of_week,
    case 
        when extract(dow from scheduled_departure) in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type,
    status,
    -- Calculate delays in minutes
    case 
        when actual_departure is not null and scheduled_departure is not null 
        then extract(epoch from (actual_departure - scheduled_departure)) / 60
        else 0
    end as departure_delay_minutes,
    case 
        when actual_arrival is not null and scheduled_arrival is not null 
        then extract(epoch from (actual_arrival - scheduled_arrival)) / 60
        else 0
    end as arrival_delay_minutes,
    -- Calculate actual flight duration
    case 
        when actual_arrival is not null and actual_departure is not null 
        then extract(epoch from (actual_arrival - actual_departure)) / 60
        else null
    end as actual_flight_duration_minutes,
    -- Compare with route average
    case 
        when actual_arrival is not null and actual_departure is not null 
        then (extract(epoch from (actual_arrival - actual_departure)) / 60) - route_avg_duration
        else null
    end as duration_variance_minutes,
    -- Performance indicators
    case 
        when extract(epoch from (actual_departure - scheduled_departure)) / 60 <= 15 then true
        else false
    end as on_time_departure,
    case 
        when extract(epoch from (actual_arrival - scheduled_arrival)) / 60 <= 15 then true
        else false
    end as on_time_arrival,
    current_timestamp as dw_created_at
from flight_staging
{% if IsIncremental() %}
where actual_arrival > (select coalesce(max(actual_arrival), '1900-01-01'::timestamp) from {{ this() }})
{% endif %}
`


const SQL_DDS_FACT_FLIGHTS_CREATE_TABLE = `
create table dds.fact_flights
as (with flight_staging as (
    select 
        f.*,
        -- Join with route dimension to get route key
        r.route_key,
        r.origin_airport_key,
        r.destination_airport_key,
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance_km,
        r.average_duration_minutes as route_avg_duration,
        r.route_category
    from staging.stg_flights f
    inner join dds.dim_routes r 
        on f.route_id = r.route_id
    where f.status = 'COMPLETED'  -- Only finished flights
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number) as flight_key,
    flight_id,
    flight_number,
    route_id,
    route_key,
    origin_airport_key,
    destination_airport_key,
    route_category,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    date(scheduled_departure) as flight_date,
    extract(year from scheduled_departure) as flight_year,
    extract(month from scheduled_departure) as flight_month,
    extract(dow from scheduled_departure) as flight_day_of_week,
    case 
        when extract(dow from scheduled_departure) in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type,
    status,
    -- Calculate delays in minutes
    case 
        when actual_departure is not null and scheduled_departure is not null 
        then extract(epoch from (actual_departure - scheduled_departure)) / 60
        else 0
    end as departure_delay_minutes,
    case 
        when actual_arrival is not null and scheduled_arrival is not null 
        then extract(epoch from (actual_arrival - scheduled_arrival)) / 60
        else 0
    end as arrival_delay_minutes,
    -- Calculate actual flight duration
    case 
        when actual_arrival is not null and actual_departure is not null 
        then extract(epoch from (actual_arrival - actual_departure)) / 60
        else null
    end as actual_flight_duration_minutes,
    -- Compare with route average
    case 
        when actual_arrival is not null and actual_departure is not null 
        then (extract(epoch from (actual_arrival - actual_departure)) / 60) - route_avg_duration
        else null
    end as duration_variance_minutes,
    -- Performance indicators
    case 
        when extract(epoch from (actual_departure - scheduled_departure)) / 60 <= 15 then true
        else false
    end as on_time_departure,
    case 
        when extract(epoch from (actual_arrival - scheduled_arrival)) / 60 <= 15 then true
        else false
    end as on_time_arrival,
    current_timestamp as dw_created_at
from flight_staging
{% if IsIncremental() %}
where actual_arrival > (select coalesce(max(actual_arrival), '1900-01-01'::timestamp) from {{ this() }})
{% endif %});
create unique index fact_flights_pkey on dds.fact_flights (flight_key);
create unique index fact_flights_flight_id_idx_idx on dds.fact_flights (flight_id);
create index fact_flights_flight_date_idx_idx on dds.fact_flights (flight_date);
create index fact_flights_actual_arrival_idx_idx on dds.fact_flights (actual_arrival);

`
const SQL_DDS_FACT_FLIGHTS_INSERT = `
insert into dds.fact_flights ({{ ModelFields }}) (with flight_staging as (
    select 
        f.*,
        -- Join with route dimension to get route key
        r.route_key,
        r.origin_airport_key,
        r.destination_airport_key,
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance_km,
        r.average_duration_minutes as route_avg_duration,
        r.route_category
    from staging.stg_flights f
    inner join dds.dim_routes r 
        on f.route_id = r.route_id
    where f.status = 'COMPLETED'  -- Only finished flights
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number) as flight_key,
    flight_id,
    flight_number,
    route_id,
    route_key,
    origin_airport_key,
    destination_airport_key,
    route_category,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    date(scheduled_departure) as flight_date,
    extract(year from scheduled_departure) as flight_year,
    extract(month from scheduled_departure) as flight_month,
    extract(dow from scheduled_departure) as flight_day_of_week,
    case 
        when extract(dow from scheduled_departure) in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type,
    status,
    -- Calculate delays in minutes
    case 
        when actual_departure is not null and scheduled_departure is not null 
        then extract(epoch from (actual_departure - scheduled_departure)) / 60
        else 0
    end as departure_delay_minutes,
    case 
        when actual_arrival is not null and scheduled_arrival is not null 
        then extract(epoch from (actual_arrival - scheduled_arrival)) / 60
        else 0
    end as arrival_delay_minutes,
    -- Calculate actual flight duration
    case 
        when actual_arrival is not null and actual_departure is not null 
        then extract(epoch from (actual_arrival - actual_departure)) / 60
        else null
    end as actual_flight_duration_minutes,
    -- Compare with route average
    case 
        when actual_arrival is not null and actual_departure is not null 
        then (extract(epoch from (actual_arrival - actual_departure)) / 60) - route_avg_duration
        else null
    end as duration_variance_minutes,
    -- Performance indicators
    case 
        when extract(epoch from (actual_departure - scheduled_departure)) / 60 <= 15 then true
        else false
    end as on_time_departure,
    case 
        when extract(epoch from (actual_arrival - scheduled_arrival)) / 60 <= 15 then true
        else false
    end as on_time_arrival,
    current_timestamp as dw_created_at
from flight_staging
{% if IsIncremental() %}
where actual_arrival > (select coalesce(max(actual_arrival), '1900-01-01'::timestamp) from {{ this() }})
{% endif %})
`
const SQL_DDS_FACT_FLIGHTS_DROP_TABLE = `
drop table dds.fact_flights
`
const SQL_DDS_FACT_FLIGHTS_TRUNCATE = `
delete from dds.fact_flights where true;
truncate table dds.fact_flights;
`




var ddsFactFlightsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"dds.fact_flights",
	RawSQL: 			RAW_SQL_DDS_FACT_FLIGHTS,

	CreateTableSQL: 	SQL_DDS_FACT_FLIGHTS_CREATE_TABLE,
	InsertSQL: 			SQL_DDS_FACT_FLIGHTS_INSERT,
	DropTableSQL: 		SQL_DDS_FACT_FLIGHTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_DDS_FACT_FLIGHTS_TRUNCATE,


	Upstreams: []string {

		"staging.stg_flights",

		"dds.dim_routes",

	},
	Downstreams: []string {

		"dds.fact_crew_assignments",

		"mart.mart_airport_statistics",

		"mart.mart_crew_utilization",

		"mart.mart_flight_performance",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"fact_flights",
		Description: 		`IyMgRmxpZ2h0IE9wZXJhdGlvbnMgRmFjdCBUYWJsZSAoSW5jcmVtZW50YWwpCgoqKlB1cnBvc2UqKjogQ2VudHJhbCBmYWN0IGZvciBmbGlnaHQgcGVyZm9ybWFuY2UgYW5kIG9wZXJhdGlvbmFsIG1ldHJpY3MKCioqTG9hZCBTdHJhdGVneSoqOgotICoqSW5jcmVtZW50YWwqKjogT25seSBuZXcgY29tcGxldGVkL2NhbmNlbGxlZCBmbGlnaHRzCi0gKipGaWx0ZXIqKjogYHN0YXR1cyBJTiAoJ0NvbXBsZXRlZCcsICdDYW5jZWxsZWQnKWAKLSAqKlByZXZlbnRzKio6IEluY29tcGxldGUgZGF0YSBhbmQgdXBkYXRlcwoKKipQZXJmb3JtYW5jZSBNZXRyaWNzKio6CmBgYApkZXBhcnR1cmVfZGVsYXkgPSBhY3R1YWxfZGVwYXJ0dXJlIC0gc2NoZWR1bGVkX2RlcGFydHVyZQphcnJpdmFsX2RlbGF5ID0gYWN0dWFsX2Fycml2YWwgLSBzY2hlZHVsZWRfYXJyaXZhbApvbl90aW1lID0gZGVsYXkg4omkIDE1IG1pbnV0ZXMKYGBgCgoqKktleSBQZXJmb3JtYW5jZSBJbmRpY2F0b3JzKio6Ci0g4pyFIE9uLVRpbWUgUGVyZm9ybWFuY2UgKE9UUCkKLSDij7HvuI8gQXZlcmFnZSBkZWxheSBtaW51dGVzCi0g8J+TiiBTY2hlZHVsZSBhZGhlcmVuY2UgcmF0ZQotIOKaoO+4jyBDYW5jZWxsYXRpb24gdHJhY2tpbmcKCioqVGVtcG9yYWwgQW5hbHlzaXMqKjoKLSBIb3VyIG9mIGRheSBwYXR0ZXJucwotIERheSBvZiB3ZWVrIHRyZW5kcwotIFNlYXNvbmFsIHZhcmlhdGlvbnMKCioqRGVwZW5kZW5jaWVzKio6Ci0gUmVxdWlyZXM6IHN0YWdpbmcuc3RnX2ZsaWdodHMsIGRpbV9yb3V0ZXMKLSBTdXBwb3J0czogQWxsIG1hcnQgbGF5ZXIgYW5hbHl0aWNzCg==`,
		Stage: 				"dds",
		Connection: 		"default",
		Materialization: 	"incremental",
		IsDataFramed: 		true,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

		},
	},
}

var ddsFactFlightsAsset processing.Asset = processing.InitSQLModelAsset(ddsFactFlightsModelDescriptor)
