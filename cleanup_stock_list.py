#!/usr/bin/env python3
"""
Clean up stock_list table - keep only specific stock types
Keep: 920 (北证), 688 (科创板), 600 (上证主板), 300 (创业板), 002 (深证中小板)
"""
import sys
sys.path.append('./lib')
from lib.db_sqlite import DbSqlite
from logger import Logger

logger = Logger("cleanup")

def cleanup_stock_list(db_path='jia-stk.db'):
    """Remove stocks that don't match the specified prefixes"""
    logger.info("=" * 60)
    logger.info("Cleaning up stock_list table")
    logger.info("=" * 60)
    
    db = DbSqlite(db_path)
    
    # Count total stocks before cleanup
    result = db.exec_and_fetch("SELECT COUNT(*) FROM stock_list")
    total_before = result[0][0] if result else 0
    logger.info(f"Total stocks before cleanup: {total_before}")
    
    # Count stocks to keep
    keep_query = """
        SELECT COUNT(*) FROM stock_list 
        WHERE ts_code LIKE '920%' 
           OR ts_code LIKE '688%' 
           OR ts_code LIKE '600%' 
           OR ts_code LIKE '300%' 
           OR ts_code LIKE '002%'
    """
    result = db.exec_and_fetch(keep_query)
    to_keep = result[0][0] if result else 0
    logger.info(f"Stocks to keep: {to_keep}")
    logger.info(f"Stocks to remove: {total_before - to_keep}")
    
    # Show some examples of stocks to be removed
    logger.info("\nExamples of stocks to be removed:")
    remove_examples = db.exec_and_fetch("""
        SELECT ts_code, symbol, name FROM stock_list 
        WHERE ts_code NOT LIKE '920%' 
          AND ts_code NOT LIKE '688%' 
          AND ts_code NOT LIKE '600%' 
          AND ts_code NOT LIKE '300%' 
          AND ts_code NOT LIKE '002%'
        LIMIT 10
    """)
    for row in remove_examples:
        logger.info(f"  {row[0]} | {row[1]} | {row[2]}")
    
    # Ask for confirmation
    print("\n" + "=" * 60)
    print("WARNING: This will permanently delete stocks from the database!")
    print("=" * 60)
    response = input("Do you want to continue? (yes/no): ").strip().lower()
    
    if response != 'yes':
        logger.info("Cleanup cancelled by user")
        return False
    
    # Delete stocks that don't match the prefixes
    delete_query = """
        DELETE FROM stock_list 
        WHERE ts_code NOT LIKE '920%' 
          AND ts_code NOT LIKE '688%' 
          AND ts_code NOT LIKE '600%' 
          AND ts_code NOT LIKE '300%' 
          AND ts_code NOT LIKE '002%'
    """
    
    logger.info("\nDeleting stocks...")
    db.exec(delete_query)
    
    # Count stocks after cleanup
    result = db.exec_and_fetch("SELECT COUNT(*) FROM stock_list")
    total_after = result[0][0] if result else 0
    
    logger.info("=" * 60)
    logger.info(f"✓ Cleanup completed!")
    logger.info(f"Stocks before: {total_before}")
    logger.info(f"Stocks after: {total_after}")
    logger.info(f"Stocks removed: {total_before - total_after}")
    logger.info("=" * 60)
    
    # Show breakdown by exchange
    logger.info("\nBreakdown by prefix:")
    for prefix, name in [('920', '北证'), ('688', '科创板'), ('600', '上证主板'), 
                          ('300', '创业板'), ('002', '深证中小板')]:
        result = db.exec_and_fetch(f"SELECT COUNT(*) FROM stock_list WHERE ts_code LIKE '{prefix}%'")
        count = result[0][0] if result else 0
        logger.info(f"  {prefix} ({name}): {count} stocks")
    
    return True

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Clean up stock_list table')
    parser.add_argument('--db', default='jia-stk.db', help='SQLite database file path')
    
    args = parser.parse_args()
    
    try:
        cleanup_stock_list(args.db)
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
