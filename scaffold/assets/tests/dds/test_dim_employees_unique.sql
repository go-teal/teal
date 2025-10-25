{{ define "profile.yaml" }}
    connection: 'default'
    description: |
        ## 🔍 Employee Dimension Uniqueness Test

        **Test Type**: Data Quality - Primary Key Constraint

        **Business Rule**: Each employee must have exactly one dimension record

        **SQL Logic**:
        ```sql
        GROUP BY employee_key HAVING COUNT(*) > 1
        ```

        **Failure Consequences**:
        - 👥 Double-counting in utilization metrics
        - 💰 Incorrect salary aggregations
        - 📊 Wrong headcount reports
        - ⚠️ Compliance reporting errors

        **Investigation Steps**:
        1. Check source data for duplicate employee_ids
        2. Verify SHA256 hash generation logic
        3. Review incremental load process

        **Success Criteria**: Empty result set
{{ end }}

-- Test for duplicate employee keys
select 
    employee_key, 
    count(*) as duplicate_count 
from {{ Ref("dds.dim_employees") }} 
group by employee_key 
having count(*) > 1