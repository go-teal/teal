{{ define "profile.yaml" }}
    materialization: 'incremental'
    is_data_framed: True
    description: |
        ## Flight Operations Fact Table (Incremental)

        **Purpose**: Central fact for flight performance and operational metrics

        **Load Strategy**:
        - **Incremental**: Only new completed/cancelled flights
        - **Filter**: `status IN ('Completed', 'Cancelled')`
        - **Prevents**: Incomplete data and updates

        **Performance Metrics**:
        ```
        departure_delay = actual_departure - scheduled_departure
        arrival_delay = actual_arrival - scheduled_arrival
        on_time = delay ≤ 15 minutes
        ```

        **Key Performance Indicators**:
        - ✅ On-Time Performance (OTP)
        - ⏱️ Average delay minutes
        - 📊 Schedule adherence rate
        - ⚠️ Cancellation tracking

        **Temporal Analysis**:
        - Hour of day patterns
        - Day of week trends
        - Seasonal variations

        **Dependencies**:
        - Requires: staging.stg_flights, dim_routes
        - Supports: All mart layer analytics
    primary_key_fields: ['flight_key']
    indexes:
      - name: 'flight_id_idx'
        unique: true
        fields: ['flight_id']
      - name: 'flight_date_idx'
        fields: ['flight_date']
      - name: 'actual_arrival_idx'
        fields: ['actual_arrival']
{{ end }}

with flight_staging as (
    select 
        f.*,
        -- Join with route dimension to get route key
        r.route_key,
        r.origin_airport_key,
        r.destination_airport_key,
        r.origin_airport,
        r.destination_airport,
        r.distance_km as route_distance_km,
        r.average_duration_minutes as route_avg_duration,
        r.route_category
    from {{ Ref "staging.stg_flights" }} f
    inner join {{ Ref "dds.dim_routes" }} r 
        on f.route_id = r.route_id
    where f.status = 'COMPLETED'  -- Only finished flights
)
select
    sha256(origin_airport || '-' || destination_airport || '-' || flight_number) as flight_key,
    flight_id,
    flight_number,
    route_id,
    route_key,
    origin_airport_key,
    destination_airport_key,
    route_category,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    date(scheduled_departure) as flight_date,
    extract(year from scheduled_departure) as flight_year,
    extract(month from scheduled_departure) as flight_month,
    extract(dow from scheduled_departure) as flight_day_of_week,
    case 
        when extract(dow from scheduled_departure) in (0, 6) then 'Weekend'
        else 'Weekday'
    end as day_type,
    status,
    -- Calculate delays in minutes
    case 
        when actual_departure is not null and scheduled_departure is not null 
        then extract(epoch from (actual_departure - scheduled_departure)) / 60
        else 0
    end as departure_delay_minutes,
    case 
        when actual_arrival is not null and scheduled_arrival is not null 
        then extract(epoch from (actual_arrival - scheduled_arrival)) / 60
        else 0
    end as arrival_delay_minutes,
    -- Calculate actual flight duration
    case 
        when actual_arrival is not null and actual_departure is not null 
        then extract(epoch from (actual_arrival - actual_departure)) / 60
        else null
    end as actual_flight_duration_minutes,
    -- Compare with route average
    case 
        when actual_arrival is not null and actual_departure is not null 
        then (extract(epoch from (actual_arrival - actual_departure)) / 60) - route_avg_duration
        else null
    end as duration_variance_minutes,
    -- Performance indicators
    case 
        when extract(epoch from (actual_departure - scheduled_departure)) / 60 <= 15 then true
        else false
    end as on_time_departure,
    case 
        when extract(epoch from (actual_arrival - scheduled_arrival)) / 60 <= 15 then true
        else false
    end as on_time_arrival,
    current_timestamp as dw_created_at
from flight_staging
{{{ if IsIncremental }}}
where actual_arrival > (select coalesce(max(actual_arrival), '1900-01-01'::timestamp) from {{ this }})
{{{ end }}}