#!/usr/bin/env python3
"""
TDX Historical Data Importer
Import historical stock data from TDX export files to SQLite database.
"""
import sys
sys.path.append('./lib')
from lib.db_sqlite import DbSqlite
from lib.logger import Logger
import os
import re
from datetime import datetime
import sqlite3

logger = Logger("import_tdx", log_level=1)

def parse_tdx_file(file_path):
    """
    Parse a single TDX export file.
    
    Args:
        file_path: Path to TDX file
        
    Returns:
        tuple: (stock_code, stock_name, exchange, data_rows)
    """
    try:
        with open(file_path, 'r', encoding='gbk') as f:
            lines = f.readlines()
        
        if len(lines) < 3:
            logger.warn(f"File {file_path} has insufficient data")
            return None
        
        # Parse header line (first line contains stock code and name)
        header = lines[0].strip()
        parts = header.split()
        if len(parts) < 2:
            logger.warn(f"Invalid header in {file_path}")
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
        
        # Parse data rows (skip first 2 lines: header and column names)
        data_rows = []
        for line in lines[2:]:
            line = line.strip()
            if not line:
                continue
            
            parts = line.split('\t')
            if len(parts) < 7:
                continue
            
            try:
                # Parse date (format: YYYY/MM/DD)
                date_str = parts[0].strip()
                date_obj = datetime.strptime(date_str, '%Y/%m/%d')
                trade_date = date_obj.strftime('%Y%m%d')
                
                # Parse OHLCVA data
                open_price = float(parts[1])
                high_price = float(parts[2])
                low_price = float(parts[3])
                close_price = float(parts[4])
                volume = int(float(parts[5]))
                amount = float(parts[6])
                
                data_rows.append({
                    'trade_date': trade_date,
                    'open': open_price,
                    'high': high_price,
                    'low': low_price,
                    'close': close_price,
                    'volume': volume,
                    'amount': amount
                })
            except (ValueError, IndexError) as e:
                logger.warn(f"Error parsing line in {file_path}: {line[:50]}... - {str(e)}")
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
        
        # Create indexes for better query performance
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_stock_code 
            ON stock_history(stock_code)
        """)
        
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_trade_date 
            ON stock_history(trade_date)
        """)
        
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_stock_date 
            ON stock_history(stock_code, trade_date)
        """)
        
        conn.commit()
        logger.info("stock_history table created successfully")
    except Exception as e:
        logger.err(f"Error creating table: {str(e)}")
        raise
    finally:
        conn.close()

def import_tdx_data(db, tdx_folder='tdx', clear_existing=False):
    """
    Import all TDX files from folder to database.
    
    Args:
        db: Database connection object
        tdx_folder: Path to TDX data folder
        clear_existing: If True, clear existing data before import
    """
    if not os.path.exists(tdx_folder):
        logger.err(f"TDX folder not found: {tdx_folder}")
        return False
    
    # Create table
    create_history_table(db)
    
    # Clear existing data if requested
    if clear_existing:
        logger.info("Clearing existing historical data...")
        conn = db.connect()
        cursor = conn.cursor()
        cursor.execute("DELETE FROM stock_history")
        conn.commit()
        conn.close()
        logger.info("Existing data cleared")
    
    # Get all TDX files
    tdx_files = [f for f in os.listdir(tdx_folder) if f.endswith('.txt')]
    total_files = len(tdx_files)
    logger.info(f"Found {total_files} TDX files to import")
    
    # Import each file
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
        
        stock_code, stock_name_tdx, exchange, data_rows = result
        
        if not data_rows:
            logger.warn(f"No data rows in {filename}")
            error_count += 1
            continue
        
        # Get correct stock name from stock_list table
        cursor.execute("SELECT name FROM stock_list WHERE symbol = ?", (stock_code,))
        name_result = cursor.fetchone()
        stock_name = name_result[0] if name_result else stock_name_tdx
        
        # Insert data rows
        try:
            for row in data_rows:
                cursor.execute("""
                    INSERT OR REPLACE INTO stock_history 
                    (stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    stock_code,
                    stock_name,
                    exchange,
                    row['trade_date'],
                    row['open'],
                    row['high'],
                    row['low'],
                    row['close'],
                    row['volume'],
                    row['amount']
                ))
            
            conn.commit()
            success_count += 1
            total_records += len(data_rows)
            
        except Exception as e:
            logger.err(f"Error importing {filename}: {str(e)}")
            error_count += 1
            conn.rollback()
    
    conn.close()
    
    # Summary
    logger.info("=" * 60)
    logger.info(f"Import completed!")
    logger.info(f"Total files: {total_files}")
    logger.info(f"Success: {success_count}")
    logger.info(f"Errors: {error_count}")
    logger.info(f"Total records imported: {total_records}")
    logger.info("=" * 60)
    
    return True

def main():
    """Main execution function."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Import TDX historical data to SQLite')
    parser.add_argument('--db', type=str, default='jia-stk.db',
                        help='SQLite database path (default: jia-stk.db)')
    parser.add_argument('--folder', type=str, default='tdx',
                        help='TDX data folder path (default: tdx)')
    parser.add_argument('--clear', action='store_true',
                        help='Clear existing historical data before import')
    
    args = parser.parse_args()
    
    try:
        logger.info("Starting TDX historical data import...")
        logger.info(f"Database: {args.db}")
        logger.info(f"TDX folder: {args.folder}")
        
        # Initialize database
        db = DbSqlite(args.db)
        
        # Import data
        success = import_tdx_data(db, args.folder, args.clear)
        
        if success:
            logger.info("Import completed successfully!")
            sys.exit(0)
        else:
            logger.err("Import failed!")
            sys.exit(1)
    
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    main()
