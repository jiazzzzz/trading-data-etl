# Database Configuration Guide

This project now supports both **SQLite** and **MySQL** databases.

## SQLite (Default - Recommended for Single User)

SQLite is a lightweight, file-based database that requires no server setup. Perfect for personal use.

### Configuration

Edit `settings.conf`:

```ini
[db]
type = sqlite
sqlite_path = jia-stk.db
```

### Advantages
- No database server required
- Zero configuration
- Portable (single file)
- Perfect for development and personal use

## MySQL (For Production/Multi-User)

MySQL is a full-featured database server suitable for production environments.

### Configuration

Edit `settings.conf`:

```ini
[db]
type = mysql
ip = 127.0.0.1
user = root
passwd = your_password
```

### Requirements

Install MySQL driver:
```bash
pip install pymysql
```

## Usage Examples

### Using SQLite (default)
```bash
python dump.py
```

### Force MySQL (override config)
```bash
python dump.py --db-type mysql
```

### Force SQLite (override config)
```bash
python dump.py --db-type sqlite
```

### Other Options
```bash
# Only update stock list
python dump.py --stock-list-only

# Only dump daily data
python dump.py --daily-only

# Dump specific date
python dump.py --date 20250108

# Force overwrite existing data
python dump.py --force
```

## Database Files

- **SQLite**: Data stored in `jia-stk.db` file
- **MySQL**: Data stored on MySQL server

## Migration

To migrate from MySQL to SQLite or vice versa, you'll need to export and import your data manually using database tools.


## Querying the Database

### Method 1: Using the Query Tool (Recommended)

I've created a handy `query_db.py` tool for you:

```bash
# Interactive mode (easiest)
python query_db.py

# Show all tables
python query_db.py --tables

# Show table structure
python query_db.py --info stock_list

# View table data (first 10 rows)
python query_db.py --show stock_list

# View more rows
python query_db.py --show stock_list --limit 50

# Custom SQL query
python query_db.py --sql "SELECT * FROM stock_list WHERE area='上海'"
```

### Method 2: Using SQLite Command Line

If you have SQLite installed:

```bash
# Open database
sqlite3 jia-stk.db

# Inside SQLite shell:
.tables                          # List all tables
.schema stock_list              # Show table structure
SELECT * FROM stock_list LIMIT 10;  # Query data
.quit                           # Exit
```

### Method 3: Using Python Directly

```python
import sqlite3
import pandas as pd

# Connect to database
conn = sqlite3.connect('jia-stk.db')

# Query with pandas
df = pd.read_sql_query("SELECT * FROM stock_list LIMIT 10", conn)
print(df)

# Or use cursor
cursor = conn.cursor()
cursor.execute("SELECT COUNT(*) FROM stock_list")
print(cursor.fetchone())

conn.close()
```

### Method 4: Using GUI Tools

Download and use any SQLite GUI tool:
- **DB Browser for SQLite** (Free, cross-platform) - https://sqlitebrowser.org/
- **DBeaver** (Free, supports many databases)
- **SQLiteStudio** (Free, lightweight)

Just open `jia-stk.db` with any of these tools.

## Common Queries

```bash
# Count total stocks
python query_db.py --sql "SELECT COUNT(*) as total FROM stock_list"

# Find stocks by name
python query_db.py --sql "SELECT * FROM stock_list WHERE name LIKE '%科技%'"

# Show daily tables
python query_db.py --sql "SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily%'"

# Get today's top gainers (replace date)
python query_db.py --sql "SELECT name, changepercent FROM stock_daily_20250108 ORDER BY changepercent DESC LIMIT 10"

# Search by pinyin
python query_db.py --sql "SELECT symbol, name FROM stock_list WHERE pinyin LIKE '%zg%'"
```
