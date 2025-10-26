{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Airport Staging Layer

        **Purpose**: Initial ingestion point for airport reference data

        **Key Features**:
        - Loads raw airport data from CSV file
        - Preserves geographic coordinates (latitude/longitude)
        - Includes timezone information for schedule management
        - Adds tracking fields (task_id, task_uuid) for lineage

        **Downstream Usage**:
        - Foundation for `dim_airports` dimension table
        - Required for route network analysis
        - Critical for hub performance metrics
{{ end }}

select
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    '{ TaskID }' as task_id,
    '{ TaskUUID }' as task_uuid
from read_csv('store/airports.csv',
    delim = ',',
    header = true,
    columns = {
        'airport_code': 'VARCHAR',
        'airport_name': 'VARCHAR',
        'city': 'VARCHAR',
        'country': 'VARCHAR',
        'latitude': 'DOUBLE',
        'longitude': 'DOUBLE',
        'timezone': 'VARCHAR'
    }
)