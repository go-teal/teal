{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'incremental'  
{{ end }}

with source as (
    select 
        sha256( t.tx_hash || t.currency || t.wallet_address) as pk_id,
        t.amount::HUGEINT as amount,
        t.tx_created_on as tx_created_on,
        date_trunc('day',  t.tx_created_on) as tx_date,
        date_trunc('hour', t.tx_created_on) as tx_hour,
        t.tx_hash as tx_hash,
        sha256(t.wallet_address || t.currency) as fk_address_id,        
        t.tx_index as tx_index,
    from {{ Ref "staging.transactions" }} as t
    inner join {{ Ref "dds.dim_addresses"}} as a 
        on a.wallet_address = t.wallet_address 
        and  a.currency = t.currency 
)

select 
    pk_id,
    amount,
    tx_created_on,
    tx_date,
    tx_hour,
    tx_hash,
    fk_address_id,
    tx_index
 from source

 {{{- if IsIncremental }}}
    where tx_date > (select max(tx_created_on) from {{ this }})
 {{{- end}}}