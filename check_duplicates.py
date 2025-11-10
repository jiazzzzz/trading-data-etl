import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

# Get latest daily table
cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 1")
table = cursor.fetchone()[0]
print(f'Latest table: {table}')

# Check symbol format in daily table
print('\nSample symbols from daily table:')
cursor.execute(f'SELECT symbol, name FROM {table} LIMIT 10')
for row in cursor.fetchall():
    print(f'  {row[0]} - {row[1]}')

# Check for 300062 specifically
print('\nSymbols matching 300062:')
cursor.execute(f"SELECT symbol, name FROM {table} WHERE symbol LIKE '%300062%'")
for row in cursor.fetchall():
    print(f'  {row[0]} - {row[1]}')

# Check stock_list
print('\nStock 300062 in stock_list:')
cursor.execute("SELECT ts_code, symbol, name FROM stock_list WHERE symbol = '300062'")
for row in cursor.fetchall():
    print(f'  {row[0]} | {row[1]} | {row[2]}')

conn.close()
