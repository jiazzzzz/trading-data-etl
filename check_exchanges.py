import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

cursor.execute("SELECT COUNT(*) FROM stock_list WHERE ts_code LIKE '%.SH'")
sh_count = cursor.fetchone()[0]

cursor.execute("SELECT COUNT(*) FROM stock_list WHERE ts_code LIKE '%.SZ'")
sz_count = cursor.fetchone()[0]

cursor.execute("SELECT COUNT(*) FROM stock_list WHERE ts_code LIKE '%.BJ'")
bj_count = cursor.fetchone()[0]

print(f"SH stocks: {sh_count}")
print(f"SZ stocks: {sz_count}")
print(f"BJ stocks: {bj_count}")

# Show some examples
print("\nSH examples:")
cursor.execute("SELECT ts_code, symbol, name FROM stock_list WHERE ts_code LIKE '%.SH' LIMIT 5")
for row in cursor.fetchall():
    print(f"  {row[0]} | {row[1]} | {row[2]}")

print("\nSZ examples:")
cursor.execute("SELECT ts_code, symbol, name FROM stock_list WHERE ts_code LIKE '%.SZ' LIMIT 5")
for row in cursor.fetchall():
    print(f"  {row[0]} | {row[1]} | {row[2]}")

conn.close()
