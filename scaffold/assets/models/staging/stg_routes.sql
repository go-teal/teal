{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Route Network Staging

        **Purpose**: Define flight network topology and connections

        **Network Elements**:
        - Origin/destination airport pairs
        - Distance metrics (km)
        - Standard flight duration (minutes)

        **Analytical Support**:
        - Route categorization (short/medium/long-haul)
        - Fuel requirement estimations
        - Network optimization analysis
        - Hub-and-spoke topology

        **Business Applications**:
        - Route profitability analysis
        - Capacity planning
        - Schedule optimization
        - Network expansion decisions
{{ end }}

select
    route_id,
    origin_airport,
    destination_airport,
    distance_km,
    average_duration_minutes
from read_csv('store/routes.csv',
    delim = ',',
    header = true,
    columns = {
        'route_id': 'INT',
        'origin_airport': 'VARCHAR',
        'destination_airport': 'VARCHAR',
        'distance_km': 'DOUBLE',
        'average_duration_minutes': 'INT'
    }
)