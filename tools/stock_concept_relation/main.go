package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	_ "modernc.org/sqlite"
)

type ConceptEntry struct {
	ConceptName string
	ConceptDesc string
	Stocks      []StockEntry
}

type StockEntry struct {
	Code        string
	Name        string
	Description string
}

type ConceptLink struct {
	Slug string
	Name string
}

var (
	// Sidebar: <li><a href="./slug/">name</a></li>
	reSidebar = regexp.MustCompile(`<li><a href="\./([^/"]+)/"[^>]*>([^<]+)</a></li>`)

	// Detail page: stock table rows
	reDetailStock = regexp.MustCompile(`<tr><td class="gpxh">(\d+)</td><td><a[^>]*>(\d{6})</a></td><td[^>]*><a[^>]*>([^<]+)</a>`)

	// Detail page: info-box with description (includes stock name before colon)
	reInfoBox = regexp.MustCompile(`(?s)<tr><td class="info-box"><div class="info"><p>(.*?)</p></div></td></tr>`)
)

func main() {
	dbPath := ""
	updateMode := false
	limit := 0

	for i, arg := range os.Args[1:] {
		if arg == "-db" && i+1 < len(os.Args[1:]) {
			dbPath = os.Args[i+2]
		}
		if arg == "-update" {
			updateMode = true
		}
		if arg == "-limit" && i+1 < len(os.Args[1:]) {
			fmt.Sscanf(os.Args[i+2], "%d", &limit)
		}
	}

	if !updateMode || dbPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: concept_relation.exe -update -db DB_PATH [-limit N]\n")
		os.Exit(1)
	}

	if limit > 0 {
		fmt.Fprintf(os.Stderr, "限制爬取前 %d 个概念\n", limit)
	}

	if err := runUpdate(dbPath, limit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func decodeGBK(b []byte) string {
	decoder := simplifiedchinese.GBK.NewDecoder()
	out, err := decoder.Bytes(b)
	if err != nil {
		return string(b)
	}
	return string(out)
}

func fetchPage(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "http://ddx.gubit.cn/gainian/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return decodeGBK(body), nil
}

func parseSidebar(html string) []ConceptLink {
	seen := map[string]bool{}
	var links []ConceptLink
	matches := reSidebar.FindAllStringSubmatch(html, -1)
	for _, m := range matches {
		slug := strings.TrimSpace(m[1])
		name := strings.TrimSpace(m[2])
		if slug == "" || name == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		links = append(links, ConceptLink{Slug: slug, Name: name})
	}
	return links
}

func parseDetailPage(html string) ([]StockEntry, string) {
	var conceptDesc string

	// Extract concept description from h1 title (SEO description)
	reDesc := regexp.MustCompile(`(?s)<div class="gpintro">(.*?)</div>`)
	if m := reDesc.FindStringSubmatch(html); len(m) >= 2 {
		conceptDesc = strings.TrimSpace(m[1])
	}

	// Parse stock codes and names from the left table
	type stockInfo struct {
		Code string
		Name string
	}
	var stocks []stockInfo
	codeMatches := reDetailStock.FindAllStringSubmatch(html, -1)
	for _, m := range codeMatches {
		code := m[2]
		name := strings.TrimSpace(m[3])
		stocks = append(stocks, stockInfo{Code: code, Name: name})
	}

	if len(stocks) == 0 {
		return nil, conceptDesc
	}

	// Parse info-box descriptions from the right panel
	// Format: "StockName: description text"
	infoMatches := reInfoBox.FindAllStringSubmatch(html, -1)

	type descMatch struct {
		Name string
		Desc string
	}
	var descs []descMatch
	for _, m := range infoMatches {
		text := strings.TrimSpace(m[1])
		if text == "" {
			continue
		}
		// Split by first colon to get stock name + description
		// Format is like "均胜电子：2014年8月..."
		colonIdx := -1
		for i, r := range text {
			if r == ':' || r == '：' {
				colonIdx = i
				break
			}
		}
		if colonIdx > 0 {
			stockName := strings.TrimSpace(text[:colonIdx])
			desc := strings.TrimSpace(text[colonIdx+1:])
			if stockName != "" && desc != "" {
				descs = append(descs, descMatch{Name: stockName, Desc: desc})
			}
		}
	}

	// Match stock codes with descriptions by stock name
	nameToCode := make(map[string]string)
	for _, s := range stocks {
		nameToCode[s.Name] = s.Code
	}

	var result []StockEntry
	// First, add stocks that have descriptions
	usedCode := map[string]bool{}
	for _, d := range descs {
		if code, ok := nameToCode[d.Name]; ok {
			result = append(result, StockEntry{
				Code:        code,
				Name:        d.Name,
				Description: d.Desc,
			})
			usedCode[code] = true
		}
	}
	// Then add remaining stocks without descriptions
	for _, s := range stocks {
		if !usedCode[s.Code] {
			result = append(result, StockEntry{
				Code:        s.Code,
				Name:        s.Name,
				Description: "",
			})
		}
	}

	return result, conceptDesc
}

func initDB(dbPath string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	_, err = conn.Exec(`
		CREATE TABLE IF NOT EXISTS concept_stock_relation (
			concept_name TEXT NOT NULL,
			stock_code   TEXT NOT NULL,
			stock_name   TEXT DEFAULT '',
			description  TEXT DEFAULT '',
			concept_desc TEXT DEFAULT '',
			updated_date TEXT DEFAULT '',
			source       TEXT DEFAULT 'ddx',
			PRIMARY KEY (concept_name, stock_code)
		)
	`)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create table: %w", err)
	}

	return conn, nil
}

