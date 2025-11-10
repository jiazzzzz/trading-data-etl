# System Architecture

## Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Stock Data System                         │
└─────────────────────────────────────────────────────────────┘

┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│   Tushare    │      │ Sina Finance │      │   Xueqiu     │
│     API      │      │     API      │      │     API      │
└──────┬───────┘      └──────┬───────┘      └──────┬───────┘
       │                     │                      │
       └─────────────────────┼──────────────────────┘
                             │
                    ┌────────▼────────┐
                    │   dump.py       │
                    │  (Data Fetcher) │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  jia-stk.db     │
                    │   (SQLite)      │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────────┐    │    ┌────────▼────────┐
     │  query_db.py    │    │    │   Go Server     │
     │ (Python Tool)   │    │    │   (REST API)    │
     └────────┬────────┘    │    └────────┬────────┘
              │             │              │
     ┌────────▼────────┐    │    ┌────────▼────────┐
     │   Terminal      │    │    │   index.html    │
     │   Interface     │    │    │   (Web UI)      │
     └─────────────────┘    │    └─────────────────┘
                            │
                   ┌────────▼────────┐
                   │  External Apps  │
                   │  (curl, Python, │
                   │   JavaScript)   │
                   └─────────────────┘
```

## Data Flow

### 1. Data Collection (dump.py)

```
┌─────────────┐
│  dump.py    │
└──────┬──────┘
       │
       ├─► Fetch stock list from Tushare
       │   └─► Add pinyin for search
       │   └─► Save to stock_list table
       │
       └─► Fetch daily data from Sina
           └─► Paginated requests (100 stocks/page)
           └─► Save to stock_daily_YYYYMMDD table
```

### 2. Data Query (Python)

```
┌──────────────┐
│ query_db.py  │
└──────┬───────┘
       │
       ├─► Interactive Mode
       │   └─► User commands (tables, show, sql)
       │   └─► Display formatted results
       │
       └─► Command Line Mode
           └─► Direct queries
           └─► JSON/Table output
```

### 3. Web API (Go Server)

```
┌──────────────┐
│  Go Server   │
└──────┬───────┘
       │
       ├─► HTTP Endpoints
       │   ├─► /api/stocks
       │   ├─► /api/search
       │   ├─► /api/daily/:date
       │   └─► /api/query
       │
       └─► SQLite Connection
           └─► Parameterized queries
           └─► JSON responses
```

## Component Details

### Python Components

```
lib/
├── stock_info.py      # Data fetching from APIs
│   ├── get_stock_list()
│   ├── get_daily_info_by_page()
│   └── get_last_trading_date()
│
├── db_sqlite.py       # SQLite operations
│   ├── create_engine()
│   ├── exec_and_fetch()
│   ├── get_stock_count()
│   └── drop_table()
│
├── logger.py          # Logging utility
│   ├── info()
│   ├── warn()
│   └── err()
│
└── common.py          # Utilities
    ├── read_conf()
    └── get_py_from_name()
```

### Go Components

```
server/main.go
├── Database Layer
│   ├── initDB()
│   └── SQL queries
│
├── HTTP Handlers
│   ├── getStocksHandler()
│   ├── searchStocksHandler()
│   ├── getDailyDataHandler()
│   ├── getTopGainersHandler()
│   └── customQueryHandler()
│
└── Middleware
    └── CORS handler
