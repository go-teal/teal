{{ define "profile.yaml" }}
    connection: 'default'
    description: |
        ## 🔍 Route Dimension Uniqueness Test

        **Test Type**: Network Integrity Validation

        **Business Rule**: One route per origin-destination pair

        **Key Generation**:
        ```
        route_key = SHA256(origin || '-' || destination)
        ```

        **Duplicate Impact Analysis**:
        | Area | Impact | Severity |
        |------|--------|----------|
        | Performance Metrics | Split across duplicates | HIGH |
        | Capacity Planning | Underestimated usage | HIGH |
        | Revenue Analysis | Incorrect route profitability | CRITICAL |
        | Network Optimization | Wrong routing decisions | MEDIUM |

        **Common Failure Patterns**:
        - Same route with different route_ids
        - Bidirectional routes counted separately
        - Data integration errors

        **Resolution**: Dedup at source or in staging
{{ end }}

-- Test for duplicate route keys
-- HAVING is part of the duplicate-finding logic
-- Framework automatically wraps this with COUNT(*) check
SELECT
    route_key,
    COUNT(*) as duplicate_count
FROM {{ Ref("dds.dim_routes") }}
GROUP BY route_key
HAVING COUNT(*) > 1