

package assets

import (
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_STAGING_STG_FLIGHTS = `
select
    flight_id,
    flight_number,
    route_id,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    status
from read_csv('store/flights.csv',
    delim = ',',
    header = true,
    columns = {
        'flight_id': 'INT',
        'flight_number': 'VARCHAR',
        'route_id': 'INT',
        'aircraft_type': 'VARCHAR',
        'scheduled_departure': 'TIMESTAMP',
        'scheduled_arrival': 'TIMESTAMP',
        'actual_departure': 'TIMESTAMP',
        'actual_arrival': 'TIMESTAMP',
        'status': 'VARCHAR'
    }
)
`


const SQL_STAGING_STG_FLIGHTS_CREATE_TABLE = `
create table staging.stg_flights
as (select
    flight_id,
    flight_number,
    route_id,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    status
from read_csv('store/flights.csv',
    delim = ',',
    header = true,
    columns = {
        'flight_id': 'INT',
        'flight_number': 'VARCHAR',
        'route_id': 'INT',
        'aircraft_type': 'VARCHAR',
        'scheduled_departure': 'TIMESTAMP',
        'scheduled_arrival': 'TIMESTAMP',
        'actual_departure': 'TIMESTAMP',
        'actual_arrival': 'TIMESTAMP',
        'status': 'VARCHAR'
    }
));

`
const SQL_STAGING_STG_FLIGHTS_INSERT = `
insert into staging.stg_flights ({{ ModelFields }}) (select
    flight_id,
    flight_number,
    route_id,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    status
from read_csv('store/flights.csv',
    delim = ',',
    header = true,
    columns = {
        'flight_id': 'INT',
        'flight_number': 'VARCHAR',
        'route_id': 'INT',
        'aircraft_type': 'VARCHAR',
        'scheduled_departure': 'TIMESTAMP',
        'scheduled_arrival': 'TIMESTAMP',
        'actual_departure': 'TIMESTAMP',
        'actual_arrival': 'TIMESTAMP',
        'status': 'VARCHAR'
    }
))
`
const SQL_STAGING_STG_FLIGHTS_DROP_TABLE = `
drop table staging.stg_flights
`
const SQL_STAGING_STG_FLIGHTS_TRUNCATE = `
delete from staging.stg_flights where true;
truncate table staging.stg_flights;
`




var stagingStgFlightsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"staging.stg_flights",
	RawSQL: 			RAW_SQL_STAGING_STG_FLIGHTS,

	CreateTableSQL: 	SQL_STAGING_STG_FLIGHTS_CREATE_TABLE,
	InsertSQL: 			SQL_STAGING_STG_FLIGHTS_INSERT,
	DropTableSQL: 		SQL_STAGING_STG_FLIGHTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_STAGING_STG_FLIGHTS_TRUNCATE,


	Upstreams: []string {

	},
	Downstreams: []string {

		"dds.fact_flights",

	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"stg_flights",
		Description: 		`IyMgRmxpZ2h0IE9wZXJhdGlvbnMgU3RhZ2luZwoKKipQdXJwb3NlKio6IENvcmUgb3BlcmF0aW9uYWwgZGF0YSBmb3IgZmxpZ2h0IHBlcmZvcm1hbmNlIHRyYWNraW5nCgoqKlRlbXBvcmFsIERhdGEqKjoKLSBgc2NoZWR1bGVkX2RlcGFydHVyZS9hcnJpdmFsYDogUGxhbm5lZCB0aW1lcwotIGBhY3R1YWxfZGVwYXJ0dXJlL2Fycml2YWxgOiBSZWFsIGV4ZWN1dGlvbiB0aW1lcwotIFN0YXR1cyB0cmFja2luZyAoc2NoZWR1bGVkLCBjb21wbGV0ZWQsIGNhbmNlbGxlZCkKCioqS2V5IFJlbGF0aW9uc2hpcHMqKjoKLSBMaW5rcyB0byByb3V0ZXMgdmlhIGByb3V0ZV9pZGAKLSBBaXJjcmFmdCB0eXBlIGZvciBjYXBhY2l0eSBhbmFseXNpcwoKKipQZXJmb3JtYW5jZSBNZXRyaWNzIEZvdW5kYXRpb24qKjoKLSBEZWxheSBjYWxjdWxhdGlvbnMgKGRlcGFydHVyZSAmIGFycml2YWwpCi0gT24tdGltZSBwZXJmb3JtYW5jZSAoT1RQKQotIFNjaGVkdWxlIGFkaGVyZW5jZQotIEZsaWdodCBkdXJhdGlvbiB2YXJpYW5jZQoKKipEYXRhIFF1YWxpdHkqKjogT25seSBjb21wbGV0ZWQvY2FuY2VsbGVkIGZsaWdodHMgZm9yIGFjY3VyYWN5`,
		Stage: 				"staging",
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {

		},
	},
}

var stagingStgFlightsAsset processing.Asset = processing.InitSQLModelAsset(stagingStgFlightsModelDescriptor)
