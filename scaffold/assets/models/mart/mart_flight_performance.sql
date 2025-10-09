{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'view'
    description: |
        ## Flight Performance Analytics

        **Purpose**: Route-level operational excellence monitoring

        **Aggregation**: Route and airport-pair performance metrics

        **Performance Categories**:
        ```
        🟢 On-Time: ≤ 15 min delay (Industry Standard)
        🟡 Minor Delay: 15-60 min
        🔴 Major Delay: > 60 min
        ```

        **Key Performance Indicators**:
        - **OTP Rate**: % flights on-time
        - **Avg Delay**: Mean delay minutes
        - **Delay Variance**: Consistency measure
        - **Cancellation Rate**: Service reliability

        **Route Analysis Matrix**:
        | Metric | Excellent | Good | Needs Improvement |
        |--------|-----------|------|-------------------|
        | OTP | > 85% | 75-85% | < 75% |
        | Avg Delay | < 10 min | 10-20 min | > 20 min |
        | Cancel Rate | < 1% | 1-3% | > 3% |

        **Actionable Insights**:
        - 📅 Schedule padding requirements
        - 🛫 Turnaround time optimization
        - 🌤️ Weather impact patterns
        - 🔧 Maintenance scheduling

        **Decision Support**: Real-time for operations center
{{ end }}

select
    -- Route information
    r.origin_airport,
    origin.airport_name as origin_airport_name,
    origin.city as origin_city,
    r.destination_airport,
    dest.airport_name as destination_airport_name,
    dest.city as destination_city,
    r.route_category,
    r.distance_km,
    
    -- Flight metrics
    count(distinct f.flight_id) as total_flights,
    count(distinct f.flight_number) as unique_flight_numbers,
    
    -- Delay metrics
    avg(f.departure_delay_minutes) as avg_departure_delay_minutes,
    avg(f.arrival_delay_minutes) as avg_arrival_delay_minutes,
    max(f.departure_delay_minutes) as max_departure_delay_minutes,
    max(f.arrival_delay_minutes) as max_arrival_delay_minutes,
    
    -- On-time performance (within 15 minutes)
    sum(case when f.departure_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_departure_pct,
    sum(case when f.arrival_delay_minutes <= 15 then 1 else 0 end)::float / count(*) * 100 as on_time_arrival_pct,
    
    -- Flight duration analysis
    avg(f.actual_flight_duration_minutes) as avg_actual_duration_minutes,
    avg(f.actual_flight_duration_minutes - r.average_duration_minutes) as avg_duration_variance_minutes,
    
    -- Time period analysis
    f.flight_year,
    f.flight_month,
    case 
        when f.flight_day_of_week in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type
    
from {{ Ref "dds.fact_flights" }} f
join {{ Ref "dds.dim_routes" }} r on f.route_key = r.route_key
join {{ Ref "dds.dim_airports" }} origin on f.origin_airport_key = origin.airport_key
join {{ Ref "dds.dim_airports" }} dest on f.destination_airport_key = dest.airport_key
group by
    r.origin_airport,
    origin.airport_name,
    origin.city,
    r.destination_airport,
    dest.airport_name,
    dest.city,
    r.route_category,
    r.distance_km,
    f.flight_year,
    f.flight_month,
    f.flight_day_of_week