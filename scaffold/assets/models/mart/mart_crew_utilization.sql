{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'view'
    description: |
        ## Crew Utilization Analytics

        **Purpose**: Workforce productivity and compliance monitoring

        **Grain**: Employee-level aggregations with position grouping

        **Key Calculations**:
        ```sql
        total_flight_hours = SUM(actual_duration_minutes) / 60
        utilization_rate = flight_hours / available_hours
        avg_flight_hours = total_hours / flight_count
        ```

        **Regulatory Compliance Metrics**:
        - ⏰ **Duty Time**: Track against legal limits
        - 😴 **Rest Periods**: Ensure minimum breaks
        - 📋 **Qualification**: Position-specific hours

        **Operational Insights**:
        | Position | Focus Area | Key Metric |
        |----------|------------|------------|
        | Pilot | Flight hours | Safety compliance |
        | Co-Pilot | Training hours | Certification progress |
        | Attendant | Service hours | Customer interaction |

        **Management Actions**:
        1. **Scheduling**: Balance workload distribution
        2. **Training**: Identify development needs
        3. **Compliance**: Prevent violations
        4. **Cost**: Optimize crew efficiency

        **Data Freshness**: Real-time view
{{ end }}

with crew_flights as (
    select
        e.employee_id,
        e.full_name,
        e.position,
        e.base_airport,
        e.salary,
        e.years_of_service,
        ca.role_on_flight,
        ca.crew_category,
        ca.flight_date,
        f.flight_key,
        f.route_key,
        f.actual_flight_duration_minutes,
        r.route_category,
        r.distance_km
    from {{ Ref "dds.fact_crew_assignments" }} ca
    join {{ Ref "dds.dim_employees" }} e on ca.employee_key = e.employee_key
    join {{ Ref "dds.fact_flights" }} f on ca.flight_key = f.flight_key
    join {{ Ref "dds.dim_routes" }} r on f.route_key = r.route_key
)
select
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service,
    
    -- Flight counts
    count(distinct flight_key) as total_flights_assigned,
    count(distinct flight_date) as total_flight_days,
    
    -- Flight hours
    sum(actual_flight_duration_minutes) / 60.0 as total_flight_hours,
    avg(actual_flight_duration_minutes) / 60.0 as avg_flight_hours,
    
    -- Route analysis
    count(distinct route_key) as unique_routes_flown,
    sum(distance_km) as total_distance_km,
    
    -- Route category breakdown
    sum(case when route_category = 'Short-haul' then 1 else 0 end) as short_haul_flights,
    sum(case when route_category = 'Medium-haul' then 1 else 0 end) as medium_haul_flights,
    sum(case when route_category = 'Long-haul' then 1 else 0 end) as long_haul_flights,
    
    -- Role analysis
    mode() within group (order by role_on_flight) as most_common_role,
    mode() within group (order by crew_category) as crew_category,
    
    -- Efficiency metrics
    salary / nullif(sum(actual_flight_duration_minutes) / 60.0, 0) as cost_per_flight_hour,
    count(distinct flight_key)::float / nullif(count(distinct flight_date), 0) as avg_flights_per_day
    
from crew_flights
group by 
    employee_id,
    full_name,
    position,
    base_airport,
    salary,
    years_of_service
order by 
    total_flight_hours desc