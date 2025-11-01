{{ define "profile.yaml" }}
    connection: 'default'
    description: |
        ## 🔍 Airport Dimension Uniqueness Test

        **Test Type**: Data Quality - Primary Key Constraint

        **Validation Rule**:
        ```sql
        COUNT(DISTINCT airport_key) = COUNT(*)
        ```

        **Failure Impact**:
        - ❌ Duplicate airports in dimension
        - ❌ Incorrect flight counts by airport
        - ❌ Wrong hub performance metrics
        - ❌ Distorted network analysis

        **Root Causes of Failure**:
        1. Hash collision in key generation
        2. Duplicate airport codes in source
        3. ETL process error

        **Pass Criteria**: Zero rows returned (no duplicates found)
{{ end }}

-- Test for duplicate airport keys
-- HAVING is part of the duplicate-finding logic
-- Framework automatically wraps this with COUNT(*) check
SELECT
    airport_key,
    COUNT(*) as duplicate_count
FROM {{ Ref("dds.dim_airports") }}
GROUP BY airport_key
HAVING COUNT(*) > 1