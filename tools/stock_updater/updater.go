package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Updater struct {
	db     *DB
	dbPath string
	worker int
}

func NewUpdater(db *DB, dbPath string) *Updater {
	return &Updater{db: db, dbPath: dbPath, worker: 10}
}

func (u *Updater) UpdateToday() error {
	logSection("STEP 1: Fetching Stock List from East Money")
	stocks, err := fetchStockList()
	if err != nil {
		return fmt.Errorf("fetch stock list: %w", err)
	}
	logInfo("Fetched %d stocks", len(stocks))

	if err := u.db.ensureStockListTable(); err != nil {
		return fmt.Errorf("ensure stock_list table: %w", err)
	}
	if err := u.db.replaceStockList(stocks); err != nil {
		return fmt.Errorf("save stock list: %w", err)
	}
	logOK("Stock list updated")

	logSection("STEP 2: Fetching Daily Trading Data from East Money")
	records, tradeDate, err := fetchDailyData()
	if err != nil {
		return fmt.Errorf("fetch daily data: %w", err)
	}
	logInfo("Fetched %d records for %s", len(records), tradeDate)

	if err := u.db.ensureDailyTable(tradeDate); err != nil {
		return fmt.Errorf("ensure daily table: %w", err)
	}
	if err := u.db.insertDailyData(tradeDate, records); err != nil {
		return fmt.Errorf("save daily data: %w", err)
	}
	logOK("Daily data saved to stock_daily_%s", tradeDate)

	logSection("STEP 3: Appending to Stock History")
	if err := u.db.ensureHistoryTable(); err != nil {
		return fmt.Errorf("ensure history table: %w", err)
	}

	var historyBatch []HistoryInsert
	for _, r := range records {
		code := stringsTrimPrefix(r.Symbol)
		historyBatch = append(historyBatch, HistoryInsert{
			StockCode: code,
			StockName: r.Name,
			Exchange:  getExchange(code),
			TradeDate: tradeDate,
			Open:      r.Open,
			High:      r.High,
			Low:       r.Low,
			Close:     r.Trade,
			Volume:    r.Volume,
			Amount:    float64(r.Amount),
		})
	}

	if err := u.db.insertHistoryBatch(historyBatch); err != nil {
		return fmt.Errorf("save history: %w", err)
	}
	logOK("Appended %d records to stock_history", len(historyBatch))

	return nil
}

func (u *Updater) UpdateRange(startDate, endDate string) error {
	logSection("STEP 1: Fetching Stock List")
	stocks, err := fetchStockList()
	if err != nil {
		return fmt.Errorf("fetch stock list: %w", err)
	}
	logInfo("Fetched %d stocks", len(stocks))

	if err := u.db.ensureStockListTable(); err != nil {
		return fmt.Errorf("ensure stock_list table: %w", err)
	}
	if err := u.db.replaceStockList(stocks); err != nil {
		return fmt.Errorf("save stock list: %w", err)
	}
	logOK("Stock list updated")

	today := time.Now().Format("20060102")
	if today >= startDate && today <= endDate {
		logSection("STEP 2: Fetching Today's Real-Time Data")
		records, tradeDate, err := fetchDailyData()
		if err != nil {
			return fmt.Errorf("fetch daily data: %w", err)
		}
		logInfo("Fetched %d records for %s", len(records), tradeDate)

		if err := u.db.ensureDailyTable(tradeDate); err != nil {
			return fmt.Errorf("ensure daily table: %w", err)
		}
		if err := u.db.insertDailyData(tradeDate, records); err != nil {
			return fmt.Errorf("save daily data: %w", err)
		}
		logOK("Daily data saved to stock_daily_%s", tradeDate)

		var historyBatch []HistoryInsert
		for _, r := range records {
			code := stringsTrimPrefix(r.Symbol)
			historyBatch = append(historyBatch, HistoryInsert{
				StockCode: code,
				StockName: r.Name,
				Exchange:  getExchange(code),
				TradeDate: tradeDate,
				Open:      r.Open,
				High:      r.High,
				Low:       r.Low,
				Close:     r.Trade,
				Volume:    r.Volume,
				Amount:    float64(r.Amount),
			})
		}
		if err := u.db.insertHistoryBatch(historyBatch); err != nil {
			return fmt.Errorf("save today history: %w", err)
		}
		logOK("Appended %d records to stock_history", len(historyBatch))

		if startDate == endDate && endDate == today {
			return nil
		}
	}

	return u.Backfill(startDate, endDate)
}

