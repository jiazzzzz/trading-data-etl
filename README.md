# 📊 Stock Data Management System

A complete Python + Go solution for collecting, storing, and querying Chinese stock market data from TDX and Sina Finance.

## 🎯 Features

- ✅ **TDX Historical Data Import** - Import historical stock data from TDX export files
- ✅ **Sina Daily Data** - Fetch latest daily trading data from Sina Finance
- ✅ **SQLite Storage** - Lightweight, file-based database
- ✅ **Automated Updates** - Single script for daily routine updates
- ✅ **REST API** - Go web server with Gin framework
- ✅ **Web UI** - Interactive browser interface with charts
- ✅ **Search** - By name, symbol, or pinyin
- ✅ **Analytics** - Top gainers, historical data, custom queries

## ⚡ Quick Reference

**Daily routine:** `python daily_update.py`

**Query data:** `python query_db.py`

**Start web server:** `cd server && go run main.go`

**Web UI:** Open `frontend/index.html` in browser

## 🚀 Quick Start

### 1. Install Python Dependencies
```bash
pip install -r requirements
```

### 2. Prepare TDX Data
Export your stock data from TDX software to the `./tdx` folder as `.txt` files.

### 3. Run Daily Update
```bash
# Run this daily - imports TDX history + fetches Sina daily data
python daily_update.py
```

This single command:
- ✅ Imports TDX historical data from `./tdx` folder
- ✅ Updates stock list and names
- ✅ Fetches latest daily trading data from Sina
- ✅ Creates/updates SQLite database automatically

### 3. Query Data (Python)
```bash
# Interactive mode
python query_db.py

# Quick queries
python query_db.py --tables
python query_db.py --show stock_list
python query_db.py --sql "SELECT * FROM stock_list WHERE area='上海' LIMIT 10"
```

### 4. Start Servers

**Option A: One-Click Start (Recommended)**
```bash
# Double-click this file or run:
start_servers.bat
```
This will:
- Start Go backend server (Port 8080)
- Start frontend server (Port 3000)
- Open browser automatically

**Option B: Manual Start**
```bash
# Terminal 1: Start backend
cd server
go run main.go

# Terminal 2: Start frontend
cd frontend
python -m http.server 3000
```

**Stop Servers:**
```bash
# Double-click or run:
stop_servers.bat
```

### 5. Access Web UI
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api

## 📁 Project Structure

```
jia-stk/
├── daily_update.py        # ⭐ Main daily routine script (TDX + Sina)
├── query_db.py            # Python query tool
├── jia-stk.db             # SQLite database (auto-created)
├── settings.conf          # Database configuration
├── requirements           # Python dependencies
│
├── tdx/                   # TDX export files (.txt)
│   ├── SH#600000.txt     # Shanghai stocks
│   ├── SZ#000001.txt     # Shenzhen stocks
│   └── ...
│
├── lib/                   # Python libraries
│   ├── db_sqlite.py      # SQLite database handler
│   ├── stock_info.py     # Stock data fetching (Sina API)
│   ├── logger.py         # Logging utility
│   └── common.py         # Common utilities (pinyin, etc.)
│
├── server/               # Go REST API server
│   ├── main.go          # Server code
│   ├── go.mod           # Go dependencies
│   └── index.html       # Simple web UI
│
├── frontend/            # Modern Vue 3 dashboard
│   ├── index.html      # Main application
│   ├── app.js          # Vue 3 logic
│   └── history.html    # Historical data viewer
│
├── cleanup_*.py         # Database maintenance scripts
└── ARCHITECTURE.md      # System architecture docs
```

## 🔄 Data Workflow

