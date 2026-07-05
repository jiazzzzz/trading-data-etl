package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"golang.org/x/text/encoding/simplifiedchinese"
)

type DB struct {
	conn *sql.DB
}

func openDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	return &DB{conn: conn}, nil
}

func (db *DB) close() {
	db.conn.Close()
}

func (db *DB) ensureStockListTable() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS stock_list (
			ts_code TEXT,
			symbol TEXT,
			name TEXT,
			area TEXT,
			industry TEXT,
			list_date TEXT,
			pinyin TEXT
		)
	`)
	return err
}

func (db *DB) ensureDailyTable(date string) error {
	tableName := "stock_daily_" + date
	_, err := db.conn.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			symbol TEXT,
			code INTEGER,
			name TEXT,
			trade REAL,
			pricechange REAL,
			changepercent REAL,
			buy REAL,
			sell REAL,
			settlement REAL,
			open REAL,
			high REAL,
			low REAL,
			volume INTEGER,
			amount INTEGER,
			ticktime TEXT,
			per REAL,
			pb REAL,
			mktcap REAL,
			nmc REAL,
			turnoverratio REAL,
			dump_time TEXT
		)
	`, tableName))
	return err
}

func (db *DB) ensureHistoryTable() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS stock_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			stock_code TEXT NOT NULL,
			stock_name TEXT,
			exchange TEXT,
			trade_date TEXT NOT NULL,
			open REAL,
			high REAL,
			low REAL,
			close REAL,
			volume INTEGER,
			amount REAL,
			import_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(stock_code, trade_date)
		)
	`)
	if err != nil {
		return err
	}
	db.conn.Exec("CREATE INDEX IF NOT EXISTS idx_stock_code ON stock_history(stock_code)")
	db.conn.Exec("CREATE INDEX IF NOT EXISTS idx_trade_date ON stock_history(trade_date)")
	db.conn.Exec("CREATE INDEX IF NOT EXISTS idx_stock_date ON stock_history(stock_code, trade_date)")
	return nil
}

func (db *DB) replaceStockList(stocks []StockInfo) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM stock_list")
	stmt, err := tx.Prepare("INSERT INTO stock_list (ts_code, symbol, name, area, industry, list_date, pinyin) VALUES (?, ?, ?, '', '', '', '')")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range stocks {
		pinyin := getPinyin(s.Name)
		if _, err := stmt.Exec(s.TsCode, s.Symbol, s.Name, pinyin); err != nil {
			return fmt.Errorf("insert %s: %w", s.Symbol, err)
		}
	}
	return tx.Commit()
}

func (db *DB) insertDailyData(date string, records []DailyRecord) error {
	tableName := "stock_daily_" + date
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(fmt.Sprintf("DELETE FROM %s", tableName))
	stmt, err := tx.Prepare(fmt.Sprintf(`
		INSERT INTO %s (symbol, code, name, trade, pricechange, changepercent, buy, sell, settlement, open, high, low, volume, amount, ticktime, per, pb, mktcap, nmc, turnoverratio, dump_time)
		VALUES (?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.Exec(r.Symbol, r.Code, r.Name, r.Trade, r.PriceChange, r.ChangePercent,
			r.Settlement, r.Open, r.High, r.Low, r.Volume, r.Amount, r.TickTime,
			r.Per, r.Pb, r.Mktcap, r.Nmc, r.TurnoverRatio, r.DumpTime); err != nil {
			return fmt.Errorf("insert %s: %w", r.Symbol, err)
		}
	}
	return tx.Commit()
}

func (db *DB) insertHistoryBatch(records []HistoryInsert) error {
	if len(records) == 0 {
		return nil
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO stock_history
		(stock_code, stock_name, exchange, trade_date, open, high, low, close, volume, amount)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.Exec(r.StockCode, r.StockName, r.Exchange, r.TradeDate,
			r.Open, r.High, r.Low, r.Close, r.Volume, r.Amount); err != nil {
			return fmt.Errorf("insert history %s %s: %w", r.StockCode, r.TradeDate, err)
		}
	}
	return tx.Commit()
}

type HistoryInsert struct {
	StockCode string
	StockName string
	Exchange  string
	TradeDate string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Amount    float64
}