```

## Database Schema

### stock_list Table
```sql
CREATE TABLE stock_list (
    ts_code TEXT,      -- Tushare code (e.g., "000001.SZ")
    symbol TEXT,       -- Stock symbol (e.g., "000001")
    name TEXT,         -- Stock name (e.g., "平安银行")
    area TEXT,         -- Location (e.g., "深圳")
    industry TEXT,     -- Industry (e.g., "银行")
    list_date TEXT,    -- Listing date (e.g., "19910403")
    pinyin TEXT        -- Pinyin (e.g., "payx")
);
```

### stock_daily_YYYYMMDD Table
```sql
CREATE TABLE stock_daily_20251107 (
    symbol TEXT,           -- Stock symbol
    code INTEGER,          -- Stock code
    name TEXT,             -- Stock name
    trade REAL,            -- Current price
    pricechange REAL,      -- Price change
    changepercent REAL,    -- Change percentage
    buy REAL,              -- Buy price
    sell REAL,             -- Sell price
    settlement REAL,       -- Settlement price
    open REAL,             -- Open price
    high REAL,             -- High price
    low REAL,              -- Low price
    volume INTEGER,        -- Trading volume
    amount INTEGER,        -- Trading amount
    ticktime TEXT,         -- Tick time
    per REAL,              -- P/E ratio
    pb REAL,               -- P/B ratio
    mktcap REAL,           -- Market cap
    nmc REAL,              -- Net market cap
    turnoverratio REAL,    -- Turnover ratio
    dump_time DATETIME     -- Data dump timestamp
);
```

## API Request Flow

```
Client Request
     │
     ▼
┌─────────────┐
│ Gin Router  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Handler    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   SQLite    │
│   Query     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ JSON Format │
└──────┬──────┘
       │
       ▼
Client Response
```

## Security Layers

```
┌──────────────────────────────────┐
│  Input Validation                │
│  - Query parameter sanitization  │
│  - SQL type checking             │
└────────────┬─────────────────────┘
             │
┌────────────▼─────────────────────┐
│  Query Restrictions              │
│  - Only SELECT allowed           │
│  - Result limits enforced        │
└────────────┬─────────────────────┘
             │
┌────────────▼─────────────────────┐
│  Parameterized Queries           │
│  - SQL injection prevention      │
│  - Type-safe operations          │
└────────────┬─────────────────────┘
             │
┌────────────▼─────────────────────┐
│  CORS Policy                     │
│  - Cross-origin access control   │
└──────────────────────────────────┘
```

## Deployment Options

### Development
```
Local Machine
├── Python scripts (dump, query)
├── SQLite database (file)
└── Go server (localhost:8080)
```

### Production
```
Server
├── Systemd service (Go server)
├── Nginx reverse proxy (HTTPS)
├── Cron job (daily data dump)
└── SQLite database (persistent storage)
```

## Performance Considerations

### Data Collection
- **Rate Limiting**: 1 second delay between pages
- **Batch Size**: 100 stocks per request
- **Error Recovery**: Continue on page failure

### Database
- **Indexes**: Consider adding for frequently queried columns
- **Partitioning**: Daily tables by date
- **Cleanup**: Old tables can be archived/deleted

### API Server
- **Connection Pool**: SQLite connection management
- **Result Limits**: Max 1000 rows per query
- **Caching**: Consider Redis for hot data

## Scalability Path

```
Current: SQLite (Single file)
    │
    ├─► Add indexes for performance
    │
    ├─► Add Redis cache layer
    │
    ├─► Migrate to PostgreSQL (if needed)
    │
    └─► Add load balancer (multiple servers)
```

## Monitoring Points

1. **Data Collection**
   - Success rate of API calls
   - Number of stocks fetched
   - Execution time

2. **Database**
   - Database size
   - Query performance
   - Table count

3. **API Server**
   - Request rate
   - Response time
   - Error rate

## Integration Points

### Python Integration
```python
import sqlite3
conn = sqlite3.connect('jia-stk.db')
# Query data
```

### JavaScript Integration
```javascript
fetch('http://localhost:8080/api/stocks')
  .then(res => res.json())
```

### Command Line Integration
```bash
curl http://localhost:8080/api/search?q=科技
```

## Future Architecture

```
┌─────────────────────────────────────────┐
│         Microservices (Optional)        │
├─────────────────────────────────────────┤
│  Data Collector │ API Server │ Analytics│
│     Service     │   Service  │  Service │
└────────┬────────┴─────┬──────┴────┬─────┘
         │              │           │
    ┌────▼──────────────▼───────────▼────┐
    │      Message Queue (RabbitMQ)      │
    └────────────────┬───────────────────┘
                     │
            ┌────────▼────────┐
            │   PostgreSQL    │
            │   (Primary DB)  │
            └────────┬────────┘
                     │
            ┌────────▼────────┐
            │  Redis Cache    │
            └─────────────────┘
```
