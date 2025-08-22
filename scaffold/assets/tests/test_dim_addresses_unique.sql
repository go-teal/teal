{{- define "profile.yaml" }}
    connection: 'default'    
{{- end }}
select count(pk_id) as c from {{ Ref "dds.dim_addresses" }} group by pk_id having c > 1