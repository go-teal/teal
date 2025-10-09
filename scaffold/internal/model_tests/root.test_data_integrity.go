package modeltests

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_ROOT_TEST_DATA_INTEGRITY = `


-- Root test to verify data integrity across the entire pipeline
-- This test runs after all DAG tasks complete when using --with-tests flag

with orphaned_flights as (
    -- Check for flights with routes that don't exist in dim_routes
    select 
        f.flight_id,
        f.route_key,
        'Missing route in dim_routes' as issue
    from dds.fact_flights f
    left join dds.dim_routes r on f.route_key = r.route_key
    where r.route_key is null
),
orphaned_assignments as (
    -- Check for crew assignments with employees that don't exist
    select 
        ca.assignment_id,
        ca.employee_key,
        'Missing employee in dim_employees' as issue
    from dds.fact_crew_assignments ca
    left join dds.dim_employees e on ca.employee_key = e.employee_key
    where e.employee_key is null
),
invalid_airports as (
    -- Check for routes with airports that don't exist
    select 
        r.route_id,
        r.origin_airport_key,
        'Missing origin airport in dim_airports' as issue
    from dds.dim_routes r
    left join dds.dim_airports a on r.origin_airport_key = a.airport_key
    where a.airport_key is null
    
    union all
    
    select 
        r.route_id,
        r.destination_airport_key,
        'Missing destination airport in dim_airports' as issue
    from dds.dim_routes r
    left join dds.dim_airports a on r.destination_airport_key = a.airport_key
    where a.airport_key is null
)
-- Return all integrity issues (test passes if no rows returned)
select * from orphaned_flights
union all
select * from orphaned_assignments
union all
select * from invalid_airports
`

const COUNT_TEST_SQL_ROOT_TEST_DATA_INTEGRITY = `
select count(*) as test_count from 
(


-- Root test to verify data integrity across the entire pipeline
-- This test runs after all DAG tasks complete when using --with-tests flag

with orphaned_flights as (
    -- Check for flights with routes that don't exist in dim_routes
    select 
        f.flight_id,
        f.route_key,
        'Missing route in dim_routes' as issue
    from dds.fact_flights f
    left join dds.dim_routes r on f.route_key = r.route_key
    where r.route_key is null
),
orphaned_assignments as (
    -- Check for crew assignments with employees that don't exist
    select 
        ca.assignment_id,
        ca.employee_key,
        'Missing employee in dim_employees' as issue
    from dds.fact_crew_assignments ca
    left join dds.dim_employees e on ca.employee_key = e.employee_key
    where e.employee_key is null
),
invalid_airports as (
    -- Check for routes with airports that don't exist
    select 
        r.route_id,
        r.origin_airport_key,
        'Missing origin airport in dim_airports' as issue
    from dds.dim_routes r
    left join dds.dim_airports a on r.origin_airport_key = a.airport_key
    where a.airport_key is null
    
    union all
    
    select 
        r.route_id,
        r.destination_airport_key,
        'Missing destination airport in dim_airports' as issue
    from dds.dim_routes r
    left join dds.dim_airports a on r.destination_airport_key = a.airport_key
    where a.airport_key is null
)
-- Return all integrity issues (test passes if no rows returned)
select * from orphaned_flights
union all
select * from orphaned_assignments
union all
select * from invalid_airports
) having test_count > 0 limit 1
`


var rootTestDataIntegrityTestDescriptor = &models.SQLModelTestDescriptor{
	Name: 				"root.test_data_integrity",
	RawSQL: 			RAW_SQL_ROOT_TEST_DATA_INTEGRITY,
	CountTestSQL: 		COUNT_TEST_SQL_ROOT_TEST_DATA_INTEGRITY,
	TestProfile: 		&configs.TestProfile {
		Name: 				"root.test_data_integrity",
		Stage: 				"root",
		Description: 		`IyMg8J+boe+4jyBFbmQtdG8tRW5kIERhdGEgSW50ZWdyaXR5IFZhbGlkYXRpb24KCioqVGVzdCBUeXBlKio6IENyb3NzLURpbWVuc2lvbmFsIFJlZmVyZW50aWFsIEludGVncml0eQoKKipTY29wZSoqOiBGdWxsIHBpcGVsaW5lIHZhbGlkYXRpb24gYWZ0ZXIgYWxsIHRyYW5zZm9ybWF0aW9ucwoKKipWYWxpZGF0aW9uIENvdmVyYWdlKio6CmBgYG1lcm1haWQKZ3JhcGggTFIKICAgIEFbZmFjdF9mbGlnaHRzXSAtLT58cm91dGVfa2V5fCBCW2RpbV9yb3V0ZXNdCiAgICBCIC0tPnxhaXJwb3J0X2tleXN8IENbZGltX2FpcnBvcnRzXQogICAgRFtmYWN0X2NyZXdfYXNzaWdubWVudHNdIC0tPnxlbXBsb3llZV9rZXl8IEVbZGltX2VtcGxveWVlc10KICAgIEQgLS0+fGZsaWdodF9rZXl8IEEKYGBgCgoqKlRlc3QgU2NlbmFyaW9zKio6CnwgQ2hlY2sgfCBEZXNjcmlwdGlvbiB8IEJ1c2luZXNzIEltcGFjdCB8CnwtLS0tLS0tfC0tLS0tLS0tLS0tLS18LS0tLS0tLS0tLS0tLS0tLS18CnwgT3JwaGFuZWQgRmxpZ2h0cyB8IEZsaWdodHMgd2l0aG91dCByb3V0ZXMgfCBJbnZhbGlkIGZsaWdodCByZWNvcmRzIHwKfCBNaXNzaW5nIEVtcGxveWVlcyB8IEFzc2lnbm1lbnRzIHdpdGhvdXQgY3JldyB8IEdob3N0IGFzc2lnbm1lbnRzIHwKfCBJbnZhbGlkIEFpcnBvcnRzIHwgUm91dGVzIHdpdGhvdXQgZW5kcG9pbnRzIHwgTmV0d29yayBnYXBzIHwKfCBCcm9rZW4gUmVmZXJlbmNlcyB8IEFueSBGSyB2aW9sYXRpb24gfCBEYXRhIGluY29uc2lzdGVuY3kgfAoKKipGYWlsdXJlIFJlc3BvbnNlIFByb3RvY29sKio6CjEuIPCflLQgKipTVE9QKiogLSBIYWx0IGRvd25zdHJlYW0gcHJvY2Vzc2luZwoyLiDwn5SNICoqSU5WRVNUSUdBVEUqKiAtIENoZWNrIHNvdXJjZSBzeXN0ZW1zCjMuIPCflKcgKipGSVgqKiAtIFJlcGFpciBvciBmaWx0ZXIgYmFkIHJlY29yZHMKNC4g8J+UhCAqKlJFUlVOKiogLSBSZXN0YXJ0IHBpcGVsaW5lCgoqKlNMQSoqOiBNdXN0IHBhc3MgZm9yIHByb2R1Y3Rpb24gcmVsZWFzZQ==`,
		Connection: 		"default",
	},
}

var rootTestDataIntegritySimpleTestCase = processing.InitSQLModelTesting(rootTestDataIntegrityTestDescriptor)