package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Stock struct {
	TsCode   string `json:"ts_code"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Area     string `json:"area"`
	Industry string `json:"industry"`
	ListDate string `json:"list_date"`
	Pinyin   string `json:"pinyin"`
}

type StockWithTrading struct {
	TsCode        string   `json:"ts_code"`
	Symbol        string   `json:"symbol"`
	Name          string   `json:"name"`
	Pinyin        string   `json:"pinyin"`
	Trade         *float64 `json:"trade"`          // 现价
	ChangePercent *float64 `json:"changepercent"`  // 涨跌幅
	Mktcap        *float64 `json:"mktcap"`         // 市值
	TurnoverRatio *float64 `json:"turnoverratio"`  // 换手率
}

type DailyData struct {
	Symbol        string  `json:"symbol"`
	Code          int64   `json:"code"`
	Name          string  `json:"name"`
	Trade         float64 `json:"trade"`
	PriceChange   float64 `json:"pricechange"`
	ChangePercent float64 `json:"changepercent"`
	Buy           float64 `json:"buy"`
	Sell          float64 `json:"sell"`
	Settlement    float64 `json:"settlement"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        int64   `json:"volume"`
	Amount        int64   `json:"amount"`
	TickTime      string  `json:"ticktime"`
	Per           float64 `json:"per"`
	Pb            float64 `json:"pb"`
	Mktcap        float64 `json:"mktcap"`
	Nmc           float64 `json:"nmc"`
	TurnoverRatio float64 `json:"turnoverratio"`
	DumpTime      string  `json:"dump_time"`
}

type TableInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type HistoryData struct {
	ID         int64   `json:"id"`
	StockCode  string  `json:"stock_code"`
	StockName  string  `json:"stock_name"`
	Exchange   string  `json:"exchange"`
	TradeDate  string  `json:"trade_date"`
	Open       float64 `json:"open"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Close      float64 `json:"close"`
	Volume     int64   `json:"volume"`
	Amount     float64 `json:"amount"`
	ImportTime string  `json:"import_time"`
}

type HistoricalMover struct {
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Exchange      string  `json:"exchange"`
	TradeDate     string  `json:"trade_date"`
	Open          float64 `json:"open"`
	Close         float64 `json:"close"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        int64   `json:"volume"`
	Amount        float64 `json:"amount"`
	PrevClose     float64 `json:"prev_close"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

func initDB(dbPath string) error {
	var err error
	// Add _loc=auto for proper UTF-8 handling
	db, err = sql.Open("sqlite", dbPath+"?_loc=auto")
	if err != nil {
		return err
	}
	return db.Ping()
}

func main() {
	// Initialize database
	if err := initDB("../jia-stk.db"); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	r.GET("/api", homeHandler)
	r.GET("/api/tables", getTablesHandler)
	r.GET("/api/stats", getStatsHandler)
	r.GET("/api/stocks", getStocksHandler)
	r.GET("/api/stocks/:symbol", getStockBySymbolHandler)
	r.GET("/api/daily/:date", getDailyDataHandler)
	r.GET("/api/daily/:date/top", getTopGainersHandler)
	r.GET("/api/daily/:date/bottom", getTopLosersHandler)
	r.GET("/api/search", searchStocksHandler)
	r.GET("/api/query", customQueryHandler)
	r.GET("/api/history/:stock_code", getStockHistoryHandler)
	r.GET("/api/history/:stock_code/range", getStockHistoryRangeHandler)
	r.GET("/api/history/gainers/:date", getHistoricalGainersHandler)
	r.GET("/api/history/losers/:date", getHistoricalLosersHandler)
	r.GET("/api/history/dates/available", getAvailableDatesHandler)
	r.GET("/api/filter/stocks", filterStocksHandler)
	r.GET("/api/realtime/mktcap/:stock_code", getRealtimeMktcapHandler)
	r.GET("/api/strategy/scan", strategyScanHandler)
	
	// Serve static files from frontend directory
	r.StaticFile("/", "../frontend/index.html")
	r.StaticFile("/index.html", "../frontend/index.html")
	r.StaticFile("/history.html", "../frontend/history.html")
	r.StaticFile("/debug.html", "../frontend/debug.html")
	r.StaticFile("/test-cache.html", "../frontend/test-cache.html")
	r.StaticFile("/app.js", "../frontend/app.js")
	r.StaticFile("/echarts.min.js", "../frontend/echarts.min.js")

	log.Println("Server starting on http://localhost:8080")
	r.Run(":8080")
}

func homeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock Data API Server",
		"version": "1.0",
		"endpoints": []string{
			"GET /api/tables - List all tables",
			"GET /api/stocks?limit=10&offset=0 - Get stock list",
			"GET /api/stocks/:symbol - Get stock by symbol",
			"GET /api/daily/:date?limit=10 - Get daily data",
			"GET /api/daily/:date/top?limit=10 - Get top gainers",
			"GET /api/search?q=keyword - Search stocks",
			"GET /api/query?sql=SELECT... - Custom SQL query",
			"GET /api/history/:stock_code?limit=60 - Get stock history",
			"GET /api/history/:stock_code/range?start=20250801&end=20251108 - Get stock history by date range",
			"GET /api/history/gainers/:date?limit=10 - Get top gainers from history",
			"GET /api/history/losers/:date?limit=10 - Get top losers from history",
			"GET /api/history/dates/available?limit=7 - Get available trading dates",
		},
	})
}

func getTablesHandler(c *gin.Context) {
	rows, err := db.Query("SELECT name, type FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Type); err != nil {
			continue
		}
		tables = append(tables, table)
	}

	c.JSON(http.StatusOK, gin.H{
		"count":  len(tables),
		"tables": tables,
	})
}

func getStatsHandler(c *gin.Context) {
	// Get optional date parameter
	date := c.DefaultQuery("date", "")

	// Get total stocks
	var totalStocks int
	db.QueryRow("SELECT COUNT(*) FROM stock_list").Scan(&totalStocks)

	// Get total tables
	var totalTables int
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&totalTables)

	var gainers, losers int
	var tableName string

	if date != "" {
		// Use specific date's table
		tableName = fmt.Sprintf("stock_daily_%s", date)
		// Check if table exists
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&exists)
		if exists == 0 {
			tableName = ""
		}
	} else {
		// Get latest daily table
		db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 1").Scan(&tableName)
	}

	if tableName != "" {
		// Count gainers (changepercent > 0)
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE changepercent > 0", tableName)
		db.QueryRow(query).Scan(&gainers)

		// Count losers (changepercent < 0)
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE changepercent < 0", tableName)
		db.QueryRow(query).Scan(&losers)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_stocks": totalStocks,
		"gainers":      gainers,
		"losers":       losers,
		"tables":       totalTables,
		"date":         date,
	})
}

func getStocksHandler(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	// Get latest daily table
	var latestTable string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 1").Scan(&latestTable)
	
	var stocks []StockWithTrading
	
	if err != nil || latestTable == "" {
		// No daily table found, return stocks without trading data
		query := "SELECT ts_code, symbol, name, pinyin FROM stock_list LIMIT ? OFFSET ?"
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var stock StockWithTrading
			if err := rows.Scan(&stock.TsCode, &stock.Symbol, &stock.Name, &stock.Pinyin); err != nil {
				continue
			}
			stocks = append(stocks, stock)
		}
	} else {
		// Join with latest daily table, using MAX(rowid) to handle duplicates
		query := fmt.Sprintf(`
			SELECT 
				sl.ts_code, sl.symbol, sl.name, sl.pinyin,
				d.trade, d.changepercent, d.mktcap, d.turnoverratio
			FROM stock_list sl
			LEFT JOIN (
				SELECT 
					REPLACE(REPLACE(REPLACE(symbol, 'sh', ''), 'sz', ''), 'bj', '') as clean_symbol,
					MAX(trade) as trade, 
					MAX(changepercent) as changepercent, 
					MAX(mktcap) as mktcap, 
					MAX(turnoverratio) as turnoverratio
				FROM %s
				GROUP BY clean_symbol
			) d ON sl.symbol = d.clean_symbol
			LIMIT ? OFFSET ?
		`, latestTable)
		
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var stock StockWithTrading
			if err := rows.Scan(&stock.TsCode, &stock.Symbol, &stock.Name, &stock.Pinyin, 
				&stock.Trade, &stock.ChangePercent, &stock.Mktcap, &stock.TurnoverRatio); err != nil {
				continue
			}
			stocks = append(stocks, stock)
		}
	}

	// Get total count
	var total int
	db.QueryRow("SELECT COUNT(*) FROM stock_list").Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"total":  total,
		"count":  len(stocks),
		"stocks": stocks,
	})
}

func getStockBySymbolHandler(c *gin.Context) {
	symbol := c.Param("symbol")

	var stock Stock
	query := "SELECT ts_code, symbol, name, area, industry, list_date, pinyin FROM stock_list WHERE symbol = ?"
	err := db.QueryRow(query, symbol).Scan(&stock.TsCode, &stock.Symbol, &stock.Name, &stock.Area, &stock.Industry, &stock.ListDate, &stock.Pinyin)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stock)
}

func getDailyDataHandler(c *gin.Context) {
	date := c.Param("date")
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	tableName := "stock_daily_" + date

	// Check if table exists
	var exists int
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&exists)
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No data for date: " + date})
		return
	}

	query := "SELECT symbol, code, name, trade, pricechange, changepercent, buy, sell, settlement, open, high, low, volume, amount, ticktime, per, pb, mktcap, nmc, turnoverratio, dump_time FROM " + tableName + " LIMIT ? OFFSET ?"
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dailyData []DailyData
	for rows.Next() {
		var data DailyData
		if err := rows.Scan(&data.Symbol, &data.Code, &data.Name, &data.Trade, &data.PriceChange, &data.ChangePercent, &data.Buy, &data.Sell, &data.Settlement, &data.Open, &data.High, &data.Low, &data.Volume, &data.Amount, &data.TickTime, &data.Per, &data.Pb, &data.Mktcap, &data.Nmc, &data.TurnoverRatio, &data.DumpTime); err != nil {
			continue
		}
		dailyData = append(dailyData, data)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":  date,
		"count": len(dailyData),
		"data":  dailyData,
	})
}

func getTopGainersHandler(c *gin.Context) {
	date := c.Param("date")
	limit := c.DefaultQuery("limit", "10")

	tableName := "stock_daily_" + date

	// Check if table exists
	var exists int
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&exists)
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No data for date: " + date})
		return
	}

	query := "SELECT symbol, code, name, trade, pricechange, changepercent, buy, sell, settlement, open, high, low, volume, amount, ticktime, per, pb, mktcap, nmc, turnoverratio, dump_time FROM " + tableName + " ORDER BY changepercent DESC LIMIT ?"
	rows, err := db.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dailyData []DailyData
	for rows.Next() {
		var data DailyData
		if err := rows.Scan(&data.Symbol, &data.Code, &data.Name, &data.Trade, &data.PriceChange, &data.ChangePercent, &data.Buy, &data.Sell, &data.Settlement, &data.Open, &data.High, &data.Low, &data.Volume, &data.Amount, &data.TickTime, &data.Per, &data.Pb, &data.Mktcap, &data.Nmc, &data.TurnoverRatio, &data.DumpTime); err != nil {
			continue
		}
		dailyData = append(dailyData, data)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        date,
		"count":       len(dailyData),
		"top_gainers": dailyData,
	})
}

func getTopLosersHandler(c *gin.Context) {
	date := c.Param("date")
	limit := c.DefaultQuery("limit", "10")

	tableName := "stock_daily_" + date

	// Check if table exists
	var exists int
	db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&exists)
	if exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No data for date: " + date})
		return
	}

	query := "SELECT symbol, code, name, trade, pricechange, changepercent, buy, sell, settlement, open, high, low, volume, amount, ticktime, per, pb, mktcap, nmc, turnoverratio, dump_time FROM " + tableName + " ORDER BY changepercent ASC LIMIT ?"
	rows, err := db.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dailyData []DailyData
	for rows.Next() {
		var data DailyData
		if err := rows.Scan(&data.Symbol, &data.Code, &data.Name, &data.Trade, &data.PriceChange, &data.ChangePercent, &data.Buy, &data.Sell, &data.Settlement, &data.Open, &data.High, &data.Low, &data.Volume, &data.Amount, &data.TickTime, &data.Per, &data.Pb, &data.Mktcap, &data.Nmc, &data.TurnoverRatio, &data.DumpTime); err != nil {
			continue
		}
		dailyData = append(dailyData, data)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        date,
		"count":       len(dailyData),
		"top_losers": dailyData,
	})
}

func searchStocksHandler(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit := c.DefaultQuery("limit", "20")
	searchTerm := "%" + q + "%"

	query := "SELECT ts_code, symbol, name, area, industry, list_date, pinyin FROM stock_list WHERE name LIKE ? OR symbol LIKE ? OR pinyin LIKE ? LIMIT ?"
	rows, err := db.Query(query, searchTerm, searchTerm, searchTerm, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var stocks []Stock
	for rows.Next() {
		var stock Stock
		if err := rows.Scan(&stock.TsCode, &stock.Symbol, &stock.Name, &stock.Area, &stock.Industry, &stock.ListDate, &stock.Pinyin); err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}

	c.JSON(http.StatusOK, gin.H{
		"query":  q,
		"count":  len(stocks),
		"stocks": stocks,
	})
}

func customQueryHandler(c *gin.Context) {
	sqlQuery := c.Query("sql")
	if sqlQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'sql' is required"})
		return
	}

	// Security: Only allow SELECT queries
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlQuery)), "SELECT") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only SELECT queries are allowed"})
		return
	}

	// Limit results
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 1000 {
		limit = 1000
	}

	rows, err := db.Query(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare result
	var results []map[string]interface{}
	for rows.Next() && len(results) < limit {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   sqlQuery,
		"count":   len(results),
		"results": results,
	})
}

func getStockHistoryHandler(c *gin.Context) {
	stockCode := c.Param("stock_code")
	limit := c.DefaultQuery("limit", "60")
	offset := c.DefaultQuery("offset", "0")

	query := `SELECT id, stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount, import_time 
	          FROM stock_history 
	          WHERE stock_code = ? 
	          ORDER BY trade_date DESC 
	          LIMIT ? OFFSET ?`

	rows, err := db.Query(query, stockCode, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var historyData []HistoryData
	for rows.Next() {
		var data HistoryData
		if err := rows.Scan(&data.ID, &data.StockCode, &data.StockName, &data.Exchange, &data.TradeDate, &data.Open, &data.High, &data.Low, &data.Close, &data.Volume, &data.Amount, &data.ImportTime); err != nil {
			continue
		}
		historyData = append(historyData, data)
	}

	// Get total count for this stock
	var total int
	db.QueryRow("SELECT COUNT(*) FROM stock_history WHERE stock_code = ?", stockCode).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"stock_code": stockCode,
		"total":      total,
		"count":      len(historyData),
		"data":       historyData,
	})
}

func getStockHistoryRangeHandler(c *gin.Context) {
	stockCode := c.Param("stock_code")
	startDate := c.Query("start")
	endDate := c.Query("end")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both 'start' and 'end' query parameters are required (format: YYYYMMDD)"})
		return
	}

	query := `SELECT id, stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount, import_time 
	          FROM stock_history 
	          WHERE stock_code = ? AND trade_date >= ? AND trade_date <= ? 
	          ORDER BY trade_date ASC`

	rows, err := db.Query(query, stockCode, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var historyData []HistoryData
	for rows.Next() {
		var data HistoryData
		if err := rows.Scan(&data.ID, &data.StockCode, &data.StockName, &data.Exchange, &data.TradeDate, &data.Open, &data.High, &data.Low, &data.Close, &data.Volume, &data.Amount, &data.ImportTime); err != nil {
			continue
		}
		historyData = append(historyData, data)
	}

	c.JSON(http.StatusOK, gin.H{
		"stock_code": stockCode,
		"start_date": startDate,
		"end_date":   endDate,
		"count":      len(historyData),
		"data":       historyData,
	})
}

func getHistoricalGainersHandler(c *gin.Context) {
	date := c.Param("date")
	limit := c.DefaultQuery("limit", "10")

	// Query to get stocks with their previous day's close price and calculate change
	query := `
		WITH current_day AS (
			SELECT stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount
			FROM stock_history
			WHERE trade_date = ?
		),
		previous_day AS (
			SELECT h1.stock_code, h1.close as prev_close
			FROM stock_history h1
			INNER JOIN (
				SELECT stock_code, MAX(trade_date) as prev_date
				FROM stock_history
				WHERE trade_date < ?
				GROUP BY stock_code
			) h2 ON h1.stock_code = h2.stock_code AND h1.trade_date = h2.prev_date
		)
		SELECT 
			c.stock_code,
			c.stock_name,
			c.exchange,
			c.trade_date,
			c.open,
			c.close,
			c.high,
			c.low,
			c.volume,
			c.amount,
			COALESCE(p.prev_close, c.open) as prev_close,
			c.close - COALESCE(p.prev_close, c.open) as change,
			((c.close - COALESCE(p.prev_close, c.open)) / COALESCE(p.prev_close, c.open) * 100) as change_percent
		FROM current_day c
		LEFT JOIN previous_day p ON c.stock_code = p.stock_code
		WHERE COALESCE(p.prev_close, c.open) > 0
		ORDER BY change_percent DESC
		LIMIT ?
	`

	rows, err := db.Query(query, date, date, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var movers []HistoricalMover
	for rows.Next() {
		var mover HistoricalMover
		if err := rows.Scan(
			&mover.StockCode,
			&mover.StockName,
			&mover.Exchange,
			&mover.TradeDate,
			&mover.Open,
			&mover.Close,
			&mover.High,
			&mover.Low,
			&mover.Volume,
			&mover.Amount,
			&mover.PrevClose,
			&mover.Change,
			&mover.ChangePercent,
		); err != nil {
			continue
		}
		movers = append(movers, mover)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        date,
		"count":       len(movers),
		"top_gainers": movers,
	})
}

func getHistoricalLosersHandler(c *gin.Context) {
	date := c.Param("date")
	limit := c.DefaultQuery("limit", "10")

	// Query to get stocks with their previous day's close price and calculate change
	query := `
		WITH current_day AS (
			SELECT stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount
			FROM stock_history
			WHERE trade_date = ?
		),
		previous_day AS (
			SELECT h1.stock_code, h1.close as prev_close
			FROM stock_history h1
			INNER JOIN (
				SELECT stock_code, MAX(trade_date) as prev_date
				FROM stock_history
				WHERE trade_date < ?
				GROUP BY stock_code
			) h2 ON h1.stock_code = h2.stock_code AND h1.trade_date = h2.prev_date
		)
		SELECT 
			c.stock_code,
			c.stock_name,
			c.exchange,
			c.trade_date,
			c.open,
			c.close,
			c.high,
			c.low,
			c.volume,
			c.amount,
			COALESCE(p.prev_close, c.open) as prev_close,
			c.close - COALESCE(p.prev_close, c.open) as change,
			((c.close - COALESCE(p.prev_close, c.open)) / COALESCE(p.prev_close, c.open) * 100) as change_percent
		FROM current_day c
		LEFT JOIN previous_day p ON c.stock_code = p.stock_code
		WHERE COALESCE(p.prev_close, c.open) > 0
		ORDER BY change_percent ASC
		LIMIT ?
	`

	rows, err := db.Query(query, date, date, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var movers []HistoricalMover
	for rows.Next() {
		var mover HistoricalMover
		if err := rows.Scan(
			&mover.StockCode,
			&mover.StockName,
			&mover.Exchange,
			&mover.TradeDate,
			&mover.Open,
			&mover.Close,
			&mover.High,
			&mover.Low,
			&mover.Volume,
			&mover.Amount,
			&mover.PrevClose,
			&mover.Change,
			&mover.ChangePercent,
		); err != nil {
			continue
		}
		movers = append(movers, mover)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":       date,
		"count":      len(movers),
		"top_losers": movers,
	})
}

func getAvailableDatesHandler(c *gin.Context) {
	limit := c.DefaultQuery("limit", "7")

	// Get dates from stock_daily_* tables (actual trading days only)
	query := `
		SELECT REPLACE(name, 'stock_daily_', '') as trade_date
		FROM sqlite_master 
		WHERE type='table' AND name LIKE 'stock_daily_%'
		ORDER BY name DESC 
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			continue
		}
		dates = append(dates, date)
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(dates),
		"dates": dates,
	})
}

