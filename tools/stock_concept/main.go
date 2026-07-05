package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type ConceptBoard struct {
	Code        string `json:"Code"`
	Name        string `json:"Name"`
	QuoteID     string `json:"QuoteID"`
	ConceptDesc string
}

type StockResult struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CacheEntry struct {
	Concept string        `json:"concept"`
	Board   ConceptBoard  `json:"board"`
	Stocks  []StockResult `json:"stocks"`
	Updated string        `json:"updated"`
}

type SinaItem struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

func main() {
	conceptName := ""
	dbPath := ""
	listMode := false
	workers := 30
	start := time.Now()

	for i, arg := range os.Args[1:] {
		if arg == "-name" && i+1 < len(os.Args[1:]) {
			conceptName = os.Args[i+2]
		}
		if arg == "-db" && i+1 < len(os.Args[1:]) {
			dbPath = os.Args[i+2]
		}
		if arg == "-workers" && i+1 < len(os.Args[1:]) {
			fmt.Sscanf(os.Args[i+2], "%d", &workers)
		}
		if arg == "-list" {
			listMode = true
		}
	}

	if hasFlag(os.Args[1:], "--help") || hasFlag(os.Args[1:], "-h") {
		printUsage()
		return
	}

	if listMode {
		listConcepts(dbPath)
		return
	}

	if conceptName == "" {
		printUsage()
		os.Exit(1)
	}

	// Resolve dbPath to absolute path for reliable directory derivation
	if dbPath != "" {
		abs, err := filepath.Abs(dbPath)
		if err == nil {
			dbPath = abs
		}
	}

	fmt.Fprintf(os.Stderr, "搜索概念: %s ...\n", conceptName)
	board := searchConcept(conceptName)

	if board.Code != "" {
		fmt.Fprintf(os.Stderr, "找到概念: %s (%s)\n", board.Name, board.Code)

		// Try JSON index first (more complete than DB)
		if dbPath != "" {
			idxPath := filepath.Join(filepath.Dir(dbPath), "concept_index.json")
			for _, key := range []string{board.Name, conceptName} {
				if stocks := readConceptIndex(idxPath, key); stocks != nil {
					updateBoardName(&board, key)
					fmt.Fprintf(os.Stderr, "从索引查询到结果 (%.1fs)\n", time.Since(start).Seconds())
					board.ConceptDesc = getConceptDesc(dbPath, board.Name)
					stocks = enrichDescriptions(dbPath, board.Name, stocks)
					printResult(board, stocks, "概念索引(concept_index.json)")
					return
				}
			}
		}

		// Try stock_concepts table
		if dbPath != "" {
			if stocks, ok := queryDB(dbPath, board.Name); ok {
				fmt.Fprintf(os.Stderr, "从数据库查询到结果 (%.1fs)\n", time.Since(start).Seconds())
				board.ConceptDesc = getConceptDesc(dbPath, board.Name)
				stocks = enrichDescriptions(dbPath, board.Name, stocks)
				printResult(board, stocks, "数据库(stock_concepts表)")
				return
			}
		}
	}

	// Try concept_stock_relation table directly (for concepts not in 东方财富/stock_concepts)
	if dbPath != "" {
		for _, key := range []string{conceptName, board.Name} {
			if key == "" {
				continue
			}
			if stocks, ok := queryRelationDB(dbPath, key); ok {
				fmt.Fprintf(os.Stderr, "从概念关联表查询到结果 (%.1fs, %d 只股票)\n", time.Since(start).Seconds(), len(stocks))
				board.ConceptDesc = getConceptDesc(dbPath, key)
				board.Name = key
				stocks = enrichDescriptions(dbPath, key, stocks)
				printResult(board, stocks, "数据库(concept_stock_relation表)")
				return
			}
		}
	}

	if board.Code == "" {
		fmt.Fprintf(os.Stderr, "未找到概念: %s\n", conceptName)
		os.Exit(1)
	}

	// Check JSON cache — try both names
	for _, key := range []string{board.Name, conceptName} {
		if cached := checkCache(key); cached != nil {
			updateBoardName(&board, key)
			fmt.Fprintf(os.Stderr, "使用缓存结果\n")
			board.ConceptDesc = getConceptDesc(dbPath, board.Name)
			cached = enrichDescriptions(dbPath, board.Name, cached)
			printResult(board, cached, "本地缓存")
			return
		}
	}

	// Fall back to scanning
	fmt.Fprintf(os.Stderr, "数据库无数据，从新浪获取全A股列表...\n")
	stockList := fetchAllStocks()
	if len(stockList) == 0 {
		fmt.Fprintf(os.Stderr, "获取股票列表失败\n")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "获取到 %d 只股票，开始扫描概念匹配 (并发 %d) ...\n", len(stockList), workers)

	scanKey := board.Name
	if !strings.HasPrefix(board.Code, "BK") {
		scanKey = conceptName
	}
	matches := scanConcept(scanKey, stockList, workers)
	board.ConceptDesc = getConceptDesc(dbPath, board.Name)
	matches = enrichDescriptions(dbPath, board.Name, matches)

	saveCache(board.Name, board, matches)
	printResult(board, matches, "爱股网实时扫描")
}

