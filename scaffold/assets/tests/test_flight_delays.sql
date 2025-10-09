{{- define "profile.yaml" }}
    connection: 'default'
    description: |
        ## ⚠️ Flight Delay Anomaly Detection

        **Test Type**: Business Rule / Data Reasonableness Check

        **Purpose**: Identify unrealistic operational data

        **Validation Thresholds**:
        | Condition | Threshold | Flag Reason |
        |-----------|-----------|-------------|
        | Excessive Departure Delay | > 1440 min (24h) | Likely data error |
        | Excessive Arrival Delay | > 1440 min (24h) | System glitch |
        | Too Early Departure | < -120 min (2h) | Wrong date/time |
        | Too Early Arrival | < -120 min (2h) | Timezone error |

        **Business Impact of Bad Data**:
        - 📊 **KPIs**: Skewed OTP metrics
        - 💰 **Compensation**: False delay claims
        - 📈 **Forecasting**: Incorrect predictions
        - 🎯 **Targets**: Unrealistic goals

        **Common Root Causes**:
        ```
        1. Timezone conversion errors
        2. Date rollover issues
        3. Manual data entry mistakes
        4. System integration failures
        ```

        **Action Items**:
        - Filter outliers from analytics
        - Alert operations team
        - Investigate source systems
        - Add data validation at ingestion
{{- end }}

-- Root test to check for unrealistic flight delays
-- Test passes if no flights have delays greater than 24 hours (1440 minutes)

select 
    flight_id,
    flight_number,
    departure_delay_minutes,
    arrival_delay_minutes,
    case 
        when departure_delay_minutes > 1440 then 'Excessive departure delay'
        when arrival_delay_minutes > 1440 then 'Excessive arrival delay'
        when departure_delay_minutes < -120 then 'Departed too early (>2 hours)'
        when arrival_delay_minutes < -120 then 'Arrived too early (>2 hours)'
    end as issue_type
from {{ Ref "dds.fact_flights" }}
where 
    departure_delay_minutes > 1440 
    or arrival_delay_minutes > 1440
    or departure_delay_minutes < -120
    or arrival_delay_minutes < -120