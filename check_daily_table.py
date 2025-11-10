import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 1")
result = cursor.fetchone()
if result:
    table = result[0]
    print(f'Latest table: {table}')
    
    cursor.execute(f'PRAGMA table_info({table})')
    print('\nColumns:')
    for row in cursor.fetchall():
        print(f'  {row[1]} ({row[2]})')
    
    print('\nSample data:')
    cursor.execute(f'SELECT * FROM {table} LIMIT 3')
    for row in cursor.fetchall():
        print(row)
else:
    print('No daily tables found')

conn.close()
