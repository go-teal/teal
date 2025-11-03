

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
		Description: 		`IyMgQWlycG9ydCBEaW1lbnNpb24gKFNDRCBUeXBlIDEpCioqUHVycG9zZSoqOiBNYXN0ZXIgZGltZW5zaW9uIGZvciBhaXJwb3J0IHJlZmVyZW5jZSBkYXRhIHdpdGggc3Vycm9nYXRlIGtleXMKCioqS2V5IERlc2lnbioqOgotIFN1cnJvZ2F0ZSBrZXk6IFNIQTI1NiBoYXNoIG9mIGFpcnBvcnRfY29kZQotIFVuaXF1ZSBjb25zdHJhaW50IG9uIGFpcnBvcnRfY29kZQotIFdhcmVob3VzZSBhdWRpdCBjb2x1bW5zIChkd19jcmVhdGVkX2F0LCBkd191cGRhdGVkX2F0KQoKKipHZW9ncmFwaGljIEF0dHJpYnV0ZXMqKjoKLSBDb29yZGluYXRlcyAobGF0aXR1ZGUvbG9uZ2l0dWRlKSBmb3IgZGlzdGFuY2UgY2FsY3VsYXRpb25zCi0gVGltZXpvbmUgZm9yIHNjaGVkdWxlIGNvbnZlcnNpb25zCi0gQ2l0eS9jb3VudHJ5IGZvciByZWdpb25hbCBhbmFseXNpcwoKKipCdXNpbmVzcyBVc2FnZSoqOgotIOKciO+4jyBIdWIgcGVyZm9ybWFuY2UgYW5hbHlzaXMKLSDwn5ONIFJvdXRlIG5ldHdvcmsgdmlzdWFsaXphdGlvbgotIPCfjI0gR2VvZ3JhcGhpYyBkaXN0cmlidXRpb24gc3R1ZGllcwotIOKPsCBTY2hlZHVsZSB0aW1lem9uZSBtYW5hZ2VtZW50CgoqKlF1YWxpdHkgVGVzdHMqKjogYHRlc3RfZGltX2FpcnBvcnRzX3VuaXF1ZWAsIGB0ZXN0X2RpbV9haXJwb3J0c19pbnZhbGlkX2NvZGVzYAo=`,
		Stage: 				"dds",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		true,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

			{
				Name: 			"dds.test_dim_airports_unique",
			},

			{
				Name: 			"dds.test_dim_airports_invalid_codes",
			},

		},
	},
}

var ddsDimAirportsAsset processing.Asset = processing.InitSQLModelAsset(ddsDimAirportsModelDescriptor)
