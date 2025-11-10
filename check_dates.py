import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

# Check daily tables
print("Daily tables (stock_daily_*):")
cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 10")
for row in cursor.fetchall():
    print(f"  {row[0]}")

# Check history dates
print("\nHistory dates (from stock_history):")
cursor.execute("SELECT DISTINCT trade_date FROM stock_history ORDER BY trade_date DESC LIMIT 10")
for row in cursor.fetchall():
    print(f"  {row[0]}")

conn.close()
