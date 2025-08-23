{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'table'  
    persist_inputs: true
{{ end }}

SELECT 
    sha256(wallet_address || currency) as pk_id,    
    wallet_address as wallet_address,
    currency
 from {{ Ref "staging.addresses" }}