#!/usr/bin/env python3
"""
Daily Stock Data Update Script
This script should be run daily to:
1. Import TDX historical data from ./tdx folder
2. Update stock list and names from TDX
3. Fetch latest daily trading data from Sina
"""
import sys
sys.path.append('./lib')

from lib.stock_info import StockInfo
from lib.db_sqlite import DbSqlite
from lib.common import Common
from logger import Logger
import pandas as pd
import datetime
import time
import os
import sqlite3
from datetime import datetime as dt

logger = Logger("daily_update")

# ============================================================================
# TDX Historical Data Import Functions
# ============================================================================

def parse_tdx_file(file_path):
    """Parse a single TDX export file."""
    try:
        with open(file_path, 'r', encoding='gbk') as f:
            lines = f.readlines()
        
        if len(lines) < 3:
            return None
        
        # Parse header
        header = lines[0].strip()
        parts = header.split()
        if len(parts) < 2:
            return None
        
        stock_code = parts[0]
        stock_name = parts[1] if len(parts) > 1 else ""
        
        # Determine exchange from filename
        filename = os.path.basename(file_path)
        if filename.startswith('SH#'):
            exchange = 'SH'
        elif filename.startswith('SZ#'):
            exchange = 'SZ'
        elif filename.startswith('BJ#'):
            exchange = 'BJ'
        else:
            exchange = 'UNKNOWN'
        
        # Parse data rows
        data_rows = []
        for line in lines[2:]:
            line = line.strip()
            if not line:
                continue
            
            parts = line.split('\t')
            if len(parts) < 7:
                continue
            
            try:
                date_str = parts[0].strip()
                date_obj = dt.strptime(date_str, '%Y/%m/%d')
                trade_date = date_obj.strftime('%Y%m%d')
                
                data_rows.append({
                    'trade_date': trade_date,
                    'open': float(parts[1]),
                    'high': float(parts[2]),
                    'low': float(parts[3]),
                    'close': float(parts[4]),
                    'volume': int(float(parts[5])),
                    'amount': float(parts[6])
                })
            except (ValueError, IndexError):
                continue
        
        return (stock_code, stock_name, exchange, data_rows)
    
    except Exception as e:
        logger.err(f"Error reading file {file_path}: {str(e)}")
        return None

def create_history_table(db):
    """Create stock_history table if not exists."""
    conn = db.connect()
    cursor = conn.cursor()
    
    try:
        cursor.execute("""
            CREATE TABLE IF NOT EXISTS stock_history (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                stock_code TEXT NOT NULL,
                stock_name TEXT,
                exchange TEXT,
                trade_date TEXT NOT NULL,
                open REAL,
                high REAL,
                low REAL,
                close REAL,
                volume INTEGER,
                amount REAL,
                import_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                UNIQUE(stock_code, trade_date)
            )
        """)
        
        cursor.execute("CREATE INDEX IF NOT EXISTS idx_stock_code ON stock_history(stock_code)")
        cursor.execute("CREATE INDEX IF NOT EXISTS idx_trade_date ON stock_history(trade_date)")
        cursor.execute("CREATE INDEX IF NOT EXISTS idx_stock_date ON stock_history(stock_code, trade_date)")
        
        conn.commit()
    finally:
        conn.close()

