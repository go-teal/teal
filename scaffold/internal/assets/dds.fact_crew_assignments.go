
package assets

import (	
	"github.com/go-teal/teal/pkg/models"
	"github.com/go-teal/teal/pkg/configs"
	"github.com/go-teal/teal/pkg/processing"
)

const RAW_SQL_DDS_FACT_CREW_ASSIGNMENTS = `


with assignment_staging as (
    select 
        ca.*,
        -- Get flight details from fact_flights
        f.flight_key,
        f.flight_number,
        f.route_key,
        f.origin_airport_key,
        f.destination_airport_key,
        f.flight_date,
        f.scheduled_departure,
        f.route_category,
        f.aircraft_type,
        -- Get employee details from dim_employees
        e.employee_key,
        e.full_name,
        e.position as employee_position,
        e.base_airport,
        e.base_airport_key,
        e.years_of_service,
        -- Get route details from dim_routes
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance
    from staging.stg_crew_assignments ca
    inner join dds.fact_flights f 
        on ca.flight_id = f.flight_id
    inner join dds.dim_employees e 
        on ca.employee_id = e.employee_id
    inner join dds.dim_routes r 
        on f.route_key = r.route_key
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number || '-' || flight_date::varchar || '-' || employee_id::varchar) as assignment_key,
    assignment_id,
    flight_id,
    flight_key,
    route_key,
    origin_airport_key,
    destination_airport_key,
    employee_id,
    employee_key,
    base_airport_key,
    full_name as employee_name,
    employee_position,
    role_on_flight,
    case 
        when role_on_flight in ('Captain', 'First Officer') then 'Cockpit Crew'
        when role_on_flight in ('Flight Attendant', 'Lead Flight Attendant', 'Purser') then 'Cabin Crew'
        else 'Other'
    end as crew_category,
    -- Check if flying from/to base
    case 
        when base_airport = origin_airport or base_airport = destination_airport then true
        else false
    end as involves_base_airport,
    assignment_date,
    flight_date,
    scheduled_departure,
    route_category,
    route_distance,
    aircraft_type,
    years_of_service,
    current_timestamp as dw_created_at
from assignment_staging
`
const SQL_DDS_FACT_CREW_ASSIGNMENTS_CREATE_TABLE = `
create table dds.fact_crew_assignments 
as (

with assignment_staging as (
    select 
        ca.*,
        -- Get flight details from fact_flights
        f.flight_key,
        f.flight_number,
        f.route_key,
        f.origin_airport_key,
        f.destination_airport_key,
        f.flight_date,
        f.scheduled_departure,
        f.route_category,
        f.aircraft_type,
        -- Get employee details from dim_employees
        e.employee_key,
        e.full_name,
        e.position as employee_position,
        e.base_airport,
        e.base_airport_key,
        e.years_of_service,
        -- Get route details from dim_routes
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance
    from staging.stg_crew_assignments ca
    inner join dds.fact_flights f 
        on ca.flight_id = f.flight_id
    inner join dds.dim_employees e 
        on ca.employee_id = e.employee_id
    inner join dds.dim_routes r 
        on f.route_key = r.route_key
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number || '-' || flight_date::varchar || '-' || employee_id::varchar) as assignment_key,
    assignment_id,
    flight_id,
    flight_key,
    route_key,
    origin_airport_key,
    destination_airport_key,
    employee_id,
    employee_key,
    base_airport_key,
    full_name as employee_name,
    employee_position,
    role_on_flight,
    case 
        when role_on_flight in ('Captain', 'First Officer') then 'Cockpit Crew'
        when role_on_flight in ('Flight Attendant', 'Lead Flight Attendant', 'Purser') then 'Cabin Crew'
        else 'Other'
    end as crew_category,
    -- Check if flying from/to base
    case 
        when base_airport = origin_airport or base_airport = destination_airport then true
        else false
    end as involves_base_airport,
    assignment_date,
    flight_date,
    scheduled_departure,
    route_category,
    route_distance,
    aircraft_type,
    years_of_service,
    current_timestamp as dw_created_at
from assignment_staging);
create unique index fact_crew_assignments_pkey on dds.fact_crew_assignments (assignment_key);
create unique index fact_crew_assignments_assignment_id_idx_idx on dds.fact_crew_assignments (assignment_id);
create unique index fact_crew_assignments_flight_employee_idx_idx on dds.fact_crew_assignments (flight_id, employee_id);

`
const SQL_DDS_FACT_CREW_ASSIGNMENTS_INSERT = `
insert into dds.fact_crew_assignments ({{ ModelFields }}) (

with assignment_staging as (
    select 
        ca.*,
        -- Get flight details from fact_flights
        f.flight_key,
        f.flight_number,
        f.route_key,
        f.origin_airport_key,
        f.destination_airport_key,
        f.flight_date,
        f.scheduled_departure,
        f.route_category,
        f.aircraft_type,
        -- Get employee details from dim_employees
        e.employee_key,
        e.full_name,
        e.position as employee_position,
        e.base_airport,
        e.base_airport_key,
        e.years_of_service,
        -- Get route details from dim_routes
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance
    from staging.stg_crew_assignments ca
    inner join dds.fact_flights f 
        on ca.flight_id = f.flight_id
    inner join dds.dim_employees e 
        on ca.employee_id = e.employee_id
    inner join dds.dim_routes r 
        on f.route_key = r.route_key
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number || '-' || flight_date::varchar || '-' || employee_id::varchar) as assignment_key,
    assignment_id,
    flight_id,
    flight_key,
    route_key,
    origin_airport_key,
    destination_airport_key,
    employee_id,
    employee_key,
    base_airport_key,
    full_name as employee_name,
    employee_position,
    role_on_flight,
    case 
        when role_on_flight in ('Captain', 'First Officer') then 'Cockpit Crew'
        when role_on_flight in ('Flight Attendant', 'Lead Flight Attendant', 'Purser') then 'Cabin Crew'
        else 'Other'
    end as crew_category,
    -- Check if flying from/to base
    case 
        when base_airport = origin_airport or base_airport = destination_airport then true
        else false
    end as involves_base_airport,
    assignment_date,
    flight_date,
    scheduled_departure,
    route_category,
    route_distance,
    aircraft_type,
    years_of_service,
    current_timestamp as dw_created_at
from assignment_staging)
`
const SQL_DDS_FACT_CREW_ASSIGNMENTS_DROP_TABLE = `
drop table dds.fact_crew_assignments
`
const SQL_DDS_FACT_CREW_ASSIGNMENTS_TRUNCATE = `
delete from dds.fact_crew_assignments where true;
truncate table dds.fact_crew_assignments;
`

