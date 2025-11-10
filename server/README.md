# Stock Data API Server

A simple Go web server using Gin framework to query SQLite stock database.

## Quick Start

### 1. Install Dependencies

```bash
cd server
go mod download
```

### 2. Run Server

```bash
go run main.go
```

Server will start at: http://localhost:8080

## API Endpoints

### Home
```
GET /
```
Returns API information and available endpoints.

### List All Tables
```
GET /api/tables
```
Returns all tables in the database.

**Example:**
```bash
curl http://localhost:8080/api/tables
```

### Get Stock List
```
GET /api/stocks?limit=10&offset=0
```
Returns paginated stock list.

**Parameters:**
- `limit` (optional): Number of records (default: 10)
- `offset` (optional): Offset for pagination (default: 0)

**Example:**
```bash
curl http://localhost:8080/api/stocks?limit=20
```

### Get Stock by Symbol
```
GET /api/stocks/:symbol
```
Returns details for a specific stock.

**Example:**
```bash
curl http://localhost:8080/api/stocks/000001
```

### Get Daily Data
```
GET /api/daily/:date?limit=10&offset=0
```
Returns daily trading data for a specific date.

**Parameters:**
- `date` (required): Date in YYYYMMDD format
- `limit` (optional): Number of records (default: 10)
- `offset` (optional): Offset for pagination (default: 0)

**Example:**
```bash
curl http://localhost:8080/api/daily/20251107?limit=20
```

### Get Top Gainers
```
GET /api/daily/:date/top?limit=10
```
Returns top gaining stocks for a specific date.

**Parameters:**
- `date` (required): Date in YYYYMMDD format
- `limit` (optional): Number of records (default: 10)

**Example:**
```bash
curl http://localhost:8080/api/daily/20251107/top?limit=10
```

### Search Stocks
```
GET /api/search?q=keyword&limit=20
```
Search stocks by name, symbol, or pinyin.

**Parameters:**
- `q` (required): Search keyword
- `limit` (optional): Number of records (default: 20)

**Example:**
```bash
curl http://localhost:8080/api/search?q=科技
curl http://localhost:8080/api/search?q=payx
```

### Custom SQL Query
```
GET /api/query?sql=SELECT...&limit=100
```
Execute custom SQL query (SELECT only).

**Parameters:**
- `sql` (required): SQL SELECT query
- `limit` (optional): Max results (default: 100, max: 1000)

**Example:**
```bash
curl "http://localhost:8080/api/query?sql=SELECT%20*%20FROM%20stock_list%20WHERE%20area='上海'%20LIMIT%2010"
```

## Usage Examples

### Using curl

```bash
# Get all tables
curl http://localhost:8080/api/tables

# Get first 10 stocks
curl http://localhost:8080/api/stocks

# Search for tech stocks
curl http://localhost:8080/api/search?q=科技

# Get top 10 gainers
curl http://localhost:8080/api/daily/20251107/top?limit=10

# Custom query
curl "http://localhost:8080/api/query?sql=SELECT%20COUNT(*)%20FROM%20stock_list"
```

### Using Browser

Simply open in your browser:
- http://localhost:8080/
- http://localhost:8080/api/stocks
- http://localhost:8080/api/search?q=银行

### Using JavaScript (fetch)

```javascript
// Get stocks
fetch('http://localhost:8080/api/stocks?limit=10')
  .then(res => res.json())
  .then(data => console.log(data));

// Search
fetch('http://localhost:8080/api/search?q=科技')
  .then(res => res.json())
  .then(data => console.log(data));

// Top gainers
fetch('http://localhost:8080/api/daily/20251107/top?limit=10')
  .then(res => res.json())
  .then(data => console.log(data));
```

## Build for Production

```bash
# Build binary
go build -o stock-server main.go

# Run
./stock-server
```

### Cross-compile for Linux
```bash
GOOS=linux GOARCH=amd64 go build -o stock-server-linux main.go
```

## Configuration

The server looks for the database at `../jia-stk.db` (relative to server directory).

To change the database path, edit `main.go`:
```go
if err := initDB("path/to/your/database.db"); err != nil {
```

## Security Features

- CORS enabled for cross-origin requests
- Only SELECT queries allowed in custom query endpoint
- Result limits to prevent excessive data transfer
- SQL injection protection via parameterized queries

## Port Configuration

Default port is 8080. To change:
```go
r.Run(":8080")  // Change to your desired port
```

Or set via environment variable:
```bash
PORT=3000 go run main.go
```

Then update the code:
```go
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
r.Run(":" + port)
```

## Troubleshooting

**Database not found:**
- Make sure `jia-stk.db` exists in the parent directory
- Check the path in `initDB()` function

**Port already in use:**
- Change the port in `r.Run(":8080")`
- Or kill the process using the port

**Module errors:**
- Run `go mod tidy` to clean up dependencies
- Run `go mod download` to download missing packages
