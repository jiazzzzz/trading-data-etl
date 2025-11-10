import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

print("Stock names in stock_history table:")
cursor.execute("SELECT DISTINCT stock_code, stock_name FROM stock_history WHERE stock_code IN ('000514', '002136', '000007')")
for row in cursor.fetchall():
    print(f"  {row[0]} | '{row[1]}' | Length: {len(row[1])}")

conn.close()
