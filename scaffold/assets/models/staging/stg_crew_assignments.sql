{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'table'
    description: |
        ## Crew Assignment Staging Layer

        **Purpose**: Maps employees to flights with role specifications

        **Key Features**:
        - Establishes many-to-many relationship between flights and crew
        - Tracks role assignments (pilot, co-pilot, flight attendant)
        - Records assignment dates for scheduling analysis

        **Data Model**:
        - Links to: `employee_id`, `flight_id`
        - Primary key: `assignment_id`

        **Business Value**:
        - Enables crew utilization analysis
        - Supports regulatory compliance tracking
        - Foundation for workload distribution metrics
{{ end }}

select
    assignment_id,
    flight_id,
    employee_id,
    role_on_flight,
    assignment_date
from read_csv('store/crew_assignments.csv',
    delim = ',',
    header = true,
    columns = {
        'assignment_id': 'INT',
        'flight_id': 'INT',
        'employee_id': 'INT',
        'role_on_flight': 'VARCHAR',
        'assignment_date': 'DATE'
    }
)