package modeltests

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_ROOT_TEST_FLIGHT_DELAYS = `


-- Root test to check for unrealistic flight delays
-- Test passes if no flights have delays greater than 24 hours (1440 minutes)

select 
    flight_id,
    flight_number,
    departure_delay_minutes,
    arrival_delay_minutes,
    case 
        when departure_delay_minutes > 1440 then 'Excessive departure delay'
        when arrival_delay_minutes > 1440 then 'Excessive arrival delay'
        when departure_delay_minutes < -120 then 'Departed too early (>2 hours)'
        when arrival_delay_minutes < -120 then 'Arrived too early (>2 hours)'
    end as issue_type
from dds.fact_flights
where 
    departure_delay_minutes > 1440 
    or arrival_delay_minutes > 1440
    or departure_delay_minutes < -120
    or arrival_delay_minutes < -120
`

const COUNT_TEST_SQL_ROOT_TEST_FLIGHT_DELAYS = `
select count(*) as test_count from 
(


-- Root test to check for unrealistic flight delays
-- Test passes if no flights have delays greater than 24 hours (1440 minutes)

select 
    flight_id,
    flight_number,
    departure_delay_minutes,
    arrival_delay_minutes,
    case 
        when departure_delay_minutes > 1440 then 'Excessive departure delay'
        when arrival_delay_minutes > 1440 then 'Excessive arrival delay'
        when departure_delay_minutes < -120 then 'Departed too early (>2 hours)'
        when arrival_delay_minutes < -120 then 'Arrived too early (>2 hours)'
    end as issue_type
from dds.fact_flights
where 
    departure_delay_minutes > 1440 
    or arrival_delay_minutes > 1440
    or departure_delay_minutes < -120
    or arrival_delay_minutes < -120
) having test_count > 0 limit 1
`


var rootTestFlightDelaysTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"root.test_flight_delays",
	RawSQL: 			RAW_SQL_ROOT_TEST_FLIGHT_DELAYS,
	CountTestSQL: 		COUNT_TEST_SQL_ROOT_TEST_FLIGHT_DELAYS,
	TestProfile: 		&configs.TestProfile {
		Name: 				"root.test_flight_delays",
		Stage: 				"root",
		Description: 		`IyMg4pqg77iPIEZsaWdodCBEZWxheSBBbm9tYWx5IERldGVjdGlvbgoKKipUZXN0IFR5cGUqKjogQnVzaW5lc3MgUnVsZSAvIERhdGEgUmVhc29uYWJsZW5lc3MgQ2hlY2sKCioqUHVycG9zZSoqOiBJZGVudGlmeSB1bnJlYWxpc3RpYyBvcGVyYXRpb25hbCBkYXRhCgoqKlZhbGlkYXRpb24gVGhyZXNob2xkcyoqOgp8IENvbmRpdGlvbiB8IFRocmVzaG9sZCB8IEZsYWcgUmVhc29uIHwKfC0tLS0tLS0tLS0tfC0tLS0tLS0tLS0tfC0tLS0tLS0tLS0tLS18CnwgRXhjZXNzaXZlIERlcGFydHVyZSBEZWxheSB8ID4gMTQ0MCBtaW4gKDI0aCkgfCBMaWtlbHkgZGF0YSBlcnJvciB8CnwgRXhjZXNzaXZlIEFycml2YWwgRGVsYXkgfCA+IDE0NDAgbWluICgyNGgpIHwgU3lzdGVtIGdsaXRjaCB8CnwgVG9vIEVhcmx5IERlcGFydHVyZSB8IDwgLTEyMCBtaW4gKDJoKSB8IFdyb25nIGRhdGUvdGltZSB8CnwgVG9vIEVhcmx5IEFycml2YWwgfCA8IC0xMjAgbWluICgyaCkgfCBUaW1lem9uZSBlcnJvciB8CgoqKkJ1c2luZXNzIEltcGFjdCBvZiBCYWQgRGF0YSoqOgotIPCfk4ogKipLUElzKio6IFNrZXdlZCBPVFAgbWV0cmljcwotIPCfkrAgKipDb21wZW5zYXRpb24qKjogRmFsc2UgZGVsYXkgY2xhaW1zCi0g8J+TiCAqKkZvcmVjYXN0aW5nKio6IEluY29ycmVjdCBwcmVkaWN0aW9ucwotIPCfjq8gKipUYXJnZXRzKio6IFVucmVhbGlzdGljIGdvYWxzCgoqKkNvbW1vbiBSb290IENhdXNlcyoqOgpgYGAKMS4gVGltZXpvbmUgY29udmVyc2lvbiBlcnJvcnMKMi4gRGF0ZSByb2xsb3ZlciBpc3N1ZXMKMy4gTWFudWFsIGRhdGEgZW50cnkgbWlzdGFrZXMKNC4gU3lzdGVtIGludGVncmF0aW9uIGZhaWx1cmVzCmBgYAoKKipBY3Rpb24gSXRlbXMqKjoKLSBGaWx0ZXIgb3V0bGllcnMgZnJvbSBhbmFseXRpY3MKLSBBbGVydCBvcGVyYXRpb25zIHRlYW0KLSBJbnZlc3RpZ2F0ZSBzb3VyY2Ugc3lzdGVtcwotIEFkZCBkYXRhIHZhbGlkYXRpb24gYXQgaW5nZXN0aW9u`,
		Connection: 		"default",
	},
}

var rootTestFlightDelaysSimpleTestCase = processing.InitSQLModelTesting(rootTestFlightDelaysTestDescriptor)