{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Employee Dimension (SCD Type 1)

        **Purpose**: Comprehensive workforce dimension with derived attributes

        **Surrogate Key Design**:
        - Primary: SHA256(employee_id)
        - Foreign: SHA256(base_airport) → dim_airports

        **Calculated Fields**:
        - `age`: Current age for compliance
        - `years_of_service`: Seniority calculation
        - `full_name`: Concatenated display name

        **Workforce Analytics**:
        - 👥 Position distribution (pilot, co-pilot, attendant)
        - 📊 Seniority analysis for assignments
        - 💰 Compensation benchmarking
        - 🏠 Base airport crew distribution

        **Regulatory Compliance**:
        - Age restrictions monitoring
        - Service year requirements
        - Base assignment rules

        **Quality Tests**: `test_dim_employees_unique`
    primary_key_fields: ['employee_key']
    indexes:
      - name: 'employee_id_idx'
        unique: true
        fields: ['employee_id']
    tests:
      - name: 'dds.test_dim_employees_unique'
{{ end }}

select
    sha256(employee_id::varchar) as employee_key,
    employee_id,
    first_name,
    last_name,
    first_name || ' ' || last_name as full_name,
    date_of_birth,
    date_part('year', age(current_date, date_of_birth)) as age,
    position,
    salary,
    hire_date,
    date_part('year', age(current_date, hire_date)) as years_of_service,
    base_airport,
    sha256(base_airport::varchar) as base_airport_key,
    current_timestamp as dw_created_at,
    current_timestamp as dw_updated_at
from {{ Ref "staging.stg_employees" }}