func (db *DB) ensureConceptTable() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS stock_concepts (
			symbol TEXT NOT NULL,
			concept TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (symbol, concept)
		)
	`)
	if err != nil {
		return err
	}
	db.conn.Exec("CREATE INDEX IF NOT EXISTS idx_concept ON stock_concepts(concept)")
	return nil
}

func (db *DB) replaceConcepts(entries []ConceptEntry) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	seen := make(map[string]bool)
	for _, e := range entries {
		if !seen[e.Symbol] {
			seen[e.Symbol] = true
			if _, err := tx.Exec("DELETE FROM stock_concepts WHERE symbol = ?", e.Symbol); err != nil {
				return fmt.Errorf("delete concepts for %s: %w", e.Symbol, err)
			}
		}
	}

	stmt, err := tx.Prepare("INSERT INTO stock_concepts (symbol, concept, updated_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	for _, e := range entries {
		if _, err := stmt.Exec(e.Symbol, e.Concept, now); err != nil {
			return fmt.Errorf("insert concept %s %s: %w", e.Symbol, e.Concept, err)
		}
	}
	return tx.Commit()
}

func (db *DB) getConceptStocks(concept string) ([]StockResult, error) {
	rows, err := db.conn.Query(`
		SELECT sc.symbol, sl.name 
		FROM stock_concepts sc 
		LEFT JOIN stock_list sl ON sc.symbol = sl.symbol 
		WHERE sc.concept = ? 
		ORDER BY sc.symbol
	`, concept)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []StockResult
	for rows.Next() {
		var s StockResult
		if err := rows.Scan(&s.Symbol, &s.Name); err != nil {
			continue
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (db *DB) getConceptsCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(DISTINCT symbol) FROM stock_concepts").Scan(&count)
	return count, err
}

type ConceptEntry struct {
	Symbol  string
	Concept string
}

type StockResult struct {
	Symbol string
	Name   string
}

func (db *DB) getStockSymbols() ([]StockInfo, error) {
	rows, err := db.conn.Query("SELECT symbol, name FROM stock_list ORDER BY symbol")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []StockInfo
	for rows.Next() {
		var s StockInfo
		if err := rows.Scan(&s.Symbol, &s.Name); err != nil {
			continue
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (db *DB) getExistingDates() (map[string]bool, error) {
	rows, err := db.conn.Query("SELECT DISTINCT trade_date FROM stock_history")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dates := make(map[string]bool)
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates[d] = true
	}
	return dates, nil
}

func (db *DB) hasDailyTable(date string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", "stock_daily_"+date).Scan(&count)
	return count > 0, err
}

func (db *DB) tableExists(name string) (bool, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&count)
	return count > 0, err
}

func getExchange(code string) string {
	if strings.HasPrefix(code, "6") {
		return "SH"
	} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
		return "SZ"
	} else if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") || strings.HasPrefix(code, "9") {
		return "BJ"
	}
	return "UNKNOWN"
}

func isTradingDay(dateStr string) bool {
	t, err := time.Parse("20060102", dateStr)
	if err != nil {
		return false
	}
	w := t.Weekday()
	return w >= time.Monday && w <= time.Friday
}

func getPinyin(name string) string {
	if name == "" {
		return ""
	}
	enc := simplifiedchinese.GBK.NewEncoder()
	gbkBytes, err := enc.Bytes([]byte(name))
	if err != nil {
		return ""
	}

	py := strings.Builder{}
	py.Grow(len(name))

	i := 0
	for i < len(gbkBytes) {
		if gbkBytes[i] < 0x80 {
			c := gbkBytes[i]
			if c >= 'a' && c <= 'z' {
				py.WriteByte(c)
			} else if c >= 'A' && c <= 'Z' {
				py.WriteByte(c + 32)
			}
			i++
			continue
		}

		if i+1 >= len(gbkBytes) {
			break
		}

		asc := int(gbkBytes[i])*256 + int(gbkBytes[i+1]) - 65536
		i += 2

		p := pinyinFromAsc(asc)
		if p != 0 {
			py.WriteByte(p)
		}
	}
	return py.String()
}

func pinyinFromAsc(asc int) byte {
	switch {
	case asc >= -20319 && asc <= -20284:
		return 'a'
	case asc >= -20283 && asc <= -19776:
		return 'b'
	case asc >= -19775 && asc <= -19219:
		return 'c'
	case asc >= -19218 && asc <= -18711:
		return 'd'
	case asc >= -18710 && asc <= -18527:
		return 'e'
	case asc >= -18526 && asc <= -18240:
		return 'f'
	case asc >= -18239 && asc <= -17923:
		return 'g'
	case asc >= -17922 && asc <= -17418:
		return 'h'
	case asc >= -17417 && asc <= -16475:
		return 'j'
	case asc >= -16474 && asc <= -16213:
		return 'k'
	case asc >= -16212 && asc <= -15641:
		return 'l'
	case asc >= -15640 && asc <= -15166:
		return 'm'
	case asc >= -15165 && asc <= -14923:
		return 'n'
	case asc >= -14922 && asc <= -14915:
		return 'o'
	case asc >= -14914 && asc <= -14631:
		return 'p'
	case asc >= -14630 && asc <= -14150:
		return 'q'
	case asc >= -14149 && asc <= -14091:
		return 'r'
	case asc >= -14090 && asc <= -13119:
		return 's'
	case asc >= -13118 && asc <= -12839:
		return 't'
	case asc >= -12838 && asc <= -12557:
		return 'w'
	case asc >= -12556 && asc <= -11848:
		return 'x'
	case asc >= -11847 && asc <= -11056:
		return 'y'
	case asc >= -11055 && asc <= -10247:
		return 'z'
	}
	return 0
}



func parseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
