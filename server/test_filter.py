import sqlite3

conn = sqlite3.connect('../jia-stk.db')
cursor = conn.cursor()

# Check if there's data
cursor.execute("SELECT COUNT(*) FROM stock_history WHERE trade_date = '20251107'")
print(f"Records for 20251107: {cursor.fetchone()[0]}")

# Check sample data
cursor.execute("""
    SELECT stock_code, stock_name, trade_date, open, close, 
           ((close - open) / open * 100) as change_pct 
    FROM stock_history 
    WHERE trade_date = '20251107' AND open > 0 
    ORDER BY change_pct DESC 
    LIMIT 5
""")

print("\nTop 5 gainers on 20251107:")
for row in cursor.fetchall():
    print(f"{row[0]} {row[1]}: open={row[3]:.2f}, close={row[4]:.2f}, change={row[5]:.2f}%")

conn.close()
