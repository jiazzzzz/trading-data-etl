import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

print("Checking stock names:")
cursor.execute("SELECT ts_code, symbol, name FROM stock_list WHERE symbol IN ('000514', '002136')")
for row in cursor.fetchall():
    print(f"  {row[0]} | {row[1]} | {row[2]} | Length: {len(row[2])}")

print("\nChecking more stocks with short names:")
cursor.execute("SELECT ts_code, symbol, name, LENGTH(name) as len FROM stock_list WHERE LENGTH(name) <= 2 LIMIT 20")
for row in cursor.fetchall():
    print(f"  {row[0]} | {row[1]} | {row[2]} | Length: {row[3]}")

conn.close()
