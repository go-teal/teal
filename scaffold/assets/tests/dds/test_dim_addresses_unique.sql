{{- define "profile.yaml" }}
    connection: 'default'    
{{- end }}
select * from {{ Ref "dds.dim_addresses" }} group by pk_id having c > 1