type FilteredStock struct {
	StockCode     string   `json:"stock_code"`
	StockName     string   `json:"stock_name"`
	Exchange      string   `json:"exchange"`
	MaxChange     float64  `json:"max_change_percent"`
	MinChange     float64  `json:"min_change_percent"`
	AvgChange     float64  `json:"avg_change_percent"`
	DaysCount     int      `json:"days_count"`
	LatestPrice   float64  `json:"latest_price"`
	LatestDate    string   `json:"latest_date"`
	Mktcap        float64  `json:"mktcap"`
	TriggerDates  []string `json:"trigger_dates"`
}

func filterStocksHandler(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	minChangePercent := c.DefaultQuery("min_change", "0")
	maxChangePercent := c.DefaultQuery("max_change", "100")
	maxMktcapStr := c.DefaultQuery("max_mktcap", "10000")
	limit := c.DefaultQuery("limit", "50")

	// First get the recent dates
	daysInt, _ := strconv.Atoi(days)
	limitInt, _ := strconv.Atoi(limit)
	minChange, _ := strconv.ParseFloat(minChangePercent, 64)
	maxChange, _ := strconv.ParseFloat(maxChangePercent, 64)
	maxMktcap, _ := strconv.ParseFloat(maxMktcapStr, 64)
	maxMktcapValue := maxMktcap * 100000000 // Convert 亿 to actual value
	
	log.Printf("Filter query params: days=%d, min=%.2f, max=%.2f, mktcap<=%.2f亿, limit=%d", daysInt, minChange, maxChange, maxMktcap, limitInt)
	
	// Get recent dates
	dateQuery := `SELECT DISTINCT trade_date FROM stock_history ORDER BY trade_date DESC LIMIT ?`
	dateRows, err := db.Query(dateQuery, daysInt)
	if err != nil {
		log.Printf("Date query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	var dates []string
	for dateRows.Next() {
		var date string
		dateRows.Scan(&date)
		dates = append(dates, date)
	}
	dateRows.Close()
	
	if len(dates) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"days":            days,
			"min_change":      minChangePercent,
			"max_change":      maxChangePercent,
			"count":           0,
			"filtered_stocks": []FilteredStock{},
		})
		return
	}
	
	log.Printf("Found %d recent dates", len(dates))
	
	// Build date list for IN clause
	dateList := "'" + strings.Join(dates, "','") + "'"
	
	// Get stocks that meet criteria and collect trigger dates
	// Calculate change from previous day's close to current day's close
	query := `
		WITH daily_changes AS (
			SELECT 
				h1.stock_code,
				h1.stock_name,
				h1.exchange,
				h1.trade_date,
				h1.close,
				h2.close as prev_close,
				CASE 
					WHEN h2.close > 0 THEN ((h1.close - h2.close) / h2.close * 100)
					ELSE 0
				END as change_percent
			FROM stock_history h1
			LEFT JOIN stock_history h2 ON h1.stock_code = h2.stock_code 
				AND h2.trade_date = (
					SELECT MAX(trade_date) 
					FROM stock_history 
					WHERE stock_code = h1.stock_code AND trade_date < h1.trade_date
				)
			WHERE h1.trade_date IN (` + dateList + `)
		)
		SELECT 
			stock_code,
			stock_name,
			exchange,
			trade_date,
			change_percent
		FROM daily_changes
		WHERE prev_close > 0
			AND change_percent >= ?
			AND change_percent <= ?
		ORDER BY stock_code, trade_date DESC
	`
	
	rows, err := db.Query(query, minChange, maxChange)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Group by stock and collect trigger dates
	stockMap := make(map[string]*FilteredStock)
	var stockOrder []string
	
	for rows.Next() {
		var stockCode, stockName, exchange, tradeDate string
		var changePercent float64
		
		if err := rows.Scan(&stockCode, &stockName, &exchange, &tradeDate, &changePercent); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		
		if _, exists := stockMap[stockCode]; !exists {
			stockMap[stockCode] = &FilteredStock{
				StockCode:    stockCode,
				StockName:    stockName,
				Exchange:     exchange,
				MaxChange:    changePercent,
				MinChange:    changePercent,
				AvgChange:    changePercent,
				DaysCount:    1,
				LatestDate:   tradeDate,
				TriggerDates: []string{tradeDate},
			}
			stockOrder = append(stockOrder, stockCode)
		} else {
			stock := stockMap[stockCode]
			if changePercent > stock.MaxChange {
				stock.MaxChange = changePercent
			}
			if changePercent < stock.MinChange {
				stock.MinChange = changePercent
			}
			stock.AvgChange = (stock.AvgChange*float64(stock.DaysCount) + changePercent) / float64(stock.DaysCount+1)
			stock.DaysCount++
			stock.TriggerDates = append(stock.TriggerDates, tradeDate)
			if tradeDate > stock.LatestDate {
				stock.LatestDate = tradeDate
			}
		}
	}
	
	log.Printf("Found %d unique stocks", len(stockMap))
	
	// Find the most recent daily table that exists
	var mostRecentDailyTable string
	for _, date := range dates {
		tableName := "stock_daily_" + date
		var tableExists int
		db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&tableExists)
		if tableExists > 0 {
			mostRecentDailyTable = tableName
			log.Printf("Using daily table: %s for market cap data", mostRecentDailyTable)
			break
		}
	}
	
	// Convert map to slice and get additional data
	var stocks []FilteredStock
	for _, stockCode := range stockOrder {
		stock := stockMap[stockCode]
		
		// Get latest price
		var latestPrice float64
		priceQuery := `SELECT close FROM stock_history WHERE stock_code = ? AND trade_date = ?`
		db.QueryRow(priceQuery, stock.StockCode, stock.LatestDate).Scan(&latestPrice)
		stock.LatestPrice = latestPrice
		
		// Try to get market cap from the most recent daily table
		if mostRecentDailyTable != "" {
			// Build symbol with exchange prefix (e.g., "sz000001", "sh600000", "bj920003")
			var symbolWithPrefix string
			switch stock.Exchange {
			case "SZ":
				symbolWithPrefix = "sz" + stock.StockCode
			case "SH":
				symbolWithPrefix = "sh" + stock.StockCode
			case "BJ":
				symbolWithPrefix = "bj" + stock.StockCode
			default:
				symbolWithPrefix = strings.ToLower(stock.Exchange) + stock.StockCode
			}
			
			var mktcap float64
			mktcapQuery := `SELECT mktcap FROM ` + mostRecentDailyTable + ` WHERE symbol = ?`
			err := db.QueryRow(mktcapQuery, symbolWithPrefix).Scan(&mktcap)
			if err == nil && mktcap > 0 {
				stock.Mktcap = mktcap
				// Filter by market cap
				if maxMktcapValue > 0 && mktcap > maxMktcapValue {
					continue
				}
			}
		}
		
		stocks = append(stocks, *stock)
		
		if len(stocks) >= limitInt {
			break
		}
	}
	
	log.Printf("After filtering: %d stocks", len(stocks))

	c.JSON(http.StatusOK, gin.H{
		"days":              days,
		"min_change":        minChangePercent,
		"max_change":        maxChangePercent,
		"count":             len(stocks),
		"filtered_stocks":   stocks,
	})
}