```
┌─────────────────┐
│  TDX Software   │  Export historical data
│  (Export .txt)  │
└────────┬────────┘
         │
         ▼
    ┌────────┐
    │  ./tdx │  TDX export files
    └────┬───┘
         │
         ▼
┌────────────────────┐
│  daily_update.py   │  Daily routine script
└────────┬───────────┘
         │
         ├──► Import TDX history → stock_history table
         │
         ├──► Update stock list → stock_list table
         │
         └──► Fetch Sina data → stock_daily_YYYYMMDD table
                │
                ▼
         ┌──────────────┐
         │  jia-stk.db  │  SQLite database
         └──────┬───────┘
                │
      ┌─────────┼─────────┐
      │                   │
      ▼                   ▼
┌─────────────┐    ┌─────────────┐
│ query_db.py │    │ Go Server   │
│  (Python)   │    │  (REST API) │
└─────────────┘    └──────┬──────┘
                          │
                          ▼
                   ┌──────────────┐
                   │  Web UI      │
                   │  (Frontend)  │
                   └──────────────┘
```

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System architecture details
- **[README_DATABASE.md](README_DATABASE.md)** - Database configuration
- **[README_HISTORY.md](README_HISTORY.md)** - Historical data guide
- **[README_STRATEGY_SCANNER.md](README_STRATEGY_SCANNER.md)** - Strategy scanner

## 📥 TDX Data Preparation

### Exporting from TDX Software

1. Open TDX (通达信) software
2. Select stocks you want to export
3. Right-click → Export Data → Export to Text
4. Save files to `./tdx` folder in your project
5. Files should be named: `SH#600000.txt`, `SZ#000001.txt`, etc.

### TDX File Format

Each file contains:
```
600000 浦发银行
日期       开盘    最高    最低    收盘    成交量    成交额
2024/01/02  8.50   8.65   8.45   8.60   1234567   10500000.00
2024/01/03  8.62   8.75   8.58   8.70   1345678   11200000.00
...
```

The script automatically:
- Parses stock code and name from header
- Determines exchange (SH/SZ/BJ) from filename
- Converts dates to YYYYMMDD format
- Imports all historical records

## 🔧 Configuration

Edit `settings.conf`:

```ini
[db]
# Database type: mysql or sqlite
type = sqlite

# SQLite settings (used when type=sqlite)
sqlite_path = jia-stk.db

# MySQL settings (used when type=mysql)
ip = 127.0.0.1
user = root
passwd = your_password
```

## 💻 Daily Routine

### daily_sina_update.py - Recommended Daily Script ⭐

Run this **every day** to keep your data up-to-date (No TDX export required!):

```bash
# Daily update using Sina API only (recommended)
python daily_sina_update.py
```

**What it does:**
1. Updates stock list (if TDX folder exists, optional)
2. Fetches latest daily trading data from Sina Finance
3. **Appends Sina data to stock_history table** (for historical charts)

**Options:**
```bash
python daily_sina_update.py --db jia-stk.db           # Specify database
python daily_sina_update.py --skip-stock-list         # Skip stock list update
python daily_sina_update.py --skip-history-append     # Skip appending to history
```

**Benefits:**
- ✅ No TDX export needed daily
- ✅ Fully automated via API
- ✅ Builds historical data automatically
- ✅ Faster than TDX export

### daily_update.py - Full Update with TDX

Use this for **initial setup** or when you need to import TDX historical data:

```bash
# Full update with TDX import
python daily_update.py
```

**What it does:**
1. Imports TDX historical data from `./tdx` folder
2. Updates stock list and names from TDX
3. Fetches latest daily trading data from Sina Finance

**Options:**
```bash
python daily_update.py --db jia-stk.db           # Specify database
python daily_update.py --tdx-folder ./tdx        # Specify TDX folder
python daily_update.py --skip-history            # Skip TDX import
python daily_update.py --skip-daily              # Skip Sina daily data
python daily_update.py --history-only            # Only import TDX
python daily_update.py --daily-only              # Only fetch Sina data
```

### query_db.py - Data Query
```bash
python query_db.py                # Interactive mode
python query_db.py --tables       # List tables
python query_db.py --show stock_list --limit 10
python query_db.py --sql "SELECT * FROM stock_list WHERE area='上海'"
```

## 🌐 Web API

