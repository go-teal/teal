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
		Connection: 		"default",
	},
}

var rootTestFlightDelaysSimpleTestCase = processing.InitSQLModelTesting(rootTestFlightDelaysTestDescriptor)