type ConceptIndex map[string][]StockResult

func queryDB(dbPath, concept string) ([]StockResult, bool) {
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	var tableExists int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='stock_concepts'").Scan(&tableExists)
	if tableExists == 0 {
		return nil, false
	}

	var rowCount int
	conn.QueryRow("SELECT COUNT(*) FROM stock_concepts WHERE concept = ?", concept).Scan(&rowCount)
	if rowCount == 0 {
		return nil, false
	}

	rows, err := conn.Query(`
		SELECT sc.symbol, COALESCE(sl.name, '')
		FROM stock_concepts sc
		LEFT JOIN stock_list sl ON sc.symbol = sl.symbol
		WHERE sc.concept = ?
		ORDER BY sc.symbol
	`, concept)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var stocks []StockResult
	for rows.Next() {
		var s StockResult
		if err := rows.Scan(&s.Code, &s.Name); err != nil {
			continue
		}
		stocks = append(stocks, s)
	}
	return stocks, true
}

func queryRelationDB(dbPath, concept string) ([]StockResult, bool) {
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	var tableExists int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='concept_stock_relation'").Scan(&tableExists)
	if tableExists == 0 {
		return nil, false
	}

	relName := findRelationConceptName(conn, concept)
	if relName == "" {
		return nil, false
	}

	rows, err := conn.Query("SELECT stock_code, stock_name FROM concept_stock_relation WHERE concept_name = ? ORDER BY stock_code", relName)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	var stocks []StockResult
	for rows.Next() {
		var s StockResult
		if err := rows.Scan(&s.Code, &s.Name); err != nil {
			continue
		}
		stocks = append(stocks, s)
	}
	if len(stocks) == 0 {
		return nil, false
	}
	fmt.Fprintf(os.Stderr, "概念关联表匹配: \"%s\" → \"%s\" (%d 只股票)\n", concept, relName, len(stocks))
	return stocks, true
}

func readConceptIndex(path, concept string) []StockResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var idx ConceptIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil
	}
	// Exact match first
	if stocks, ok := idx[concept]; ok {
		return stocks
	}
	// Partial match: concept is substring of a key, or vice versa
	for key, stocks := range idx {
		if strings.Contains(key, concept) || strings.Contains(concept, key) {
			fmt.Fprintf(os.Stderr, "索引模糊匹配: \"%s\" → \"%s\"\n", concept, key)
			return stocks
		}
	}
	return nil
}

func enrichDescriptions(dbPath, conceptName string, stocks []StockResult) []StockResult {
	if dbPath == "" || len(stocks) == 0 {
		return stocks
	}
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return stocks
	}
	defer conn.Close()

	var tableExists int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='concept_stock_relation'").Scan(&tableExists)
	if tableExists == 0 {
		return stocks
	}

	// Find the best matching concept_name in the relation table
	relName := findRelationConceptName(conn, conceptName)
	if relName == "" {
		return stocks
	}

	codeIdx := make(map[string]int, len(stocks))
	for i, s := range stocks {
		codeIdx[s.Code] = i
	}

	rows, err := conn.Query("SELECT stock_code, description FROM concept_stock_relation WHERE concept_name = ?", relName)
	if err != nil {
		return stocks
	}
	defer rows.Close()

	for rows.Next() {
		var code, desc string
		if err := rows.Scan(&code, &desc); err != nil {
			continue
		}
		if idx, ok := codeIdx[code]; ok && desc != "" {
			stocks[idx].Description = desc
		}
	}
	return stocks
}

func getConceptDesc(dbPath, conceptName string) string {
	if dbPath == "" {
		return ""
	}
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return ""
	}
	defer conn.Close()

	var tableExists int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='concept_stock_relation'").Scan(&tableExists)
	if tableExists == 0 {
		return ""
	}

	relName := findRelationConceptName(conn, conceptName)
	if relName == "" {
		return ""
	}

	var desc string
	err = conn.QueryRow("SELECT concept_desc FROM concept_stock_relation WHERE concept_name = ? LIMIT 1", relName).Scan(&desc)
	if err != nil || desc == "" {
		return ""
	}
	return desc
}