// Market cap cache
type MktcapCache struct {
	mu    sync.RWMutex
	cache map[string]CachedMktcap
}

type CachedMktcap struct {
	Value     float64
	Timestamp time.Time
}

var mktcapCache = &MktcapCache{
	cache: make(map[string]CachedMktcap),
}

func (mc *MktcapCache) Get(key string) (float64, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	cached, exists := mc.cache[key]
	if !exists {
		return 0, false
	}
	
	// Cache for 5 minutes
	if time.Since(cached.Timestamp) > 5*time.Minute {
		return 0, false
	}
	
	return cached.Value, true
}

func (mc *MktcapCache) Set(key string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.cache[key] = CachedMktcap{
		Value:     value,
		Timestamp: time.Now(),
	}
}

func getRealtimeMktcapHandler(c *gin.Context) {
	stockCode := c.Param("stock_code")
	
	// Check cache first
	if cached, exists := mktcapCache.Get(stockCode); exists {
		c.JSON(http.StatusOK, gin.H{
			"stock_code": stockCode,
			"mktcap":     cached,
			"cached":     true,
		})
		return
	}
	
	// Determine market prefix for Sina API
	var sinaCode string
	if strings.HasPrefix(stockCode, "6") {
		sinaCode = "sh" + stockCode // Shanghai
	} else if strings.HasPrefix(stockCode, "0") || strings.HasPrefix(stockCode, "3") {
		sinaCode = "sz" + stockCode // Shenzhen
	} else if strings.HasPrefix(stockCode, "8") || strings.HasPrefix(stockCode, "4") || strings.HasPrefix(stockCode, "9") {
		sinaCode = "bj" + stockCode // Beijing
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock code"})
		return
	}
	
	// Fetch from Sina Finance API
	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", sinaCode)
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching mktcap for %s: %v", stockCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response for %s: %v", stockCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read data"})
		return
	}
	
	// Parse Sina response format: var hq_str_sh600000="浦发银行,10.50,10.48,..."
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "=") {
		log.Printf("Invalid response for %s: %s", stockCode, bodyStr)
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}
	
	// Extract data between quotes
	start := strings.Index(bodyStr, "\"")
	end := strings.LastIndex(bodyStr, "\"")
	if start == -1 || end == -1 || start >= end {
		log.Printf("Cannot parse response for %s", stockCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse data"})
		return
	}
	
	data := bodyStr[start+1 : end]
	fields := strings.Split(data, ",")
	
	// Sina API returns 32+ fields, market cap is typically at index 44 (total) and 45 (circulating)
	// But for simplicity, we'll calculate: current price * total shares
	// Field 3: current price, Field 12: total shares (in 万股)
	if len(fields) < 13 {
		log.Printf("Insufficient fields for %s: got %d fields", stockCode, len(fields))
		c.JSON(http.StatusNotFound, gin.H{"error": "Insufficient data"})
		return
	}
	
	currentPrice, err1 := strconv.ParseFloat(fields[3], 64)
	totalShares, err2 := strconv.ParseFloat(fields[12], 64) // in 万股
	
	if err1 != nil || err2 != nil || currentPrice <= 0 || totalShares <= 0 {
		log.Printf("Invalid data for %s: price=%s, shares=%s", stockCode, fields[3], fields[12])
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid market data"})
		return
	}
	
	// Calculate market cap: price * shares (万股) = 万元
	// This matches the database format which stores mktcap in 万元
	mktcapWan := currentPrice * totalShares
	mktcapYi := mktcapWan / 10000
	
	if mktcapWan > 0 {
		// Cache the result (in 万元 to match database format)
		mktcapCache.Set(stockCode, mktcapWan)
		
		log.Printf("Fetched mktcap for %s: %.2f亿 (price=%.2f, shares=%.0f万, mktcap=%.0f万)", stockCode, mktcapYi, currentPrice, totalShares, mktcapWan)
		
		c.JSON(http.StatusOK, gin.H{
			"stock_code": stockCode,
			"mktcap":     mktcapWan,
			"mktcap_yi":  mktcapYi,
			"cached":     false,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Market cap not available"})
	}
}

type StrategyResult struct {
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Symbol        string  `json:"symbol"`
	TradeDate     string  `json:"trade_date"`
	Close         float64 `json:"close"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        int64   `json:"volume"`
	Amount        float64 `json:"amount"`
	ChangePercent float64 `json:"change_percent"`
	TurnoverRatio float64 `json:"turnover_ratio"`
	PrevVolume    int64   `json:"prev_volume"`
	PrevChange    float64 `json:"prev_change_percent"`
	VolumeRatio   float64 `json:"volume_ratio"`
	Mktcap        float64 `json:"mktcap"`
}

func strategyScanHandler(c *gin.Context) {
	// Get strategy parameters
	date := c.DefaultQuery("date", "")
	volumeMultiplier := c.DefaultQuery("volume_multiplier", "2.0")
	minChangeIncrease := c.DefaultQuery("min_change_increase", "5.0")
	minTurnover := c.DefaultQuery("min_turnover", "5.0")
	maxMktcapStr := c.DefaultQuery("max_mktcap", "0")
	limit := c.DefaultQuery("limit", "50")

	volumeMult, _ := strconv.ParseFloat(volumeMultiplier, 64)
	minChangeInc, _ := strconv.ParseFloat(minChangeIncrease, 64)
	minTurn, _ := strconv.ParseFloat(minTurnover, 64)
	maxMktcap, _ := strconv.ParseFloat(maxMktcapStr, 64)
	maxMktcapValue := maxMktcap * 100000000 // Convert 亿 to actual value
	limitInt, _ := strconv.Atoi(limit)

	log.Printf("Strategy scan: date=%s, volume_mult=%.1f, change_inc=%.1f%%, turnover>=%.1f%%, mktcap<=%.0f亿",
		date, volumeMult, minChangeInc, minTurn, maxMktcap)

	// If no date specified, get the most recent trading date from stock_daily tables
	if date == "" {
		err := db.QueryRow(`
			SELECT REPLACE(name, 'stock_daily_', '') 
			FROM sqlite_master 
			WHERE type='table' AND name LIKE 'stock_daily_%' 
			ORDER BY name DESC 
			LIMIT 1
		`).Scan(&date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No trading data available"})
			return
		}
	}

	// Get previous trading date from stock_daily tables
	var prevDate string
	err := db.QueryRow(`
		SELECT REPLACE(name, 'stock_daily_', '') 
		FROM sqlite_master 
		WHERE type='table' AND name LIKE 'stock_daily_%' AND name < ?
		ORDER BY name DESC 
		LIMIT 1
	`, "stock_daily_"+date).Scan(&prevDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot find previous trading date"})
		return
	}

	log.Printf("Comparing %s with previous date %s", date, prevDate)

	// Query to find stocks matching the strategy
	// Get the date before previous date for calculating previous day's change
	var prevPrevDate string
	err2 := db.QueryRow(`
		SELECT DISTINCT trade_date 
		FROM stock_history 
		WHERE trade_date < ? 
		ORDER BY trade_date DESC 
		LIMIT 1
	`, prevDate).Scan(&prevPrevDate)
	
	if err2 != nil {
		log.Printf("Warning: Cannot find date before %s, prev_change_percent will be 0", prevDate)
		prevPrevDate = prevDate // Fallback to prevent errors
	}
	
	log.Printf("Using dates: current=%s, prev=%s, prevPrev=%s", date, prevDate, prevPrevDate)

	query := `
		WITH prev_prev_day AS (
			SELECT 
				stock_code,
				close as prev_prev_close
			FROM stock_history
			WHERE trade_date = ?
		),
		previous_day AS (
			SELECT 
				h.stock_code,
				h.close as prev_close,
				h.volume as prev_volume,
				CASE 
					WHEN COALESCE(pp.prev_prev_close, 0) > 0 
					THEN ((h.close - pp.prev_prev_close) / pp.prev_prev_close * 100)
					ELSE 0
				END as prev_change_percent
			FROM stock_history h
			LEFT JOIN prev_prev_day pp ON h.stock_code = pp.stock_code
			WHERE h.trade_date = ?
		),
		current_day AS (
			SELECT 
				h.stock_code,
				h.stock_name,
				h.exchange,
				h.trade_date,
				h.open,
				h.high,
				h.low,
				h.close,
				h.volume,
				h.amount,
				CASE 
					WHEN COALESCE(p.prev_close, 0) > 0 
					THEN ((h.close - p.prev_close) / p.prev_close * 100)
					ELSE 0
				END as change_percent
			FROM stock_history h
			LEFT JOIN previous_day p ON h.stock_code = p.stock_code
			WHERE h.trade_date = ?
		)
		SELECT 
			c.stock_code,
			c.stock_name,
			c.exchange,
			c.trade_date,
			c.open,
			c.close,
			c.high,
			c.low,
			c.volume,
			c.amount,
			c.change_percent,
			COALESCE(p.prev_volume, 0) as prev_volume,
			COALESCE(p.prev_change_percent, 0) as prev_change_percent,
			CASE 
				WHEN COALESCE(p.prev_volume, 0) > 0 
				THEN CAST(c.volume AS REAL) / CAST(p.prev_volume AS REAL)
				ELSE 0
			END as volume_ratio
		FROM current_day c
		LEFT JOIN previous_day p ON c.stock_code = p.stock_code
		WHERE c.volume > 0
			AND COALESCE(p.prev_volume, 0) > 0
			AND CAST(c.volume AS REAL) / CAST(p.prev_volume AS REAL) >= ?
			AND (c.change_percent - COALESCE(p.prev_change_percent, 0)) >= ?
		ORDER BY volume_ratio DESC, c.change_percent DESC
		LIMIT ?
	`

	rows, err := db.Query(query, prevPrevDate, prevDate, date, volumeMult, minChangeInc, limitInt*2)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []StrategyResult
	
	// Find the most recent daily table for turnover and mktcap data
	var dailyTable string
	db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'stock_daily_%' ORDER BY name DESC LIMIT 1").Scan(&dailyTable)
	
	log.Printf("Using daily table: %s for additional data", dailyTable)

	for rows.Next() {
		var result StrategyResult
		if err := rows.Scan(
			&result.StockCode,
			&result.StockName,
			&result.Symbol,
			&result.TradeDate,
			&result.Open,
			&result.Close,
			&result.High,
			&result.Low,
			&result.Volume,
			&result.Amount,
			&result.ChangePercent,
			&result.PrevVolume,
			&result.PrevChange,
			&result.VolumeRatio,
		); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		// Get turnover ratio and mktcap from daily table if available
		if dailyTable != "" {
			var symbolWithPrefix string
			switch result.Symbol {
			case "SZ":
				symbolWithPrefix = "sz" + result.StockCode
			case "SH":
				symbolWithPrefix = "sh" + result.StockCode
			case "BJ":
				symbolWithPrefix = "bj" + result.StockCode
			default:
				symbolWithPrefix = strings.ToLower(result.Symbol) + result.StockCode
			}

			var turnover, mktcap float64
			dailyQuery := `SELECT turnoverratio, mktcap FROM ` + dailyTable + ` WHERE symbol = ?`
			err := db.QueryRow(dailyQuery, symbolWithPrefix).Scan(&turnover, &mktcap)
			if err == nil {
				result.TurnoverRatio = turnover
				result.Mktcap = mktcap

				// Apply turnover filter
				if minTurn > 0 && turnover < minTurn {
					continue
				}

				// Apply market cap filter
				if maxMktcapValue > 0 && mktcap > maxMktcapValue {
					continue
				}
			}
		}

		results = append(results, result)

		if len(results) >= limitInt {
			break
		}
	}

	log.Printf("Found %d stocks matching strategy", len(results))

	c.JSON(http.StatusOK, gin.H{
		"date":              date,
		"prev_date":         prevDate,
		"volume_multiplier": volumeMultiplier,
		"min_change_increase": minChangeIncrease,
		"min_turnover":      minTurnover,
		"max_mktcap":        maxMktcapStr,
		"count":             len(results),
		"results":           results,
	})
}
