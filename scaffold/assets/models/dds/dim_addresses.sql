{{ define "profile.yaml" }}
    tests:
        - name: "dds.test_dim_addresses_unique"
{{ end }}


SELECT 
    sha256(wallet_address || currency) as pk_id,    
    wallet_address as wallet_address,
    currency
 from {{ Ref "staging.addresses" }}