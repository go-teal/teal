
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_STAGING_STG_AIRPORTS = `


select
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    '{{ TaskID }}' as task_id,
    '{{ TaskUUID }}' as task_uuid
from read_csv('store/airports.csv',
    delim = ',',
    header = true,
    columns = {
        'airport_code': 'VARCHAR',
        'airport_name': 'VARCHAR',
        'city': 'VARCHAR',
        'country': 'VARCHAR',
        'latitude': 'DOUBLE',
        'longitude': 'DOUBLE',
        'timezone': 'VARCHAR'
    }
)
`
const SQL_STAGING_STG_AIRPORTS_CREATE_TABLE = `
create table staging.stg_airports 
as (

select
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    '{{ TaskID }}' as task_id,
    '{{ TaskUUID }}' as task_uuid
from read_csv('store/airports.csv',
    delim = ',',
    header = true,
    columns = {
        'airport_code': 'VARCHAR',
        'airport_name': 'VARCHAR',
        'city': 'VARCHAR',
        'country': 'VARCHAR',
        'latitude': 'DOUBLE',
        'longitude': 'DOUBLE',
        'timezone': 'VARCHAR'
    }
));

`
const SQL_STAGING_STG_AIRPORTS_INSERT = `
insert into staging.stg_airports ({{ ModelFields }}) (

select
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    '{{ TaskID }}' as task_id,
    '{{ TaskUUID }}' as task_uuid
from read_csv('store/airports.csv',
    delim = ',',
    header = true,
    columns = {
        'airport_code': 'VARCHAR',
        'airport_name': 'VARCHAR',
        'city': 'VARCHAR',
        'country': 'VARCHAR',
        'latitude': 'DOUBLE',
        'longitude': 'DOUBLE',
        'timezone': 'VARCHAR'
    }
))
`
const SQL_STAGING_STG_AIRPORTS_DROP_TABLE = `
drop table staging.stg_airports
`
const SQL_STAGING_STG_AIRPORTS_TRUNCATE = `
delete from staging.stg_airports where true;
truncate table staging.stg_airports;
`

var stagingStgAirportsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"staging.stg_airports",
	RawSQL: 			RAW_SQL_STAGING_STG_AIRPORTS,
	CreateTableSQL: 	SQL_STAGING_STG_AIRPORTS_CREATE_TABLE,
	InsertSQL: 			SQL_STAGING_STG_AIRPORTS_INSERT,
	DropTableSQL: 		SQL_STAGING_STG_AIRPORTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_STAGING_STG_AIRPORTS_TRUNCATE,	
	Upstreams: []string {
	},
	Downstreams: []string {
		"dds.dim_airports",
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"stg_airports",
		Stage: 				"staging",
		Description: 		`IyMgQWlycG9ydCBTdGFnaW5nIExheWVyCgoqKlB1cnBvc2UqKjogSW5pdGlhbCBpbmdlc3Rpb24gcG9pbnQgZm9yIGFpcnBvcnQgcmVmZXJlbmNlIGRhdGEKCioqS2V5IEZlYXR1cmVzKio6Ci0gTG9hZHMgcmF3IGFpcnBvcnQgZGF0YSBmcm9tIENTViBmaWxlCi0gUHJlc2VydmVzIGdlb2dyYXBoaWMgY29vcmRpbmF0ZXMgKGxhdGl0dWRlL2xvbmdpdHVkZSkKLSBJbmNsdWRlcyB0aW1lem9uZSBpbmZvcm1hdGlvbiBmb3Igc2NoZWR1bGUgbWFuYWdlbWVudAotIEFkZHMgdHJhY2tpbmcgZmllbGRzICh0YXNrX2lkLCB0YXNrX3V1aWQpIGZvciBsaW5lYWdlCgoqKkRvd25zdHJlYW0gVXNhZ2UqKjoKLSBGb3VuZGF0aW9uIGZvciBgZGltX2FpcnBvcnRzYCBkaW1lbnNpb24gdGFibGUKLSBSZXF1aXJlZCBmb3Igcm91dGUgbmV0d29yayBhbmFseXNpcwotIENyaXRpY2FsIGZvciBodWIgcGVyZm9ybWFuY2UgbWV0cmljcwo=`,
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
		},
	},
}

var stagingStgAirportsAsset processing.Asset = processing.InitSQLModelAsset(stagingStgAirportsModelDescriptor)