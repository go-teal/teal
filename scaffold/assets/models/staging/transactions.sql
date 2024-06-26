
select
      id,
      created_on as tx_created_on,
      tx_hash,
      currency,
      raw_amount as amount,
      wallet_address,
      index as tx_index,
      is_suspicious
 from  read_csv('store/transactions.csv',
    delim = ',',
    header = true,
    columns = {
      'id': 'INT',
      'created_on': 'DATE',
      'tx_hash': 'VARCHAR',
      'currency': 'VARCHAR',
      'raw_amount': 'VARCHAR',
      'wallet_address': 'VARCHAR',
      'index': 'INT',
      'is_suspicious':'BOOL'}
    )