def import_tdx_history(db, tdx_folder='./tdx'):
    """Import TDX historical data."""
    logger.info("=" * 60)
    logger.info("Step 1: Importing TDX Historical Data")
    logger.info("=" * 60)
    
    if not os.path.exists(tdx_folder):
        logger.err(f"TDX folder not found: {tdx_folder}")
        return False
    
    create_history_table(db)
    
    tdx_files = [f for f in os.listdir(tdx_folder) if f.endswith('.txt')]
    total_files = len(tdx_files)
    logger.info(f"Found {total_files} TDX files")
    
    success_count = 0
    error_count = 0
    total_records = 0
    
    conn = db.connect()
    cursor = conn.cursor()
    
    for idx, filename in enumerate(tdx_files, 1):
        file_path = os.path.join(tdx_folder, filename)
        
        if idx % 100 == 0:
            logger.info(f"Processing {idx}/{total_files} files...")
        
        result = parse_tdx_file(file_path)
        if not result:
            error_count += 1
            continue
        
        stock_code, stock_name, exchange, data_rows = result
        
        if not data_rows:
            error_count += 1
            continue
        
        try:
            for row in data_rows:
                cursor.execute("""
                    INSERT OR REPLACE INTO stock_history 
                    (stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    stock_code, stock_name, exchange,
                    row['trade_date'], row['open'], row['high'], 
                    row['low'], row['close'], row['volume'], row['amount']
                ))
            
            conn.commit()
            success_count += 1
            total_records += len(data_rows)
            
        except Exception as e:
            logger.err(f"Error importing {filename}: {str(e)}")
            error_count += 1
            conn.rollback()
    
    conn.close()
    
    logger.info(f"✓ Imported {success_count} files, {total_records} records")
    if error_count > 0:
        logger.warn(f"⚠ {error_count} files had errors")
    
    return True

# ============================================================================
# Stock List Update Functions
# ============================================================================

def update_stock_list(db, tdx_folder='./tdx'):
    """Update stock list and names from TDX."""
    logger.info("=" * 60)
    logger.info("Step 2: Updating Stock List from TDX")
    logger.info("=" * 60)
    
    stock_info = StockInfo()
    stock_list = stock_info.get_stock_list_from_tdx(tdx_folder)
    
    if stock_list is None or stock_list.empty:
        logger.err("Failed to retrieve stock list from TDX")
        return False
    
    logger.info(f"Found {len(stock_list)} stocks in TDX")
    
    updated_count = 0
    for index, row in stock_list.iterrows():
        try:
            ts_code = row['ts_code']
            name = row['name']
            name_escaped = name.replace("'", "''")
            
            query = f"UPDATE stock_list SET name = '{name_escaped}' WHERE ts_code = '{ts_code}'"
            db.exec(query)
            updated_count += 1
            
            if updated_count % 500 == 0:
                logger.info(f"Updated {updated_count} stocks...")
                
        except Exception as e:
            logger.err(f"Error updating {ts_code}: {str(e)}")
    
    logger.info(f"✓ Updated {updated_count} stock names")
    return True

# ============================================================================
# Daily Trading Data Functions
# ============================================================================

def insert_pinyin(row):
    """Add pinyin to stock name."""
    out = {}
    com = Common()
    out['pinyin'] = com.get_py_from_name(row['name'])
    return pd.Series(out)

def dump_stock_list(db, tdx_folder='./tdx'):
    """Dump stock list from TDX to database."""
    logger.info("=" * 60)
    logger.info("Step 3: Updating Stock List Table")
    logger.info("=" * 60)
    
    db.drop_table("stock_list")
    db_engine = db.create_engine()
    stock_info = StockInfo()
    stock_list = stock_info.get_stock_list_from_tdx(tdx_folder)
    
    if stock_list is None or stock_list.empty:
        logger.err("Failed to retrieve stock list from TDX")
        return False
    
    stock_list[['pinyin']] = stock_list.apply(insert_pinyin, axis=1)
    stock_list.to_sql("stock_list", db_engine, index=False, if_exists='append', chunksize=5000)
    
    logger.info(f"✓ Stock list updated with {len(stock_list)} stocks")
    return True

def dump_daily_data(db):
    """Dump daily trading data from Sina."""
    logger.info("=" * 60)
    logger.info("Step 4: Fetching Daily Trading Data from Sina")
    logger.info("=" * 60)
    
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()
    table_name = f"stock_daily_{last_trading_date}"
    total_stock_count = db.get_stock_count()
    estimated_pages = int(1 + int(total_stock_count) / 100)

    db_engine = db.create_engine()
    logger.info(f"Last trading date: {last_trading_date}")
    logger.info(f"Total estimated pages: {estimated_pages}")
    
    for i in range(estimated_pages):
        logger.info(f"Retrieving page {i+1}/{estimated_pages}...")
        info = stock_info.get_daily_info_by_page(i+1)
        data = pd.read_json(info)
        data.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
        time.sleep(1)
    
    logger.info(f"✓ Daily trading data saved to {table_name}")

# ============================================================================
# Main Function
# ============================================================================

def main():
    """Main execution function."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Daily stock data update')
    parser.add_argument('--db', default='jia-stk.db', help='Database file path')
    parser.add_argument('--tdx-folder', default='./tdx', help='TDX folder path')
    parser.add_argument('--skip-history', action='store_true', help='Skip TDX history import')
    parser.add_argument('--skip-daily', action='store_true', help='Skip daily data fetch')
    parser.add_argument('--history-only', action='store_true', help='Only import TDX history')
    parser.add_argument('--daily-only', action='store_true', help='Only fetch daily data')
    
    args = parser.parse_args()
    
    logger.info("=" * 60)
    logger.info("DAILY STOCK DATA UPDATE")
    logger.info(f"Database: {args.db}")
    logger.info(f"TDX Folder: {args.tdx_folder}")
    logger.info("=" * 60)
    
    db = DbSqlite(args.db)
    
    try:
        # Import TDX historical data
        if not args.skip_history and not args.daily_only:
            if not import_tdx_history(db, args.tdx_folder):
                logger.err("TDX history import failed")
                if not args.history_only:
                    logger.warn("Continuing with daily data update...")
        
        # Update stock list from TDX
        if not args.skip_history and not args.daily_only:
            update_stock_list(db, args.tdx_folder)
        
        # Fetch daily trading data
        if not args.skip_daily and not args.history_only:
            dump_stock_list(db, args.tdx_folder)
            dump_daily_data(db)
        
        logger.info("=" * 60)
        logger.info("✓ DAILY UPDATE COMPLETED SUCCESSFULLY!")
        logger.info("=" * 60)
        
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    main()
