{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Route Dimension - Network Topology

        **Purpose**: Define airline network structure and route characteristics

        **Key Design Pattern**:
        ```
        route_key = SHA256(origin || '-' || destination)
        ```

        **Route Classification**:
        | Category | Distance | Use Case |
        |----------|----------|----------|
        | Short-haul | < 500km | Regional flights |
        | Medium-haul | 500-3000km | Domestic/nearby international |
        | Long-haul | > 3000km | International/transcontinental |

        **Foreign Keys**:
        - `origin_airport_key` → dim_airports
        - `destination_airport_key` → dim_airports

        **Analytics Support**:
        - 🛫 Route performance metrics
        - 📈 Capacity utilization analysis
        - ⚡ Network efficiency optimization
        - 🗺️ Hub-and-spoke topology

        **Quality Tests**: `test_dim_routes_unique`
    primary_key_fields: ['route_key']
    indexes:
      - name: 'route_id_idx'
        unique: true
        fields: ['route_id']
    tests:
      - name: 'dds.test_dim_routes_unique'
{{ end }}

select
    sha256(origin_airport || '-' || destination_airport) as route_key,
    route_id,
    origin_airport,
    sha256(origin_airport::varchar) as origin_airport_key,
    destination_airport,
    sha256(destination_airport::varchar) as destination_airport_key,
    distance_km,
    average_duration_minutes,
    round(average_duration_minutes / 60.0, 2) as average_duration_hours,
    case 
        when distance_km < 500 then 'Short-haul'
        when distance_km < 3000 then 'Medium-haul'
        else 'Long-haul'
    end as route_category,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from {{ Ref "staging.stg_routes" }}