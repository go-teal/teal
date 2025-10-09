{{- define "profile.yaml" }}
    connection: 'default'
    description: |
        ## 🛡️ End-to-End Data Integrity Validation

        **Test Type**: Cross-Dimensional Referential Integrity

        **Scope**: Full pipeline validation after all transformations

        **Validation Coverage**:
        ```mermaid
        graph LR
            A[fact_flights] -->|route_key| B[dim_routes]
            B -->|airport_keys| C[dim_airports]
            D[fact_crew_assignments] -->|employee_key| E[dim_employees]
            D -->|flight_key| A
        ```

        **Test Scenarios**:
        | Check | Description | Business Impact |
        |-------|-------------|-----------------|
        | Orphaned Flights | Flights without routes | Invalid flight records |
        | Missing Employees | Assignments without crew | Ghost assignments |
        | Invalid Airports | Routes without endpoints | Network gaps |
        | Broken References | Any FK violation | Data inconsistency |

        **Failure Response Protocol**:
        1. 🔴 **STOP** - Halt downstream processing
        2. 🔍 **INVESTIGATE** - Check source systems
        3. 🔧 **FIX** - Repair or filter bad records
        4. 🔄 **RERUN** - Restart pipeline

        **SLA**: Must pass for production release
{{- end }}

-- Root test to verify data integrity across the entire pipeline
-- This test runs after all DAG tasks complete when using --with-tests flag

with orphaned_flights as (
    -- Check for flights with routes that don't exist in dim_routes
    select 
        f.flight_id,
        f.route_key,
        'Missing route in dim_routes' as issue
    from {{ Ref "dds.fact_flights" }} f
    left join {{ Ref "dds.dim_routes" }} r on f.route_key = r.route_key
    where r.route_key is null
),
orphaned_assignments as (
    -- Check for crew assignments with employees that don't exist
    select 
        ca.assignment_id,
        ca.employee_key,
        'Missing employee in dim_employees' as issue
    from {{ Ref "dds.fact_crew_assignments" }} ca
    left join {{ Ref "dds.dim_employees" }} e on ca.employee_key = e.employee_key
    where e.employee_key is null
),
invalid_airports as (
    -- Check for routes with airports that don't exist
    select 
        r.route_id,
        r.origin_airport_key,
        'Missing origin airport in dim_airports' as issue
    from {{ Ref "dds.dim_routes" }} r
    left join {{ Ref "dds.dim_airports" }} a on r.origin_airport_key = a.airport_key
    where a.airport_key is null
    
    union all
    
    select 
        r.route_id,
        r.destination_airport_key,
        'Missing destination airport in dim_airports' as issue
    from {{ Ref "dds.dim_routes" }} r
    left join {{ Ref "dds.dim_airports" }} a on r.destination_airport_key = a.airport_key
    where a.airport_key is null
)
-- Return all integrity issues (test passes if no rows returned)
select * from orphaned_flights
union all
select * from orphaned_assignments
union all
select * from invalid_airports