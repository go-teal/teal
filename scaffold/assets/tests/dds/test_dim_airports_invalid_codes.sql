{{ define "profile.yaml" }}
    connection: 'default'
    description: |
        ## ⚠️ Airport Invalid Code Test

        **Test Type**: Data Quality - Format Validation

        **Business Rule**: All airports must have valid 3-character IATA codes

        **Validation Logic**:
        ```sql
        LENGTH(airport_code) != 3 OR airport_code IS NULL OR airport_code = ''
        ```

        **Failure Impact Analysis**:
        | Area | Impact | Severity |
        |------|--------|----------|
        | Flight Booking | Invalid airport selection | CRITICAL |
        | Route Planning | Incorrect routing | HIGH |
        | Industry Standards | Non-compliance with IATA | CRITICAL |
        | Integration | API failures with partners | HIGH |

        **Common Failure Patterns**:
        - Empty or NULL airport codes
        - 2-character domestic codes (e.g., 'AB')
        - 4-character ICAO codes mistakenly used (e.g., 'ABCD')
        - Special characters or spaces

        **Example Violations**:
        ```
        Airport: 'AB' - Too short (2 characters)
        Airport: 'ABCD' - Too long (4 characters)
        Airport: NULL/empty - Missing code
        ```

        **Resolution**: Validate at data ingestion, apply default for private airfields
{{ end }}

-- Test for airports with invalid IATA codes
-- This test will find airports where code is not exactly 3 characters or is NULL/empty
SELECT
    airport_key,
    airport_name,
    airport_code,
    city,
    country,
    CASE
        WHEN airport_code IS NULL OR airport_code = '' THEN 'NULL or empty code'
        WHEN LENGTH(airport_code) < 3 THEN 'Too short (' || LENGTH(airport_code) || ' chars)'
        WHEN LENGTH(airport_code) > 3 THEN 'Too long (' || LENGTH(airport_code) || ' chars)'
    END as violation_type,
    LENGTH(airport_code) as actual_length
FROM {{ Ref("dds.dim_airports") }}
WHERE
    airport_code IS NULL
    OR airport_code = ''
    OR LENGTH(airport_code) != 3
