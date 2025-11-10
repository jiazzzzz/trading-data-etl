#!/usr/bin/env python3
"""
Daily data dump script for SQLite
This program should be executed daily to save stock data to SQLite database
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

logger = Logger("dump")

def insert_pinyin(row):
    """Add pinyin to stock name"""
    out = {}
    com = Common()
    out['pinyin'] = com.get_py_from_name(row['name'])
    return pd.Series(out)

def dump_stock_list(db, tdx_folder='./tdx'):
    """Dump stock list from TDX to database"""
    logger.info("=" * 60)
    logger.info("Dumping stock list to database")
    logger.info("=" * 60)
    
    db.drop_table("stock_list")
    db_engine = db.create_engine()
    stock_info = StockInfo()
    
    # Get stock list from TDX folder
    stock_list = stock_info.get_stock_list_from_tdx(tdx_folder)
    
    if stock_list is None or stock_list.empty:
        logger.err("Failed to retrieve stock list from TDX")
        return False
    
    # Add pinyin
    stock_list[['pinyin']] = stock_list.apply(insert_pinyin, axis=1)
    
    logger.info(f"Dumping {len(stock_list)} stocks to database")
    stock_list.to_sql("stock_list", db_engine, index=False, if_exists='append', chunksize=5000)
    
    logger.info(f"✓ Successfully dumped {len(stock_list)} stocks")
    return True

def dump_daily_data(db):
    """Dump daily trading data from Sina to database"""
    logger.info("=" * 60)
    logger.info("Dumping daily trading data from Sina")
    logger.info("=" * 60)
    
    stock_info = StockInfo()
    last_trading_date = stock_info.get_last_trading_date()
    table_name = f"stock_daily_{last_trading_date}"
    total_stock_count = db.get_stock_count()
    estimated_pages = int(1 + int(total_stock_count) / 100)

    db_engine = db.create_engine()
    
    logger.info(f"Last trading date: {last_trading_date}")
    logger.info(f"Total stocks: {total_stock_count}")
    logger.info(f"Estimated pages: {estimated_pages}")
    
    for i in range(estimated_pages):
        logger.info(f"Retrieving page {i+1}/{estimated_pages}...")
        try:
            info = stock_info.get_daily_info_by_page(i+1)
            data = pd.read_json(info)
            data.to_sql(table_name, db_engine, index=False, if_exists='append', chunksize=5000)
            time.sleep(1)
        except Exception as e:
            logger.err(f"Error retrieving page {i+1}: {str(e)}")
            continue
    
    logger.info(f"✓ Successfully dumped daily data to {table_name}")
    return True

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Dump stock data to SQLite database')
    parser.add_argument('--db', default='jia-stk.db', help='SQLite database file path')
    parser.add_argument('--tdx-folder', default='./tdx', help='TDX folder path')
    parser.add_argument('--skip-list', action='store_true', help='Skip stock list dump')
    parser.add_argument('--skip-daily', action='store_true', help='Skip daily data dump')
    
    args = parser.parse_args()
    
    logger.info("=" * 60)
    logger.info("DAILY DATA DUMP")
    logger.info(f"Database: {args.db}")
    logger.info(f"TDX Folder: {args.tdx_folder}")
    logger.info("=" * 60)
    
    # Initialize SQLite database
    db = DbSqlite(args.db)
    
    try:
        # Dump stock list
        if not args.skip_list:
            if not dump_stock_list(db, args.tdx_folder):
                logger.err("Stock list dump failed")
                if not args.skip_daily:
                    logger.warn("Continuing with daily data dump...")
        
        # Dump daily trading data
        if not args.skip_daily:
            dump_daily_data(db)
        
        logger.info("=" * 60)
        logger.info("✓ DUMP COMPLETED SUCCESSFULLY!")
        logger.info("=" * 60)
        
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