### Start Server
```bash
cd server
go run main.go
```

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/tables` | List all tables |
| `GET /api/stocks` | Get stock list |
| `GET /api/stocks/:symbol` | Get stock by symbol |
| `GET /api/daily/:date` | Get daily data |
| `GET /api/daily/:date/top` | Top gainers |
| `GET /api/search?q=keyword` | Search stocks |
| `GET /api/query?sql=SELECT...` | Custom SQL |

### Examples
```bash
# Search stocks
curl http://localhost:8080/api/search?q=科技

# Top 10 gainers
curl http://localhost:8080/api/daily/20251107/top?limit=10

# Custom query
curl "http://localhost:8080/api/query?sql=SELECT%20COUNT(*)%20FROM%20stock_list"
```

## 🎨 Modern Web Frontend

**New!** Professional Vue 3 + ECharts dashboard in `frontend/`:
- 📊 **Interactive Charts** - Beautiful ECharts visualizations
- 🔍 **Smart Search** - Search by name, symbol, or pinyin
- 📈 **Top Gainers** - Real-time ranking with bar charts
- 📉 **Industry Analysis** - Pie chart distribution
- 📱 **Responsive Design** - Works on all devices
- ⚡ **Fast & Modern** - Vue 3 powered

**Quick Start:**
```bash
cd frontend
python -m http.server 3000
# Open: http://localhost:3000
```

**Simple Web UI** in `server/index.html`:
- 🔍 Stock search
- 📈 Top gainers viewer
- 📋 Stock list browser
- ⚡ Custom SQL executor

## 📊 Database Schema

### stock_list
Stock list with pinyin support for search:
- `ts_code` - Tushare code (e.g., "600000.SH")
- `symbol` - Stock symbol (e.g., "600000")
- `name` - Stock name (e.g., "浦发银行")
- `area` - Location
- `industry` - Industry sector
- `list_date` - Listing date
- `pinyin` - Pinyin for search (e.g., "pfyh")

### stock_history
Historical OHLCV data imported from TDX:
- `stock_code` - Stock code
- `stock_name` - Stock name
- `exchange` - Exchange (SH/SZ/BJ)
- `trade_date` - Trading date (YYYYMMDD)
- `open`, `high`, `low`, `close` - OHLC prices
- `volume` - Trading volume
- `amount` - Trading amount
- `import_time` - Import timestamp

### stock_daily_YYYYMMDD
Daily trading data from Sina (one table per date):
- `symbol` - Stock symbol
- `name` - Stock name
- `trade` - Current price
- `changepercent` - Change percentage
- `volume` - Trading volume
- `amount` - Trading amount
- `open`, `high`, `low` - Daily OHLC
- `per`, `pb` - P/E and P/B ratios
- `mktcap` - Market capitalization
- And more...

## 🔍 Common Queries

### Python
```bash
# Count stocks
python query_db.py --sql "SELECT COUNT(*) FROM stock_list"

# Find tech stocks
python query_db.py --sql "SELECT * FROM stock_list WHERE industry LIKE '%科技%'"

# Top gainers today
python query_db.py --sql "SELECT name, changepercent FROM stock_daily_20251110 ORDER BY changepercent DESC LIMIT 10"

# Historical data for a stock
python query_db.py --sql "SELECT trade_date, close FROM stock_history WHERE stock_code='600000' ORDER BY trade_date DESC LIMIT 30"

# Check latest import
python query_db.py --sql "SELECT stock_code, MAX(trade_date) as latest FROM stock_history GROUP BY stock_code LIMIT 10"
```

### API
```bash
# Search
curl http://localhost:8080/api/search?q=银行

# Top gainers
curl http://localhost:8080/api/daily/20251107/top?limit=10

# Stocks in Shanghai
curl "http://localhost:8080/api/query?sql=SELECT%20*%20FROM%20stock_list%20WHERE%20area='上海'%20LIMIT%2010"
```

## 🤖 Automation

### Windows Task Scheduler (Recommended)
Create a daily task to run after market close:

1. Open Task Scheduler
2. Create Basic Task
3. Set trigger: Daily at 4:00 PM (after market close)
4. Action: Start a program
   - Program: `python`
   - Arguments: `C:\work\jia-stk\daily_sina_update.py`
   - Start in: `C:\work\jia-stk`

Or simply double-click `daily_sina_update.bat`

### Linux Cron
```bash
# Run at 4 PM on weekdays (after market close) - Sina only
0 16 * * 1-5 cd /path/to/jia-stk && python daily_sina_update.py