var ddsFactCrewAssignmentsModelDescriptor = &models.SQLModelDescriptor{
	Name: 				"dds.fact_crew_assignments",
	RawSQL: 			RAW_SQL_DDS_FACT_CREW_ASSIGNMENTS,
	CreateTableSQL: 	SQL_DDS_FACT_CREW_ASSIGNMENTS_CREATE_TABLE,
	InsertSQL: 			SQL_DDS_FACT_CREW_ASSIGNMENTS_INSERT,
	DropTableSQL: 		SQL_DDS_FACT_CREW_ASSIGNMENTS_DROP_TABLE,
	TruncateTableSQL: 	SQL_DDS_FACT_CREW_ASSIGNMENTS_TRUNCATE,	
	Upstreams: []string {
		"staging.stg_crew_assignments",
		"dds.fact_flights",
		"dds.dim_employees",
		"dds.dim_routes",
	},
	Downstreams: []string {
		"mart.mart_crew_utilization",
	},
	ModelProfile:  &configs.ModelProfile{
		Name: 				"fact_crew_assignments",
		Stage: 				"dds",
		Description: 		`IyMgQ3JldyBBc3NpZ25tZW50IEZhY3QgVGFibGUKCioqUHVycG9zZSoqOiBCcmlkZ2UgdGFibGUgZm9yIGNyZXctZmxpZ2h0IG1hbnktdG8tbWFueSByZWxhdGlvbnNoaXBzCgoqKkdyYWluKio6IE9uZSByb3cgcGVyIGVtcGxveWVlLWZsaWdodCBhc3NpZ25tZW50CgoqKlJvbGUgQ2xhc3NpZmljYXRpb25zKio6Ci0gYFBpbG90YDogUHJpbWFyeSBmbGlnaHQgY29tbWFuZAotIGBDby1QaWxvdGA6IFNlY29uZGFyeSBjb21tYW5kL2JhY2t1cAotIGBGbGlnaHQgQXR0ZW5kYW50YDogQ2FiaW4gY3JldyBzZXJ2aWNlcwoKKipLZXkgUmVsYXRpb25zaGlwcyoqOgpgYGBzcWwKZmFjdF9jcmV3X2Fzc2lnbm1lbnRzCiAg4pSc4pSA4pSAIGRpbV9lbXBsb3llZXMgKGVtcGxveWVlX2tleSkK4pSc4pSA4pSAIGZhY3RfZmxpZ2h0cyAoZmxpZ2h0X2tleSkKICDilJTilIDilIAgZGltX3JvdXRlcyAocm91dGVfa2V5KQpgYGAKCioqTWV0cmljcyBFbmFibGVkKio6Ci0g8J+TiiBDcmV3IHV0aWxpemF0aW9uIHJhdGVzCi0g4o+x77iPIEZsaWdodCBob3VycyBieSBlbXBsb3llZQotIPCfkaUgQ3JldyBjb21wb3NpdGlvbiBhbmFseXNpcwotIPCfk4UgQXNzaWdubWVudCBwYXR0ZXJucwoKKipDb21wbGlhbmNlIFRyYWNraW5nKio6Ci0gRHV0eSB0aW1lIGxpbWl0cwotIFJlc3QgcGVyaW9kIHJlcXVpcmVtZW50cwotIFBvc2l0aW9uIHF1YWxpZmljYXRpb25zCg==`,
		Connection: 		"default",
		Materialization: 	"table",
		IsDataFramed: 		false,
		PersistInputs: 		false,
		Tests: []*configs.TestProfile {
		},
	},
}

var ddsFactCrewAssignmentsAsset processing.Asset = processing.InitSQLModelAsset(ddsFactCrewAssignmentsModelDescriptor)