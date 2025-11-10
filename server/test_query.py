import sqlite3

conn = sqlite3.connect('../jia-stk.db')
cursor = conn.cursor()

days = 7
min_change = 1
max_change = 100

query = """
WITH recent_dates AS (
    SELECT DISTINCT trade_date 
    FROM stock_history 
    ORDER BY trade_date DESC 
    LIMIT ?
),
stock_changes AS (
    SELECT 
        h1.stock_code,
        h1.stock_name,
        h1.exchange,
        h1.trade_date,
        h1.close,
        h1.open,
        ((h1.close - h1.open) / h1.open * 100) as change_percent
    FROM stock_history h1
    WHERE h1.trade_date IN (SELECT trade_date FROM recent_dates)
        AND h1.open > 0
),
stock_stats AS (
    SELECT 
        stock_code,
        stock_name,
        exchange,
        MAX(change_percent) as max_change,
        MIN(change_percent) as min_change,
        AVG(change_percent) as avg_change,
        COUNT(*) as days_count,
        MAX(trade_date) as latest_date
    FROM stock_changes
    GROUP BY stock_code, stock_name, exchange
    HAVING max_change >= ? AND max_change <= ?
)
SELECT 
    s.stock_code,
    s.stock_name,
    s.exchange,
    s.max_change,
    s.min_change,
    s.avg_change,
    s.days_count,
    h.close as latest_price,
    s.latest_date
FROM stock_stats s
LEFT JOIN stock_history h ON s.stock_code = h.stock_code AND s.latest_date = h.trade_date
ORDER BY s.max_change DESC
LIMIT 5
"""

cursor.execute(query, (days, min_change, max_change))
results = cursor.fetchall()

print(f"Found {len(results)} stocks")
for row in results:
    print(f"{row[0]} {row[1]}: max={row[3]:.2f}%, min={row[4]:.2f}%, avg={row[5]:.2f}%, price={row[7]:.2f}")

conn.close()
