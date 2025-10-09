{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'table'
    description: |
        ## Airport Dimension (SCD Type 1)

        **Purpose**: Master dimension for airport reference data with surrogate keys

        **Key Design**:
        - Surrogate key: SHA256 hash of airport_code
        - Unique constraint on airport_code
        - Warehouse audit columns (dw_created_at, dw_updated_at)

        **Geographic Attributes**:
        - Coordinates (latitude/longitude) for distance calculations
        - Timezone for schedule conversions
        - City/country for regional analysis

        **Business Usage**:
        - ✈️ Hub performance analysis
        - 📍 Route network visualization
        - 🌍 Geographic distribution studies
        - ⏰ Schedule timezone management

        **Quality Tests**: `test_dim_airports_unique`
    primary_key_fields: ['airport_key']
    indexes:
      - name: 'airport_code_idx'
        unique: true
        fields: ['airport_code']
    tests:
      - name: 'dds.test_dim_airports_unique'
{{ end }}

select
    sha256(airport_code::varchar) as airport_key,
    airport_code,
    airport_name,
    city,
    country,
    latitude,
    longitude,
    timezone,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from {{ Ref "staging.stg_airports" }}