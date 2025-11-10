#!/usr/bin/env python3
"""
Update stock names in database from TDX folder.
This script reads the correct names from TDX files and updates the database.
"""

import sys
sys.path.append('./lib')

from stock_info import StockInfo
from db_sqlite import DbSqlite
from logger import Logger

logger = Logger("update_names")

def update_stock_names(db_path='jia-stk.db', tdx_folder='./tdx'):
    """
    Update stock names in database from TDX folder.
    """
    logger.info("=" * 60)
    logger.info("Updating stock names from TDX folder")
    logger.info("=" * 60)
    
    # Get stock list from TDX
    stock_info = StockInfo()
    logger.info(f"Reading stock list from TDX folder: {tdx_folder}")
    stock_list = stock_info.get_stock_list_from_tdx(tdx_folder)
    
    if stock_list is None or stock_list.empty:
        logger.err("Failed to retrieve stock list from TDX folder")
        return -1
    
    logger.info(f"Found {len(stock_list)} stocks in TDX folder")
    
    # Connect to database
    db = DbSqlite(db_path)
    
    # Update each stock name
    updated_count = 0
    error_count = 0
    
    for index, row in stock_list.iterrows():
        try:
            ts_code = row['ts_code']
            name = row['name']
            symbol = row['symbol']
            
            # Escape single quotes in name
            name_escaped = name.replace("'", "''")
            
            # Update the name in database
            query = f"UPDATE stock_list SET name = '{name_escaped}' WHERE ts_code = '{ts_code}'"
            db.exec(query)
            updated_count += 1
            
            if updated_count % 500 == 0:
                logger.info(f"Updated {updated_count} stocks...")
                
        except Exception as e:
            logger.err(f"Error updating {ts_code}: {str(e)}")
            error_count += 1
    
    logger.info("=" * 60)
    logger.info(f"✓ Successfully updated {updated_count} stock names")
    if error_count > 0:
        logger.warn(f"⚠ {error_count} errors occurred")
    logger.info("=" * 60)
    
    return updated_count

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Update stock names from TDX folder')
    parser.add_argument('--db', default='jia-stk.db', help='Database file path')
    parser.add_argument('--tdx-folder', default='./tdx', help='TDX folder path')
    
    args = parser.parse_args()
    
    result = update_stock_names(args.db, args.tdx_folder)
    
    if result > 0:
        print(f"\n✓ Successfully updated {result} stock names!")
    else:
        print("\n✗ Failed to update stock names")
        sys.exit(1)