# Or run at 9 PM to ensure all data is available
0 21 * * 1-5 cd /path/to/jia-stk && python daily_sina_update.py
```

### Manual Run
```bash
# Recommended: Sina API only (no TDX needed)
python daily_sina_update.py

# Or with TDX import (for initial setup)
python daily_update.py
```

## 🛠️ Development

### Python
```bash
# Install dependencies
pip install -r requirements

# Check help
python daily_update.py --help
python query_db.py --help

# Test TDX import only
python daily_update.py --history-only

# Test Sina fetch only
python daily_update.py --daily-only
```

### Go
```bash
cd server
go mod download
go run main.go

# Build binary
go build -o stock-server main.go
```

## 🔒 Security

- ✅ SQL injection protection (parameterized queries)
- ✅ Only SELECT queries allowed in custom endpoint
- ✅ Result limits to prevent excessive data transfer
- ✅ CORS enabled for web frontends

## 📦 Dependencies

### Python
- requests - HTTP client
- tushare - Stock data API
- pandas - Data manipulation
- sqlalchemy - Database ORM
- beautifulsoup4 - HTML parsing
- websocket-client - WebSocket support

### Go
- gin-gonic/gin - Web framework
- modernc.org/sqlite - Pure Go SQLite driver (no CGO required)

## 🐛 Troubleshooting

### Python Issues

**Module not found:**
```bash
pip install -r requirements
```

**TDX folder not found:**
- Make sure `./tdx` folder exists
- Export data from TDX software as `.txt` files
- Files should be named like `SH#600000.txt`, `SZ#000001.txt`

**Database not found:**
```bash
python daily_update.py  # Creates database automatically
```

**No data in stock_history:**
- Check if TDX files exist in `./tdx` folder
- Run with `--history-only` to test TDX import
- Check logs for parsing errors

**Sina API errors:**
- Verify network connection
- Check if market is open (data available after market close)
- Try again later if API is temporarily unavailable

### Go Server Issues

**Network errors (China users):**
```bash
cd server
setup.bat  # Windows
# or
./setup.sh  # Linux/Mac
```

See `server/NETWORK_SETUP.md` for detailed guide.

**Cannot download Go modules:**
```bash
# Use China mirror
go env -w GOPROXY=https://goproxy.cn,direct
cd server
go mod download
```

**Port already in use:**
- Change port in `server/main.go`: `r.Run(":8080")`

**Database path error:**
- Make sure running from `server/` directory
- Or update path in `main.go`: `initDB("../jia-stk.db")`

## ✨ Key Features Explained

### Historical Data (TDX)
- Import years of historical OHLCV data
- Stored in `stock_history` table
- Indexed by stock code and date for fast queries
- Supports SH, SZ, and BJ exchanges

### Daily Data (Sina)
- Fresh daily trading data
- Includes P/E, P/B ratios
- Market cap and turnover
- Separate table per date for easy management

### Smart Search
- Search by stock name (Chinese)
- Search by stock code
- Search by pinyin (e.g., "pfyh" finds "浦发银行")

### Web Interface
- Modern Vue 3 dashboard
- ECharts visualizations
- Historical data charts
- Top gainers ranking
- Industry distribution

## 📈 Future Enhancements

- [ ] Real-time data streaming
- [ ] Technical indicators calculation
- [ ] Advanced charting features
- [ ] Email/SMS alerts
- [ ] Portfolio tracking
- [ ] Backtesting framework
- [ ] Strategy scanner integration

## 📄 License

This project is for educational and personal use.

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📞 Support

Check the documentation in the `docs/` folder or:
- Python tools: `python dump.py --help`
- Query tool: `python query_db.py --help`
- API docs: `server/README.md`

---

**Happy Trading! 📈🚀**