func (u *Updater) Backfill(startDate, endDate string) error {
	logSection("BACKFILL: Fetching missing historical data")
	logInfo("Date range: %s ~ %s", startDate, endDate)

	if err := u.db.ensureHistoryTable(); err != nil {
		return fmt.Errorf("ensure history table: %w", err)
	}

	stocks, err := u.db.getStockSymbols()
	if err != nil {
		return fmt.Errorf("get stocks: %w", err)
	}
	logInfo("Loaded %d stocks from database", len(stocks))

	existing, err := u.db.getExistingDates()
	if err != nil {
		return fmt.Errorf("get existing dates: %w", err)
	}
	logInfo("Existing records cover %d distinct trading dates", len(existing))

	type job struct {
		index  int
		symbol string
		name   string
	}

	type result struct {
		symbol string
		name   string
		count  int
		err    error
	}

	jobs := make(chan job, len(stocks))
	results := make(chan result, len(stocks))
	var wg sync.WaitGroup

	workerCount := u.worker
	if workerCount > len(stocks) {
		workerCount = len(stocks)
	}
	if workerCount < 1 {
		workerCount = 1
	}
	logInfo("Using %d workers", workerCount)

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				klines, err := fetchKLine(j.symbol, startDate, endDate)
				if err != nil || len(klines) == 0 {
					results <- result{symbol: j.symbol, name: j.name, count: 0, err: err}
					continue
				}

				var batch []HistoryInsert
				for _, kl := range klines {
					if existing[kl.Date] {
						continue
					}
					if !isTradingDay(kl.Date) {
						continue
					}
					batch = append(batch, HistoryInsert{
						StockCode: j.symbol,
						StockName: j.name,
						Exchange:  getExchange(j.symbol),
						TradeDate: kl.Date,
						Open:      kl.Open,
						High:      kl.High,
						Low:       kl.Low,
						Close:     kl.Close,
						Volume:    kl.Volume,
						Amount:    kl.Amount,
					})
				}

				if len(batch) > 0 {
					if err := u.db.insertHistoryBatch(batch); err != nil {
						results <- result{symbol: j.symbol, name: j.name, count: 0, err: err}
						continue
					}
				}
				results <- result{symbol: j.symbol, name: j.name, count: len(batch)}
				time.Sleep(50 * time.Millisecond)
			}
		}()
	}

	for i, s := range stocks {
		jobs <- job{index: i + 1, symbol: s.Symbol, name: s.Name}
	}
	close(jobs)

	wg.Wait()
	close(results)

	totalOK := 0
	totalInsert := 0
	totalErr := 0
	for r := range results {
		if r.err != nil {
			logWarn("[%s %s] error: %v", r.symbol, r.name, r.err)
			totalErr++
		} else {
			totalOK++
			totalInsert += r.count
			if r.count > 0 {
				logDebug("[%s %s] +%d records", r.symbol, r.name, r.count)
			}
		}
	}

	logSection("BACKFILL SUMMARY")
	logInfo("Stocks processed: %d OK, %d errors", totalOK, totalErr)
	logOK("New records inserted: %d", totalInsert)
	return nil
}

