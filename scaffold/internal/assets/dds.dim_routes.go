

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_DIM_ROUTES = `
select
    sha256(origin_airport || '-' || destination_airport) as route_key,
    route_id,
    origin_airport,
    sha256(origin_airport::varchar) as origin_airport_key,
    destination_airport,
    sha256(destination_airport::varchar) as destination_airport_key,
    distance_km,
    average_duration_minutes,
    round(average_duration_minutes / 60.0, 2) as average_duration_hours,
    case 
        when distance_km < 500 then 'Short-haul'
        when distance_km < 3000 then 'Medium-haul'
        else 'Long-haul'
    end as route_category,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_routes
`


const SQL_DDS_DIM_ROUTES_CREATE_TABLE = `
create table dds.dim_routes
as (select
    sha256(origin_airport || '-' || destination_airport) as route_key,
    route_id,
    origin_airport,
    sha256(origin_airport::varchar) as origin_airport_key,
    destination_airport,
    sha256(destination_airport::varchar) as destination_airport_key,
    distance_km,
    average_duration_minutes,
    round(average_duration_minutes / 60.0, 2) as average_duration_hours,
    case 
        when distance_km < 500 then 'Short-haul'
        when distance_km < 3000 then 'Medium-haul'
        else 'Long-haul'
    end as route_category,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_routes);
create unique index dim_routes_pkey on dds.dim_routes (route_key);
create unique index dim_routes_route_id_idx_idx on dds.dim_routes (route_id);

`
const SQL_DDS_DIM_ROUTES_INSERT = `
insert into dds.dim_routes ({{ ModelFields }}) (select
    sha256(origin_airport || '-' || destination_airport) as route_key,
    route_id,
    origin_airport,
    sha256(origin_airport::varchar) as origin_airport_key,
    destination_airport,
    sha256(destination_airport::varchar) as destination_airport_key,
    distance_km,
    average_duration_minutes,
    round(average_duration_minutes / 60.0, 2) as average_duration_hours,
    case 
        when distance_km < 500 then 'Short-haul'
        when distance_km < 3000 then 'Medium-haul'
        else 'Long-haul'
    end as route_category,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_routes)
`
const SQL_DDS_DIM_ROUTES_DROP_TABLE = `
drop table dds.dim_routes
`
const SQL_DDS_DIM_ROUTES_TRUNCATE = `
delete from dds.dim_routes where true;
truncate table dds.dim_routes;
`




var ddsDimRoutesModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"dds.dim_routes",
	RawSQL: 			RAW_SQL_DDS_DIM_ROUTES,

	CreateTableSQL: 	SQL_DDS_DIM_ROUTES_CREATE_TABLE,
	InsertSQL: 			SQL_DDS_DIM_ROUTES_INSERT,
	DropTableSQL: 		SQL_DDS_DIM_ROUTES_DROP_TABLE,
	TruncateTableSQL: 	SQL_DDS_DIM_ROUTES_TRUNCATE,


	Upstreams: []string {

		"staging.stg_routes",

	},
	Downstreams: []string {

		"dds.fact_crew_assignments",

		"dds.fact_flights",

		"mart.mart_airport_statistics",

		"mart.mart_crew_utilization",

		"mart.mart_flight_performance",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"dim_routes",
		Description: 		`IyMgUm91dGUgRGltZW5zaW9uIC0gTmV0d29yayBUb3BvbG9neQoKKipQdXJwb3NlKio6IERlZmluZSBhaXJsaW5lIG5ldHdvcmsgc3RydWN0dXJlIGFuZCByb3V0ZSBjaGFyYWN0ZXJpc3RpY3MKCioqS2V5IERlc2lnbiBQYXR0ZXJuKio6CmBgYApyb3V0ZV9rZXkgPSBTSEEyNTYob3JpZ2luIHx8ICctJyB8fCBkZXN0aW5hdGlvbikKYGBgCgoqKlJvdXRlIENsYXNzaWZpY2F0aW9uKio6CnwgQ2F0ZWdvcnkgfCBEaXN0YW5jZSB8IFVzZSBDYXNlIHwKfC0tLS0tLS0tLS18LS0tLS0tLS0tLXwtLS0tLS0tLS0tfAp8IFNob3J0LWhhdWwgfCA8IDUwMGttIHwgUmVnaW9uYWwgZmxpZ2h0cyB8CnwgTWVkaXVtLWhhdWwgfCA1MDAtMzAwMGttIHwgRG9tZXN0aWMvbmVhcmJ5IGludGVybmF0aW9uYWwgfAp8IExvbmctaGF1bCB8ID4gMzAwMGttIHwgSW50ZXJuYXRpb25hbC90cmFuc2NvbnRpbmVudGFsIHwKCioqRm9yZWlnbiBLZXlzKio6Ci0gYG9yaWdpbl9haXJwb3J0X2tleWAg4oaSIGRpbV9haXJwb3J0cwotIGBkZXN0aW5hdGlvbl9haXJwb3J0X2tleWAg4oaSIGRpbV9haXJwb3J0cwoKKipBbmFseXRpY3MgU3VwcG9ydCoqOgotIPCfm6sgUm91dGUgcGVyZm9ybWFuY2UgbWV0cmljcwotIPCfk4ggQ2FwYWNpdHkgdXRpbGl6YXRpb24gYW5hbHlzaXMKLSDimqEgTmV0d29yayBlZmZpY2llbmN5IG9wdGltaXphdGlvbgotIPCfl7rvuI8gSHViLWFuZC1zcG9rZSB0b3BvbG9neQoKKipRdWFsaXR5IFRlc3RzKio6IGB0ZXN0X2RpbV9yb3V0ZXNfdW5pcXVlYAo=`,
		Stage: 				"dds",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

			{
				Name: 			"dds.test_dim_routes_unique",
			},

		},
	},
}

var ddsDimRoutesAsset processing.Asset = processing.InitSQLModelAsset(ddsDimRoutesModelDescriptor)
