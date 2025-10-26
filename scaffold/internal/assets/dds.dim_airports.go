

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_DIM_AIRPORTS = `
select
    sha256(airport_code::varchar) as airport_key,
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_airports
`


const SQL_DDS_DIM_AIRPORTS_CREATE_TABLE = `
create table dds.dim_airports
as (select
    sha256(airport_code::varchar) as airport_key,
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_airports);
create unique index dim_airports_pkey on dds.dim_airports (airport_key);
create unique index dim_airports_airport_code_idx_idx on dds.dim_airports (airport_code);

`
const SQL_DDS_DIM_AIRPORTS_INSERT = `
insert into dds.dim_airports ({{ ModelFields }}) (select
    sha256(airport_code::varchar) as airport_key,
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from staging.stg_airports)
`
const SQL_DDS_DIM_AIRPORTS_DROP_TABLE = `
drop table dds.dim_airports
`
const SQL_DDS_DIM_AIRPORTS_TRUNCATE = `
delete from dds.dim_airports where true;
truncate table dds.dim_airports;
`




var ddsDimAirportsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"dds.dim_airports",
	RawSQL: 			RAW_SQL_DDS_DIM_AIRPORTS,

	CreateTableSQL: 	SQL_DDS_DIM_AIRPORTS_CREATE_TABLE,
	InsertSQL: 			SQL_DDS_DIM_AIRPORTS_INSERT,
	DropTableSQL: 		SQL_DDS_DIM_AIRPORTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_DDS_DIM_AIRPORTS_TRUNCATE,


	Upstreams: []string {

		"staging.stg_airports",

	},
	Downstreams: []string {

		"mart.mart_airport_statistics",

		"mart.mart_flight_performance",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"dim_airports",
		Description: 		`IyMgQWlycG9ydCBEaW1lbnNpb24gKFNDRCBUeXBlIDEpCgoqKlB1cnBvc2UqKjogTWFzdGVyIGRpbWVuc2lvbiBmb3IgYWlycG9ydCByZWZlcmVuY2UgZGF0YSB3aXRoIHN1cnJvZ2F0ZSBrZXlzCgoqKktleSBEZXNpZ24qKjoKLSBTdXJyb2dhdGUga2V5OiBTSEEyNTYgaGFzaCBvZiBhaXJwb3J0X2NvZGUKLSBVbmlxdWUgY29uc3RyYWludCBvbiBhaXJwb3J0X2NvZGUKLSBXYXJlaG91c2UgYXVkaXQgY29sdW1ucyAoZHdfY3JlYXRlZF9hdCwgZHdfdXBkYXRlZF9hdCkKCioqR2VvZ3JhcGhpYyBBdHRyaWJ1dGVzKio6Ci0gQ29vcmRpbmF0ZXMgKGxhdGl0dWRlL2xvbmdpdHVkZSkgZm9yIGRpc3RhbmNlIGNhbGN1bGF0aW9ucwotIFRpbWV6b25lIGZvciBzY2hlZHVsZSBjb252ZXJzaW9ucwotIENpdHkvY291bnRyeSBmb3IgcmVnaW9uYWwgYW5hbHlzaXMKCioqQnVzaW5lc3MgVXNhZ2UqKjoKLSDinIjvuI8gSHViIHBlcmZvcm1hbmNlIGFuYWx5c2lzCi0g8J+TjSBSb3V0ZSBuZXR3b3JrIHZpc3VhbGl6YXRpb24KLSDwn4yNIEdlb2dyYXBoaWMgZGlzdHJpYnV0aW9uIHN0dWRpZXMKLSDij7AgU2NoZWR1bGUgdGltZXpvbmUgbWFuYWdlbWVudAoKKipRdWFsaXR5IFRlc3RzKio6IGB0ZXN0X2RpbV9haXJwb3J0c191bmlxdWVgCg==`,
		Stage: 				"dds",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		true,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

			{
				Name: 			"dds.test_dim_airports_unique",
			},

		},
	},
}

var ddsDimAirportsAsset processing.Asset = processing.InitSQLModelAsset(ddsDimAirportsModelDescriptor)
