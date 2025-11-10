import sqlite3

conn = sqlite3.connect('jia-stk.db')
cursor = conn.cursor()

print("Checking stock_history data by board:")
print("=" * 60)

boards = [
    ('600/601/603/605', '沪市主板', "stock_code LIKE '600%' OR stock_code LIKE '601%' OR stock_code LIKE '603%' OR stock_code LIKE '605%'"),
    ('688', '科创板', "stock_code LIKE '688%'"),
    ('000/001', '深主板', "stock_code LIKE '000%' OR stock_code LIKE '001%'"),
    ('002/003', '深证中小板', "stock_code LIKE '002%' OR stock_code LIKE '003%'"),
    ('300/301', '创业板', "stock_code LIKE '300%' OR stock_code LIKE '301%'"),
    ('920/830/430', '北证', "stock_code LIKE '920%' OR stock_code LIKE '830%' OR stock_code LIKE '430%'")
]

# Get latest date
cursor.execute("SELECT DISTINCT trade_date FROM stock_history ORDER BY trade_date DESC LIMIT 1")
latest_date = cursor.fetchone()[0]
print(f"Latest date in stock_history: {latest_date}\n")

for prefix, name, condition in boards:
    # Count total stocks
    cursor.execute(f"SELECT COUNT(DISTINCT stock_code) FROM stock_history WHERE {condition}")
    total_stocks = cursor.fetchone()[0]
    
    # Count stocks with data on latest date
    cursor.execute(f"SELECT COUNT(DISTINCT stock_code) FROM stock_history WHERE ({condition}) AND trade_date = '{latest_date}'")
    latest_stocks = cursor.fetchone()[0]
    
    # Just check if there are stocks on latest date
    cursor.execute(f"""
        SELECT stock_code, stock_name, close
        FROM stock_history
        WHERE ({condition}) 
            AND trade_date = '{latest_date}'
        LIMIT 3
    """)
    samples = cursor.fetchall()
    
    print(f"{prefix} ({name}):")
    print(f"  Total stocks: {total_stocks}")
    print(f"  Stocks on {latest_date}: {latest_stocks}")
    print(f"  Sample stocks:")
    if samples:
        for row in samples:
            print(f"    {row[0]} {row[1]}")
    else:
        print(f"    None found")
    print()

conn.close()