func findRelationConceptName(conn *sql.DB, conceptName string) string {
	// Try exact match first
	var name string
	err := conn.QueryRow("SELECT concept_name FROM concept_stock_relation WHERE concept_name = ? LIMIT 1", conceptName).Scan(&name)
	if err == nil {
		return name
	}

	// Try with "概念" suffix (relation table names often end with 概念)
	suffixed := conceptName + "概念"
	err = conn.QueryRow("SELECT concept_name FROM concept_stock_relation WHERE concept_name = ? LIMIT 1", suffixed).Scan(&name)
	if err == nil {
		return name
	}

	// Try substring match
	rows, err := conn.Query("SELECT DISTINCT concept_name FROM concept_stock_relation")
	if err != nil {
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var candidate string
		if err := rows.Scan(&candidate); err != nil {
			continue
		}
		if strings.Contains(candidate, conceptName) || strings.Contains(conceptName, candidate) {
			return candidate
		}
		// Try removing "概念" suffix
		trimmed := strings.TrimSuffix(candidate, "概念")
		if trimmed == conceptName {
			return candidate
		}
	}

	return ""
}

func searchConcept(name string) ConceptBoard {
	u := fmt.Sprintf("https://searchadapter.eastmoney.com/api/suggest/get?input=%s&type=14&token=43e6aa9dd7844fc053fb50678c1ef8cf",
		url.PathEscape(name))
	body := httpGet(u, "https://so.eastmoney.com/")
	if body == nil {
		return ConceptBoard{}
	}

	var resp struct {
		QuotationCodeTable *struct {
			Data []struct {
				Code    string `json:"Code"`
				Name    string `json:"Name"`
				QuoteID string `json:"QuoteID"`
			} `json:"Data"`
		} `json:"QuotationCodeTable"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return ConceptBoard{}
	}
	if resp.QuotationCodeTable == nil || len(resp.QuotationCodeTable.Data) == 0 {
		return ConceptBoard{}
	}
	d := resp.QuotationCodeTable.Data[0]
	return ConceptBoard{Code: d.Code, Name: d.Name, QuoteID: d.QuoteID}
}

var reConcepts = regexp.MustCompile(`<div class='tit'><span>要点一：</span><span>所属板块</span></div><p class='content1'>([^<]+)</p>`)

func scanConcept(concept string, stocks []StockResult, workers int) []StockResult {
	batchSize := 200
	var matches []StockResult
	total := len(stocks)

	for batchStart := 0; batchStart < total; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > total {
			batchEnd = total
		}
		batch := stocks[batchStart:batchEnd]

		var mu sync.Mutex
		var batchMatches []StockResult
		var wg sync.WaitGroup
		sema := make(chan struct{}, workers)

		for _, s := range batch {
			wg.Add(1)
			sema <- struct{}{}
			go func(stock StockResult) {
				defer wg.Done()
				defer func() { <-sema }()
				if stockHasConcept(stock.Code, concept) {
					mu.Lock()
					batchMatches = append(batchMatches, stock)
					mu.Unlock()
				}
			}(s)
		}
		wg.Wait()

		matches = append(matches, batchMatches...)
		fmt.Fprintf(os.Stderr, "  扫描: %d/%d (%.0f%%) 已找到 %d 只\n",
			batchEnd, total, float64(batchEnd)/float64(total)*100, len(matches))
	}
	return matches
}

func stockHasConcept(code, concept string) bool {
	u := fmt.Sprintf("https://igu888.com/ticai/%s.html", code)
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://igu888.com/")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	m := reConcepts.FindSubmatch(body)
	if len(m) < 2 {
		return false
	}
	return strings.Contains(string(m[1]), concept)
}

func fetchAllStocks() []StockResult {
	var all []StockResult
	for page := 1; ; page++ {
		u := fmt.Sprintf("http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=200&sort=code&asc=1&node=hs_a&symbol=&_s_r_a=init", page)
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Referer", "https://finance.sina.com.cn")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var items []SinaItem
		if err := json.Unmarshal(body, &items); err != nil || len(items) == 0 {
			break
		}
		for _, item := range items {
			if len(item.Symbol) < 3 {
				continue
			}
			prefix := item.Symbol[:2]
			if prefix != "sh" && prefix != "sz" && prefix != "bj" {
				continue
			}
			all = append(all, StockResult{Code: item.Symbol[2:], Name: item.Name})
		}
	}
	return all
}

func httpGet(urlStr, referer string) []byte {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func cachePath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, "concept_cache.json")
}

func checkCache(concept string) []StockResult {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil
	}
	var entries []CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	for _, e := range entries {
		if e.Concept == concept {
			return e.Stocks
		}
	}
	return nil
}

func saveCache(concept string, board ConceptBoard, stocks []StockResult) {
	entry := CacheEntry{
		Concept: concept,
		Board:   board,
		Stocks:  stocks,
		Updated: time.Now().Format("2006-01-02 15:04:05"),
	}
	var entries []CacheEntry
	existing, err := os.ReadFile(cachePath())
	if err == nil {
		json.Unmarshal(existing, &entries)
	}
	found := false
	for i, e := range entries {
		if e.Concept == concept {
			entries[i] = entry
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, entry)
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	os.WriteFile(cachePath(), data, os.FileMode(0644))
	fmt.Fprintf(os.Stderr, "结果已缓存到: %s\n", cachePath())
}

func listConcepts(dbPath string) {
	if dbPath == "" {
		fmt.Fprintf(os.Stderr, "请指定 -db 参数\n")
		os.Exit(1)
	}
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	var tableExists int
	conn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='stock_concepts'").Scan(&tableExists)
	if tableExists == 0 {
		fmt.Fprintf(os.Stderr, "数据库中没有 stock_concepts 表，请先运行 -update-concepts\n")
		os.Exit(1)
	}

	rows, err := conn.Query(`
		SELECT concept, COUNT(*) as cnt
		FROM stock_concepts
		GROUP BY concept
		ORDER BY cnt DESC, concept
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询失败: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	var total int
	fmt.Println(strings.Repeat("=", 64))
	fmt.Println("  概念列表 (按股票数降序)")
	fmt.Println(strings.Repeat("=", 64))
	fmt.Printf("\n  %-30s %s\n", "概念名称", "股票数")
	fmt.Println("  " + strings.Repeat("-", 42))
	for rows.Next() {
		var name string
		var cnt int
		rows.Scan(&name, &cnt)
		fmt.Printf("  %-30s %4d\n", name, cnt)
		total++
	}
	if total == 0 {
		fmt.Println("  (无数据)")
	}
	fmt.Println("  " + strings.Repeat("-", 42))
	fmt.Printf("  共 %d 个概念\n", total)
	fmt.Println()
}

func updateBoardName(board *ConceptBoard, name string) {
	if name != board.Name {
		board.Name = name
		board.Code = ""
	}
}

func printResult(board ConceptBoard, stocks []StockResult, source string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 58))
	fmt.Printf("  概念: %s (%s)\n", board.Name, board.Code)
	if board.ConceptDesc != "" {
		fmt.Printf("  说明: %s\n", board.ConceptDesc)
	}
	fmt.Println(strings.Repeat("=", 58))

	if len(stocks) == 0 {
		fmt.Println("\n  未找到相关股票")
	} else {
		fmt.Printf("\n  找到 %d 只相关股票:\n\n", len(stocks))
		for i, s := range stocks {
			exchange := "SZ"
			if strings.HasPrefix(s.Code, "6") {
				exchange = "SH"
			} else if strings.HasPrefix(s.Code, "8") || strings.HasPrefix(s.Code, "4") {
				exchange = "BJ"
			}
			fmt.Printf("  %3d. %-6s  %s  (%s)\n", i+1, s.Code, s.Name, exchange)
			if s.Description != "" {
				short := s.Description
				if len([]rune(short)) > 80 {
					short = string([]rune(short)[:80]) + "..."
				}
				fmt.Printf("      关联: %s\n", short)
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 58))
	fmt.Printf("  数据来源: %s\n", source)
	fmt.Println(strings.Repeat("=", 58))
}

func printUsage() {
	fmt.Println(`Stock Concept - 股票概念查询工具

Usage:
  concept.exe -name CONCEPT_NAME [-db DB_PATH]
  concept.exe -list -db DB_PATH

Required:
  -name CONCEPT_NAME  Concept name to search, e.g. "玻璃基板"

Options:
  -db DB_PATH         SQLite database path with stock_concepts table
                      If provided, queries DB first (fast).
                      If not found, scans via igu888.com (slow, ~5min)
  -list               List all concepts and stock counts (requires -db)
  -workers N          Parallel workers for scanning (default: 30)
  --help, -h          Show this help

Examples:
  concept.exe -name 白酒 -db ..\..\jia-stk.db
  concept.exe -list -db ..\..\jia-stk.db
  concept.exe -name OLED   (falls back to scanning)`)
}
