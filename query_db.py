#!/usr/bin/env python3
"""
SQLite Database Query Tool
Simple utility to explore and query jia-stk.db
"""
import sys
sys.path.append('./lib')
from lib.db_sqlite import DbSqlite
from lib.logger import Logger
import pandas as pd
import argparse

logger = Logger("query_db", log_level=1)

def show_tables(db):
    """Show all tables in database"""
    tables = db.get_db_tables()
    print("\n" + "="*60)
    print(f"Total Tables: {len(tables)}")
    print("="*60)
    for i, table in enumerate(tables, 1):
        print(f"{i}. {table}")
    print("="*60 + "\n")
    return tables

def show_table_info(db, table_name):
    """Show table structure and row count"""
    conn = db.connect()
    cursor = conn.cursor()
    
    # Get table info
    cursor.execute(f"PRAGMA table_info({table_name})")
    columns = cursor.fetchall()
    
    # Get row count
    cursor.execute(f"SELECT COUNT(*) FROM {table_name}")
    row_count = cursor.fetchone()[0]
    
    conn.close()
    
    print(f"\n{'='*60}")
    print(f"Table: {table_name}")
    print(f"Rows: {row_count}")
    print(f"{'='*60}")
    print(f"{'Column':<20} {'Type':<15} {'NotNull':<10} {'Default'}")
    print("-"*60)
    for col in columns:
        col_name = col[1]
        col_type = col[2]
        not_null = "NOT NULL" if col[3] else ""
        default = col[4] if col[4] else ""
        print(f"{col_name:<20} {col_type:<15} {not_null:<10} {default}")
    print("="*60 + "\n")

def query_table(db, table_name, limit=10):
    """Query and display table data"""
    try:
        query = f"SELECT * FROM {table_name} LIMIT {limit}"
        df = pd.read_sql_query(query, db.connect())
        
        print(f"\n{'='*60}")
        print(f"Table: {table_name} (showing first {limit} rows)")
        print(f"{'='*60}\n")
        
        # Display with pandas formatting
        pd.set_option('display.max_columns', None)
        pd.set_option('display.width', None)
        pd.set_option('display.max_colwidth', 20)
        print(df.to_string(index=False))
        print(f"\n{'='*60}\n")
        
    except Exception as e:
        logger.err(f"Error querying table: {str(e)}")

def custom_query(db, sql):
    """Execute custom SQL query"""
    try:
        df = pd.read_sql_query(sql, db.connect())
        print(f"\n{'='*60}")
        print(f"Query Results: {len(df)} rows")
        print(f"{'='*60}\n")
        
        pd.set_option('display.max_columns', None)
        pd.set_option('display.width', None)
        print(df.to_string(index=False))
        print(f"\n{'='*60}\n")
        
    except Exception as e:
        logger.err(f"Error executing query: {str(e)}")

def interactive_mode(db):
    """Interactive query mode"""
    print("\n" + "="*60)
    print("Interactive SQLite Query Mode")
    print("="*60)
    print("Commands:")
    print("  tables          - Show all tables")
    print("  info <table>    - Show table structure")
    print("  show <table>    - Show table data (first 10 rows)")
    print("  show <table> N  - Show first N rows")
    print("  sql <query>     - Execute custom SQL")
    print("  quit/exit       - Exit interactive mode")
    print("="*60 + "\n")
    
    while True:
        try:
            cmd = input("sqlite> ").strip()
            
            if not cmd:
                continue
            
            if cmd.lower() in ['quit', 'exit', 'q']:
                print("Goodbye!")
                break
            
            parts = cmd.split(maxsplit=2)
            command = parts[0].lower()
            
            if command == 'tables':
                show_tables(db)
            
            elif command == 'info' and len(parts) >= 2:
                table_name = parts[1]
                show_table_info(db, table_name)
            
            elif command == 'show' and len(parts) >= 2:
                table_name = parts[1]
                limit = int(parts[2]) if len(parts) == 3 else 10
                query_table(db, table_name, limit)
            
            elif command == 'sql' and len(parts) >= 2:
                sql = ' '.join(parts[1:])
                custom_query(db, sql)
            
            else:
                print("Unknown command. Type 'help' for available commands.")
        
        except KeyboardInterrupt:
            print("\nUse 'quit' or 'exit' to leave.")
        except Exception as e:
            print(f"Error: {str(e)}")

def main():
    parser = argparse.ArgumentParser(description='Query SQLite database')
    parser.add_argument('--db', type=str, default='jia-stk.db',
                        help='Database file path (default: jia-stk.db)')
    parser.add_argument('--tables', action='store_true',
                        help='Show all tables')
    parser.add_argument('--info', type=str, metavar='TABLE',
                        help='Show table structure')
    parser.add_argument('--show', type=str, metavar='TABLE',
                        help='Show table data')
    parser.add_argument('--limit', type=int, default=10,
                        help='Limit rows to display (default: 10)')
    parser.add_argument('--sql', type=str,
                        help='Execute custom SQL query')
    parser.add_argument('--interactive', '-i', action='store_true',
                        help='Start interactive mode')
    
    args = parser.parse_args()
    
    # Initialize database
    db = DbSqlite(args.db)
    
    # Execute commands
    if args.tables:
        show_tables(db)
    
    elif args.info:
        show_table_info(db, args.info)
    
    elif args.show:
        query_table(db, args.show, args.limit)
    
    elif args.sql:
        custom_query(db, args.sql)
    
    elif args.interactive:
        interactive_mode(db)
    
    else:
        # Default: show tables and enter interactive mode
        show_tables(db)
        interactive_mode(db)

if __name__ == '__main__':
    main()
