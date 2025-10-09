
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_STAGING_STG_ROUTES = `


select
    route_id,
    origin_airport,
    destination_airport,
    distance_km,
    average_duration_minutes
from read_csv('store/routes.csv',
    delim = ',',
    header = true,
    columns = {
        'route_id': 'INT',
        'origin_airport': 'VARCHAR',
        'destination_airport': 'VARCHAR',
        'distance_km': 'DOUBLE',
        'average_duration_minutes': 'INT'
    }
)
`
const SQL_STAGING_STG_ROUTES_CREATE_TABLE = `
create table staging.stg_routes 
as (

select
    route_id,
    origin_airport,
    destination_airport,
    distance_km,
    average_duration_minutes
from read_csv('store/routes.csv',
    delim = ',',
    header = true,
    columns = {
        'route_id': 'INT',
        'origin_airport': 'VARCHAR',
        'destination_airport': 'VARCHAR',
        'distance_km': 'DOUBLE',
        'average_duration_minutes': 'INT'
    }
));

`
const SQL_STAGING_STG_ROUTES_INSERT = `
insert into staging.stg_routes ({{ ModelFields }}) (

select
    route_id,
    origin_airport,
    destination_airport,
    distance_km,
    average_duration_minutes
from read_csv('store/routes.csv',
    delim = ',',
    header = true,
    columns = {
        'route_id': 'INT',
        'origin_airport': 'VARCHAR',
        'destination_airport': 'VARCHAR',
        'distance_km': 'DOUBLE',
        'average_duration_minutes': 'INT'
    }
))
`
const SQL_STAGING_STG_ROUTES_DROP_TABLE = `
drop table staging.stg_routes
`
const SQL_STAGING_STG_ROUTES_TRUNCATE = `
delete from staging.stg_routes where true;
truncate table staging.stg_routes;
`

var stagingStgRoutesModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"staging.stg_routes",
	RawSQL: 			RAW_SQL_STAGING_STG_ROUTES,
	CreateTableSQL: 	SQL_STAGING_STG_ROUTES_CREATE_TABLE,
	InsertSQL: 			SQL_STAGING_STG_ROUTES_INSERT,
	DropTableSQL: 		SQL_STAGING_STG_ROUTES_DROP_TABLE,
	TruncateTableSQL: 	SQL_STAGING_STG_ROUTES_TRUNCATE,	
	Upstreams: []string {
	},
	Downstreams: []string {
		"dds.dim_routes",
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"stg_routes",
		Stage: 				"staging",
		Description: 		`IyMgUm91dGUgTmV0d29yayBTdGFnaW5nCgoqKlB1cnBvc2UqKjogRGVmaW5lIGZsaWdodCBuZXR3b3JrIHRvcG9sb2d5IGFuZCBjb25uZWN0aW9ucwoKKipOZXR3b3JrIEVsZW1lbnRzKio6Ci0gT3JpZ2luL2Rlc3RpbmF0aW9uIGFpcnBvcnQgcGFpcnMKLSBEaXN0YW5jZSBtZXRyaWNzIChrbSkKLSBTdGFuZGFyZCBmbGlnaHQgZHVyYXRpb24gKG1pbnV0ZXMpCgoqKkFuYWx5dGljYWwgU3VwcG9ydCoqOgotIFJvdXRlIGNhdGVnb3JpemF0aW9uIChzaG9ydC9tZWRpdW0vbG9uZy1oYXVsKQotIEZ1ZWwgcmVxdWlyZW1lbnQgZXN0aW1hdGlvbnMKLSBOZXR3b3JrIG9wdGltaXphdGlvbiBhbmFseXNpcwotIEh1Yi1hbmQtc3Bva2UgdG9wb2xvZ3kKCioqQnVzaW5lc3MgQXBwbGljYXRpb25zKio6Ci0gUm91dGUgcHJvZml0YWJpbGl0eSBhbmFseXNpcwotIENhcGFjaXR5IHBsYW5uaW5nCi0gU2NoZWR1bGUgb3B0aW1pemF0aW9uCi0gTmV0d29yayBleHBhbnNpb24gZGVjaXNpb25zCg==`,
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
		},
	},
}

var stagingStgRoutesAsset processing.Asset = processing.InitSQLModelAsset(stagingStgRoutesModelDescriptor)