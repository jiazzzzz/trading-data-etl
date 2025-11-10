#!/usr/bin/env python3
"""
Clean up stock_history table - keep only main boards
Keep: 沪市主板(600/601/603/605), 科创板(688), 深主板(000/001), 创业板(300/301), 北证(920/830/430)
"""
import sys
sys.path.append('./lib')
from lib.db_sqlite import DbSqlite
from logger import Logger

logger = Logger("cleanup_history")

def cleanup_stock_history(db_path='jia-stk.db'):
    """Remove stocks that don't match the main board prefixes"""
    logger.info("=" * 60)
    logger.info("Cleaning up stock_history table")
    logger.info("=" * 60)
    
    db = DbSqlite(db_path)
    
    # Count total records before cleanup
    result = db.exec_and_fetch("SELECT COUNT(*) FROM stock_history")
    total_before = result[0][0] if result else 0
    logger.info(f"Total records before cleanup: {total_before}")
    
    # Count records to keep
    keep_query = """
        SELECT COUNT(*) FROM stock_history 
        WHERE stock_code LIKE '600%' 
           OR stock_code LIKE '601%' 
           OR stock_code LIKE '603%' 
           OR stock_code LIKE '605%'
           OR stock_code LIKE '688%'
           OR stock_code LIKE '000%'
           OR stock_code LIKE '001%'
           OR stock_code LIKE '002%'
           OR stock_code LIKE '003%'
           OR stock_code LIKE '300%'
           OR stock_code LIKE '301%'
           OR stock_code LIKE '920%'
           OR stock_code LIKE '830%'
           OR stock_code LIKE '430%'
    """
    result = db.exec_and_fetch(keep_query)
    to_keep = result[0][0] if result else 0
    logger.info(f"Records to keep: {to_keep}")
    logger.info(f"Records to remove: {total_before - to_keep}")
    
    # Show some examples of records to be removed
    logger.info("\nExamples of records to be removed:")
    remove_examples = db.exec_and_fetch("""
        SELECT DISTINCT stock_code, stock_name FROM stock_history 
        WHERE stock_code NOT LIKE '600%' 
          AND stock_code NOT LIKE '601%' 
          AND stock_code NOT LIKE '603%' 
          AND stock_code NOT LIKE '605%'
          AND stock_code NOT LIKE '688%'
          AND stock_code NOT LIKE '000%'
          AND stock_code NOT LIKE '001%'
          AND stock_code NOT LIKE '002%'
          AND stock_code NOT LIKE '003%'
          AND stock_code NOT LIKE '300%'
          AND stock_code NOT LIKE '301%'
          AND stock_code NOT LIKE '920%'
          AND stock_code NOT LIKE '830%'
          AND stock_code NOT LIKE '430%'
        LIMIT 10
    """)
    for row in remove_examples:
        logger.info(f"  {row[0]} | {row[1]}")
    
    # Ask for confirmation
    print("\n" + "=" * 60)
    print("WARNING: This will permanently delete records from the database!")
    print("=" * 60)
    response = input("Do you want to continue? (yes/no): ").strip().lower()
    
    if response != 'yes':
        logger.info("Cleanup cancelled by user")
        return False
    
    # Delete records that don't match the prefixes
    delete_query = """
        DELETE FROM stock_history 
        WHERE stock_code NOT LIKE '600%' 
          AND stock_code NOT LIKE '601%' 
          AND stock_code NOT LIKE '603%' 
          AND stock_code NOT LIKE '605%'
          AND stock_code NOT LIKE '688%'
          AND stock_code NOT LIKE '000%'
          AND stock_code NOT LIKE '001%'
          AND stock_code NOT LIKE '002%'
          AND stock_code NOT LIKE '003%'
          AND stock_code NOT LIKE '300%'
          AND stock_code NOT LIKE '301%'
          AND stock_code NOT LIKE '920%'
          AND stock_code NOT LIKE '830%'
          AND stock_code NOT LIKE '430%'
    """
    
    logger.info("\nDeleting records...")
    db.exec(delete_query)
    
    # Count records after cleanup
    result = db.exec_and_fetch("SELECT COUNT(*) FROM stock_history")
    total_after = result[0][0] if result else 0
    
    logger.info("=" * 60)
    logger.info(f"✓ Cleanup completed!")
    logger.info(f"Records before: {total_before}")
    logger.info(f"Records after: {total_after}")
    logger.info(f"Records removed: {total_before - total_after}")
    logger.info("=" * 60)
    
    # Show breakdown by board
    logger.info("\nBreakdown by board:")
    boards = [
        ('600/601/603/605', '沪市主板', "stock_code LIKE '600%' OR stock_code LIKE '601%' OR stock_code LIKE '603%' OR stock_code LIKE '605%'"),
        ('688', '科创板', "stock_code LIKE '688%'"),
        ('000/001', '深主板', "stock_code LIKE '000%' OR stock_code LIKE '001%'"),
        ('002/003', '深证中小板', "stock_code LIKE '002%' OR stock_code LIKE '003%'"),
        ('300/301', '创业板', "stock_code LIKE '300%' OR stock_code LIKE '301%'"),
        ('920/830/430', '北证', "stock_code LIKE '920%' OR stock_code LIKE '830%' OR stock_code LIKE '430%'")
    ]
    
    for prefix, name, condition in boards:
        result = db.exec_and_fetch(f"SELECT COUNT(*) FROM stock_history WHERE {condition}")
        count = result[0][0] if result else 0
        logger.info(f"  {prefix} ({name}): {count} records")
    
    return True

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Clean up stock_history table')
    parser.add_argument('--db', default='jia-stk.db', help='SQLite database file path')
    
    args = parser.parse_args()
    
    try:
        cleanup_stock_history(args.db)
    except Exception as e:
        logger.critical(f"Fatal error: {str(e)}")
        import traceback
        traceback.print_exc()
        sys.exit(1)
