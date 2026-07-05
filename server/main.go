package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := initDB("../jia-stk.db"); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API routes
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

	// Tag management
	r.GET("/api/tags", getTagsHandler)
	r.POST("/api/tags", createTagHandler)
	r.PUT("/api/tags/:id", updateTagHandler)
	r.DELETE("/api/tags/:id", deleteTagHandler)

	// Stock tags
	r.GET("/api/stock-tags/:code", getStockTagsHandler)
	r.POST("/api/stock-tags/:code", addStockTagHandler)
	r.DELETE("/api/stock-tags/:code/:tagId", removeStockTagHandler)
	r.GET("/api/tags/:id/stocks", getStocksByTagHandler)

	// Watchlist
	r.GET("/api/watchlist", getWatchlistHandler)
	r.POST("/api/watchlist/:code", addToWatchlistHandler)
	r.DELETE("/api/watchlist/:code", removeFromWatchlistHandler)

	// Warninglist
	r.GET("/api/warninglist", getWarninglistHandler)
	r.POST("/api/warninglist/:code", addToWarninglistHandler)
	r.DELETE("/api/warninglist/:code", removeFromWarninglistHandler)

	// Strategy management
	r.GET("/api/strategies", getStrategiesHandler)
	r.POST("/api/strategies", createStrategyHandler)
	r.PUT("/api/strategies/:id", updateStrategyHandler)
	r.DELETE("/api/strategies/:id", deleteStrategyHandler)

	// Admin
	r.POST("/api/admin/daily-update", dailyUpdateHandler)
	r.GET("/api/admin/daily-update/status", dailyUpdateStatusHandler)
	r.POST("/api/admin/daily-update/stop", dailyUpdateStopHandler)
	r.POST("/api/admin/backfill", backfillHandler)
	r.GET("/api/admin/backfill/status", backfillStatusHandler)
	r.POST("/api/admin/backfill/stop", backfillStopHandler)
	r.GET("/api/admin/backfill/gaps", backfillGapsHandler)

	// Static files
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
