{{ define "profile.yaml" }}
    materialization: 'view'
    description: |
        ## Executive Airport Performance Dashboard

        **Purpose**: 360-degree view of airport hub operations

        **Aggregation Level**: Airport-level metrics

        **Core Metrics**:
        | Metric | Description | Business Impact |
        |--------|-------------|-----------------|
        | Total Flights | Departures + Arrivals | Capacity utilization |
        | Avg Delay | Mean delay minutes | Service quality |
        | OTP % | On-time performance | Customer satisfaction |
        | Crew Count | Based employees | Resource allocation |

        **Strategic Insights**:
        - 🏢 Hub efficiency ranking
        - 📈 Growth opportunity identification
        - ⚠️ Bottleneck detection
        - 👥 Workforce distribution

        **Use Cases**:
        1. **Capacity Planning**: Identify congested hubs
        2. **Service Improvement**: Target delay reduction
        3. **Resource Allocation**: Balance crew deployment
        4. **Network Optimization**: Hub vs spoke decisions

        **Refresh**: View (real-time from facts)
{{ end }}

with airport_departures as (
    select
        f.origin_airport_key as airport_key,
        count(*) as departure_count,
        avg(f.departure_delay_minutes) as avg_departure_delay,
        sum(case when f.departure_delay_minutes > 15 then 1 else 0 end) as delayed_departures
    from {{ Ref("dds.fact_flights") }} f
    group by f.origin_airport_key
),
airport_arrivals as (
    select
        f.destination_airport_key as airport_key,
        count(*) as arrival_count,
        avg(f.arrival_delay_minutes) as avg_arrival_delay,
        sum(case when f.arrival_delay_minutes > 15 then 1 else 0 end) as delayed_arrivals
    from {{ Ref("dds.fact_flights") }} f
    group by f.destination_airport_key
),
airport_crew as (
    select
        sha256(base_airport::varchar) as airport_key,
        count(*) as based_crew_count,
        avg(salary) as avg_crew_salary,
        sum(salary) as total_crew_salary,
        count(distinct position) as unique_positions
    from {{ Ref("dds.dim_employees") }}
    group by base_airport
),
airport_routes as (
    select
        coalesce(o.airport_key, d.airport_key) as airport_key,
        coalesce(o.destinations_served, 0) as destinations_served,
        coalesce(d.origins_served, 0) as origins_served
    from (
        select
            origin_airport_key as airport_key,
            count(distinct destination_airport_key) as destinations_served
        from {{ Ref("dds.dim_routes") }}
        group by origin_airport_key
    ) o
    full outer join (
        select
            destination_airport_key as airport_key,
            count(distinct origin_airport_key) as origins_served
        from {{ Ref("dds.dim_routes") }}
        group by destination_airport_key
    ) d on o.airport_key = d.airport_key
)
select
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    
    -- Flight operations
    coalesce(dep.departure_count, 0) as total_departures,
    coalesce(arr.arrival_count, 0) as total_arrivals,
    coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) as total_movements,
    
    -- Delay metrics
    round(coalesce(dep.avg_departure_delay, 0), 2) as avg_departure_delay_minutes,
    round(coalesce(arr.avg_arrival_delay, 0), 2) as avg_arrival_delay_minutes,
    coalesce(dep.delayed_departures, 0) as delayed_departure_count,
    coalesce(arr.delayed_arrivals, 0) as delayed_arrival_count,
    
    -- On-time performance
    case 
        when dep.departure_count > 0 
        then round((1 - dep.delayed_departures::float / dep.departure_count) * 100, 2)
        else 100
    end as on_time_departure_pct,
    case 
        when arr.arrival_count > 0 
        then round((1 - arr.delayed_arrivals::float / arr.arrival_count) * 100, 2)
        else 100
    end as on_time_arrival_pct,
    
    -- Network connectivity
    coalesce(ar.destinations_served, 0) + coalesce(ar.origins_served, 0) as total_connections,
    
    -- Crew metrics
    coalesce(crew.based_crew_count, 0) as based_crew_count,
    round(coalesce(crew.avg_crew_salary, 0), 2) as avg_crew_salary,
    round(coalesce(crew.total_crew_salary, 0), 2) as total_crew_cost,
    coalesce(crew.unique_positions, 0) as unique_crew_positions,
    
    -- Hub classification
    case 
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 1000 then 'Major Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 500 then 'Regional Hub'
        when coalesce(dep.departure_count, 0) + coalesce(arr.arrival_count, 0) > 100 then 'Focus City'
        else 'Spoke'
    end as hub_type
    
from {{ Ref("dds.dim_airports") }} a
left join airport_departures dep on a.airport_key = dep.airport_key
left join airport_arrivals arr on a.airport_key = arr.airport_key
left join airport_crew crew on a.airport_key = crew.airport_key
left join airport_routes ar on a.airport_key = ar.airport_key
group by
    a.airport_code,
    a.airport_name,
    a.city,
    a.country,
    a.timezone,
    dep.departure_count,
    arr.arrival_count,
    dep.avg_departure_delay,
    arr.avg_arrival_delay,
    dep.delayed_departures,
    arr.delayed_arrivals,
    ar.destinations_served,
    ar.origins_served,
    crew.based_crew_count,
    crew.avg_crew_salary,
    crew.total_crew_salary,
    crew.unique_positions
order by 
    total_movements desc