func saveEntries(db *sql.DB, entries []ConceptEntry, today string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Use INSERT OR REPLACE to handle concept-cache (still keep old ranking data if new detail page fails)
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO concept_stock_relation (concept_name, stock_code, stock_name, description, concept_desc, updated_date, source) VALUES (?, ?, ?, ?, ?, ?, 'ddx')")
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}

	for _, entry := range entries {
		for _, stock := range entry.Stocks {
			_, err := stmt.Exec(entry.ConceptName, stock.Code, stock.Name, stock.Description, entry.ConceptDesc, today)
			if err != nil {
				return fmt.Errorf("insert %s/%s: %w", entry.ConceptName, stock.Code, err)
			}
		}
	}

	return tx.Commit()
}

func runUpdate(dbPath string, limit int) error {
	today := time.Now().Format("2006-01-02")
	start := time.Now()
	fmt.Fprintf(os.Stderr, "概念关联度数据更新 %s\n", today)

	// Step 1: Fetch the ranking page to get sidebar with all concept slugs
	fmt.Fprintf(os.Stderr, "获取概念列表...\n")
	html, err := fetchPage("http://ddx.gubit.cn/gainian/px.php?zf=1")
	if err != nil {
		return fmt.Errorf("fetch concept list: %w", err)
	}

	allLinks := parseSidebar(html)
	fmt.Fprintf(os.Stderr, "  侧边栏获取到 %d 个概念\n", len(allLinks))

	if len(allLinks) == 0 {
		return fmt.Errorf("未获取到概念列表")
	}

	links := allLinks
	if limit > 0 && limit < len(links) {
		links = links[:limit]
		fmt.Fprintf(os.Stderr, "  限制爬取前 %d 个 (共 %d)\n", limit, len(allLinks))
	}

	// Step 2: Crawl each concept's detail page with rate limiting
	type conceptResult struct {
		Name   string
		Stocks []StockEntry
		Desc   string
		Err    error
	}
	results := make([]conceptResult, len(links))

	sema := make(chan struct{}, 3) // max 3 concurrent
	var wg sync.WaitGroup

	for i, link := range links {
		wg.Add(1)
		sema <- struct{}{}
		go func(idx int, cl ConceptLink) {
			defer wg.Done()
			defer func() { <-sema }()

			url := fmt.Sprintf("http://ddx.gubit.cn/gainian/%s/", cl.Slug)
			time.Sleep(100 * time.Millisecond) // rate limit

			pageHTML, err := fetchPage(url)
			if err != nil {
				results[idx] = conceptResult{Err: fmt.Errorf("fetch %s: %w", url, err)}
				return
			}

			stocks, conceptDesc := parseDetailPage(pageHTML)
			results[idx] = conceptResult{
				Name:   cl.Name,
				Stocks: stocks,
				Desc:   conceptDesc,
			}

			// Progress
			if idx > 0 && idx%50 == 0 {
				fmt.Fprintf(os.Stderr, "  进度: %d/%d (%.0f%%) %d stocks\n",
					idx, len(links), float64(idx)/float64(len(links))*100, len(stocks))
			}
		}(i, link)
	}
	wg.Wait()

	// Step 3: Collect results
	var allEntries []ConceptEntry
	successCount := 0
	failCount := 0
	totalStocks := 0

	for _, r := range results {
		if r.Err != nil {
			failCount++
			continue
		}
		if len(r.Stocks) == 0 {
			continue
		}
		successCount++
		totalStocks += len(r.Stocks)
		allEntries = append(allEntries, ConceptEntry{
			ConceptName: r.Name,
			ConceptDesc: r.Desc,
			Stocks:      r.Stocks,
		})
	}

	fmt.Fprintf(os.Stderr, "\n总计: %d 个概念成功, %d 个失败, %d 个股票-概念对 (耗时 %.1fs)\n",
		successCount, failCount, totalStocks, time.Since(start).Seconds())

	if len(allEntries) == 0 {
		return fmt.Errorf("未获取到任何有效概念数据")
	}

	db, err := initDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := saveEntries(db, allEntries, today); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "写入完成\n")
	return nil
}
