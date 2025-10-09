{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'table'
    description: |
        ## Crew Assignment Fact Table

        **Purpose**: Bridge table for crew-flight many-to-many relationships

        **Grain**: One row per employee-flight assignment

        **Role Classifications**:
        - `Pilot`: Primary flight command
        - `Co-Pilot`: Secondary command/backup
        - `Flight Attendant`: Cabin crew services

        **Key Relationships**:
        ```sql
        fact_crew_assignments
          ├── dim_employees (employee_key)
        ├── fact_flights (flight_key)
          └── dim_routes (route_key)
        ```

        **Metrics Enabled**:
        - 📊 Crew utilization rates
        - ⏱️ Flight hours by employee
        - 👥 Crew composition analysis
        - 📅 Assignment patterns

        **Compliance Tracking**:
        - Duty time limits
        - Rest period requirements
        - Position qualifications
    primary_key_fields: ['assignment_key']
    indexes:
      - name: 'assignment_id_idx'
        unique: true
        fields: ['assignment_id']
      - name: 'flight_employee_idx'
        unique: true
        fields: ['flight_id', 'employee_id']
{{ end }}

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
    from {{ Ref "staging.stg_crew_assignments" }} ca
    inner join {{ Ref "dds.fact_flights" }} f 
        on ca.flight_id = f.flight_id
    inner join {{ Ref "dds.dim_employees" }} e 
        on ca.employee_id = e.employee_id
    inner join {{ Ref "dds.dim_routes" }} r 
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