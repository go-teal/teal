{{ define "profile.yaml" }}
    materialization: 'table'
    description: |
        ## Employee Master Data Staging

        **Purpose**: Central repository for workforce information

        **Data Elements**:
        - Personal details (name, DOB)
        - Position/role classification
        - Compensation (salary)
        - Base airport assignment
        - Employment timeline (hire_date)

        **Derived Metrics Support**:
        - Years of service calculation
        - Age for regulatory compliance
        - Base distribution analysis

        **Critical for**:
        - Crew scheduling optimization
        - Seniority-based assignments
        - Workforce cost analysis
{{ end }}

select
    employee_id,
    first_name,
    last_name,
    date_of_birth,
    position,
    salary,
    hire_date,
    base_airport
from read_csv('store/employees.csv',
    delim = ',',
    header = true,
    columns = {
        'employee_id': 'INT',
        'first_name': 'VARCHAR',
        'last_name': 'VARCHAR',
        'date_of_birth': 'DATE',
        'position': 'VARCHAR',
        'salary': 'DOUBLE',
        'hire_date': 'DATE',
        'base_airport': 'VARCHAR'
    }
)