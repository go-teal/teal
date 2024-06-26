{{define "profile.yaml"}}
    connection: 'default'
    materialization: 'view'
{{end}}

SELECT * from {{ Ref "dds.fact_transactions" }}


