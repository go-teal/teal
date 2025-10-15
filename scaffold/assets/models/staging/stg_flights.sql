{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Flight Operations Staging

        **Purpose**: Core operational data for flight performance tracking

        **Temporal Data**:
        - `scheduled_departure/arrival`: Planned times
        - `actual_departure/arrival`: Real execution times
        - Status tracking (scheduled, completed, cancelled)

        **Key Relationships**:
        - Links to routes via `route_id`
        - Aircraft type for capacity analysis

        **Performance Metrics Foundation**:
        - Delay calculations (departure & arrival)
        - On-time performance (OTP)
        - Schedule adherence
        - Flight duration variance

        **Data Quality**: Only completed/cancelled flights for accuracy
{{ end }}

select
    flight_id,
    flight_number,
    route_id,
    aircraft_type,
    scheduled_departure,
    scheduled_arrival,
    actual_departure,
    actual_arrival,
    status
from read_csv('store/flights.csv',
    delim = ',',
    header = true,
    columns = {
        'flight_id': 'INT',
        'flight_number': 'VARCHAR',
        'route_id': 'INT',
        'aircraft_type': 'VARCHAR',
        'scheduled_departure': 'TIMESTAMP',
        'scheduled_arrival': 'TIMESTAMP',
        'actual_departure': 'TIMESTAMP',
        'actual_arrival': 'TIMESTAMP',
        'status': 'VARCHAR'
    }
)