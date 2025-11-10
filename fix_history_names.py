#!/usr/bin/env python3
"""
Fix stock names in stock_history table by updating from stock_list
"""
import sys
sys.path.append('./lib')
from lib.db_sqlite import DbSqlite
from logger import Logger

logger = Logger("fix_names")

def fix_history_names(db_path='jia-stk.db'):
    """Update stock names in stock_history from stock_list"""
    logger.info("=" * 60)
    logger.info("Fixing stock names in stock_history table")
    logger.info("=" * 60)
    
    db = DbSqlite(db_path)
    
    # Get all unique stock codes from history
    result = db.exec_and_fetch("SELECT DISTINCT stock_code FROM stock_history")
    stock_codes = [row[0] for row in result]
    logger.info(f"Found {len(stock_codes)} unique stocks in history")
    
    # Update each stock's name from stock_list
    updated_count = 0
    not_found_count = 0
    
    for stock_code in stock_codes:
        # Get correct name from stock_list
        result = db.exec_and_fetch(f"SELECT name FROM stock_list WHERE symbol = '{stock_code}'")
        
        if result and len(result) > 0:
            correct_name = result[0][0]
            
            # Update in stock_history
            escaped_name = correct_name.replace("'", "''")
            query = f"UPDATE stock_history SET stock_name = '{escaped_name}' WHERE stock_code = '{stock_code}'"
            db.exec(query)
            updated_count += 1
            
            if updated_count % 500 == 0:
                logger.info(f"Updated {updated_count} stocks...")
        else:
            not_found_count += 1
            if not_found_count <= 10:
                logger.warn(f"Stock {stock_code} not found in stock_list")
    
    logger.info("=" * 60)
    logger.info(f"✓ Updated {updated_count} stock names")
    if not_found_count > 0:
        logger.warn(f"⚠ {not_found_count} stocks not found in stock_list")
    logger.info("=" * 60)
    
    # Verify the fix
    logger.info("\nVerifying fixes:")
    for code in ['000514', '002136']:
        result = db.exec_and_fetch(f"SELECT DISTINCT stock_name FROM stock_history WHERE stock_code = '{code}'")
        if result:
            logger.info(f"  {code}: {result[0][0]}")
    
    return True

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Fix stock names in stock_history')
    parser.add_argument('--db', default='jia-stk.db', help='SQLite database file path')
    
    args = parser.parse_args()
    
    try:
        fix_history_names(args.db)
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