func (u *Updater) UpdateConcepts() error {
	logSection("UPDATE CONCEPTS: Fetching concept boards from Sina")

	if err := u.db.ensureConceptTable(); err != nil {
		return fmt.Errorf("ensure concept table: %w", err)
	}

	// Step 1: Refresh stock list
	logInfo("Step 1: Updating stock list...")
	stocks, err := fetchStockList()
	if err != nil {
		return fmt.Errorf("fetch stock list: %w", err)
	}
	if err := u.db.ensureStockListTable(); err != nil {
		return fmt.Errorf("ensure stock_list table: %w", err)
	}
	if err := u.db.replaceStockList(stocks); err != nil {
		return fmt.Errorf("save stock list: %w", err)
	}
	logOK("Stock list updated: %d stocks", len(stocks))

	// Step 2: Get all concept boards from Sina node tree
	boards, err := fetchSinaConceptBoards()
	if err != nil {
		return fmt.Errorf("fetch concept boards: %w", err)
	}
	logOK("Fetched %d concept boards from Sina", len(boards))

	// Step 3: Concurrently fetch stocks for each board
	type boardResult struct {
		board  SinaConceptBoard
		stocks []StockResult
	}

	jobs := make(chan SinaConceptBoard, len(boards))
	results := make(chan boardResult, len(boards))

	workerCount := u.worker
	if workerCount > len(boards) {
		workerCount = len(boards)
	}
	if workerCount < 1 {
		workerCount = 1
	}
	logInfo("Using %d workers", workerCount)

	var wg sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range jobs {
				stocks, err := fetchSinaBoardStocks(b.ID)
				if err != nil {
					logWarn("Board %s (%s): %v", b.ID, b.Name, err)
					results <- boardResult{board: b}
				} else {
					results <- boardResult{board: b, stocks: stocks}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	for _, b := range boards {
		jobs <- b
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Step 4: Build stock -> concepts mapping
	stockConcepts := make(map[string]map[string]bool)
	totalBoardOK := 0
	for r := range results {
		if len(r.stocks) > 0 {
			totalBoardOK++
			for _, s := range r.stocks {
				if stockConcepts[s.Symbol] == nil {
					stockConcepts[s.Symbol] = make(map[string]bool)
				}
				stockConcepts[s.Symbol][r.board.Name] = true
			}
		}
	}

	// Step 5: Convert to entries and save
	var allEntries []ConceptEntry
	for symbol, concepts := range stockConcepts {
		for c := range concepts {
			allEntries = append(allEntries, ConceptEntry{Symbol: symbol, Concept: c})
		}
	}

	if err := u.db.replaceConcepts(allEntries); err != nil {
		return fmt.Errorf("save concepts: %w", err)
	}

	// Step 6: Export JSON
	exportPath := filepath.Join(filepath.Dir(u.dbPath), "concept_index.json")
	u.exportConceptsJSON(exportPath)
	logInfo("Concept index exported to %s", exportPath)

	logSection("CONCEPT UPDATE SUMMARY")
	logInfo("Boards processed: %d / %d", totalBoardOK, len(boards))
	logInfo("Unique stocks with concepts: %d", len(stockConcepts))
	logOK("Total concept entries saved: %d", len(allEntries))
	return nil
}

func (u *Updater) exportConceptsJSON(path string) error {
	type stockInfo struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	type conceptIndex map[string][]stockInfo

	idx := make(conceptIndex)

	rows, err := u.db.conn.Query(`
		SELECT sc.concept, sc.symbol, COALESCE(sl.name, '')
		FROM stock_concepts sc
		LEFT JOIN stock_list sl ON sc.symbol = sl.symbol
		ORDER BY sc.concept, sc.symbol
	`)
	if err != nil {
		return fmt.Errorf("query concepts for export: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var concept, code, name string
		if err := rows.Scan(&concept, &code, &name); err != nil {
			continue
		}
		idx[concept] = append(idx[concept], stockInfo{Code: code, Name: name})
	}

	data, _ := json.MarshalIndent(idx, "", "  ")
	os.WriteFile(path, data, os.FileMode(0644))
	return nil
}

func stringsTrimPrefix(symbol string) string {
	for _, p := range []string{"sh", "sz", "bj"} {
		if len(symbol) > len(p) && symbol[:len(p)] == p {
			return symbol[len(p):]
		}
	}
	return symbol
}
