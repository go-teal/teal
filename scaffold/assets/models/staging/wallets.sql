{{ define "profile.yaml" }}
    connection: 'default'
    materialization: 'table'  
    is_data_framed: false
{{ end }}

select
      id,    
      name,
      ledger_id,
      wallet_address,
      currency,
      ticker,
      contract_id,
      raw_balance,
      created_at,
      updated_at
    from read_csv('store/wallets.csv',
    delim = ',',
    header = true,
    columns = {
      'id': 'INT',
      'name': 'VARCHAR',
      'ledger_id': 'INT',
      'wallet_address': 'VARCHAR',
      'currency': 'VARCHAR',
      'ticker': 'VARCHAR',
      'contract_id': 'INT',
      'raw_balance': 'VARCHAR',
      'created_at':'DATE',
      'updated_at':'DATE'
      }
    )
