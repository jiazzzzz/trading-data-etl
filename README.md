# 📊 Stock Data Management System

A complete Python + Go solution for collecting, storing, and querying Chinese stock market data.

## 🎯 Features

- ✅ **Data Collection** - Daily stock data from Tushare & Sina Finance
- ✅ **SQLite Storage** - Lightweight, file-based database
- ✅ **Python Tools** - Data dumping and querying utilities
- ✅ **REST API** - Go web server with Gin framework
- ✅ **Web UI** - Interactive browser interface
- ✅ **Search** - By name, symbol, or pinyin
- ✅ **Analytics** - Top gainers, custom queries

## 🚀 Quick Start

### 1. Install Python Dependencies
```bash
pip install -r requirements
```

### 2. Dump Stock Data
```bash
# Full dump (stock list + daily data)
python dump.py

# Only stock list
python dump.py --stock-list-only

# Only daily data
python dump.py --daily-only
```

### 3. Query Data (Python)
```bash
# Interactive mode
python query_db.py

# Quick queries
python query_db.py --tables
python query_db.py --show stock_list
python query_db.py --sql "SELECT * FROM stock_list WHERE area='上海' LIMIT 10"
```

### 4. Start Web Server (Go)
```bash
cd server
go mod download
go run main.go
```

Server runs at: http://localhost:8080

### 5. Use Web UI
Open `server/index.html` in your browser for an interactive interface.

## 📁 Project Structure

```
jia-stk/
├── dump.py                 # Main data dumping script
├── query_db.py            # Python query tool
├── jia-stk.db             # SQLite database (created after first dump)
├── settings.conf          # Database configuration
├── requirements           # Python dependencies
│
├── lib/                   # Python libraries
│   ├── db.py             # MySQL database handler
│   ├── db_sqlite.py      # SQLite database handler
│   ├── stock_info.py     # Stock data fetching
│   ├── logger.py         # Logging utility
│   └── common.py         # Common utilities
│
├── server/               # Go REST API server
│   ├── main.go          # Server code
│   ├── go.mod           # Go dependencies
│   ├── index.html       # Simple web UI
│   ├── README.md        # API documentation
│   └── API_EXAMPLES.md  # Usage examples
│
├── frontend/            # Modern Vue 3 dashboard ⭐ NEW!
│   ├── index.html      # Main application
│   ├── app.js          # Vue 3 logic
│   ├── README.md       # Frontend docs
│   ├── QUICK_START.md  # Quick guide
│   └── FEATURES.md     # Feature list
│
└── docs/               # Documentation
    ├── QUICK_START.md
    ├── README_DATABASE.md
    ├── FRONTEND_GUIDE.md
    └── SERVER_GUIDE.md
```

## 📚 Documentation

- **[QUICK_START.md](QUICK_START.md)** - Get started in 5 minutes
- **[README_DATABASE.md](README_DATABASE.md)** - Database configuration
- **[EXAMPLES.md](EXAMPLES.md)** - Query examples
- **[SERVER_GUIDE.md](SERVER_GUIDE.md)** - Web server guide
- **[server/README.md](server/README.md)** - Full API documentation
- **[server/API_EXAMPLES.md](server/API_EXAMPLES.md)** - API usage examples

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

## 💻 Python Tools

### dump.py - Data Collection
```bash
python dump.py                    # Full dump
python dump.py --stock-list-only  # Only stock list
python dump.py --daily-only       # Only daily data
python dump.py --date 20250108    # Specific date
python dump.py --force            # Force overwrite
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
- `ts_code` - Tushare code
- `symbol` - Stock symbol
- `name` - Stock name
- `area` - Location
- `industry` - Industry sector
- `list_date` - Listing date
- `pinyin` - Pinyin for search

### stock_daily_YYYYMMDD
- `symbol` - Stock symbol
- `name` - Stock name
- `trade` - Current price
- `changepercent` - Change percentage
- `volume` - Trading volume
- `amount` - Trading amount
- And more...

## 🔍 Common Queries

### Python
```bash
# Count stocks
python query_db.py --sql "SELECT COUNT(*) FROM stock_list"

# Find tech stocks
python query_db.py --sql "SELECT * FROM stock_list WHERE industry LIKE '%科技%'"

# Top gainers
python query_db.py --sql "SELECT name, changepercent FROM stock_daily_20251107 ORDER BY changepercent DESC LIMIT 10"
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

### Windows Task Scheduler
Create a daily task:
```
Program: python
Arguments: C:\work\jia-stk\dump.py
Start in: C:\work\jia-stk
```

### Linux Cron
```bash
# Run at 4 PM on weekdays
0 16 * * 1-5 cd /path/to/jia-stk && python dump.py
```

## 🛠️ Development

### Python
```bash
# Install dependencies
pip install -r requirements

# Run tests
python dump.py --help
python query_db.py --help
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

**Database not found:**
```bash
python dump.py  # Create database first
```

**Tushare API errors:**
- Check your API key in `lib/stock_info.py`
- Verify network connection

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

## 📈 Future Enhancements

- [ ] Real-time data streaming
- [ ] Technical indicators calculation
- [ ] Chart visualization
- [ ] Email/SMS alerts
- [ ] Portfolio tracking
- [ ] Backtesting framework

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
