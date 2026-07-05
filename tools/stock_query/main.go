package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 || hasFlag(args, "--help") || hasFlag(args, "-h") {
		printUsage()
		return
	}

	dbPath := extractFlag(args, "-db", "")
	if dbPath == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -db is required")
		printUsage()
		os.Exit(1)
	}

	maMode := hasFlag(args, "-ma")

	if maMode {
		topStr := extractFlag(args, "-top", "50")
		toStr := extractFlag(args, "-to", "")
		maPeriodsStr := extractFlag(args, "-ma-periods", "5,10,20,60")

		limit := parseInt(topStr)
		if limit <= 0 {
			limit = 50
		}

		var queryTo string
		if toStr == "" {
			queryTo = time.Now().Format("20060102")
		} else {
			parsed, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid -to '%s'\n", toStr)
				os.Exit(1)
			}
			queryTo = parsed.Format("20060102")
		}

		if err := runMAQuery(dbPath, queryTo, maPeriodsStr, limit); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: query failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	gainStr := extractFlag(args, "-gain", "")
	if gainStr == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -gain is required")
		printUsage()
		os.Exit(1)
	}

	minGain := parseFloat(gainStr)
	if minGain == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid -gain value")
		os.Exit(1)
	}

	daysStr := extractFlag(args, "-days", "")
	topStr := extractFlag(args, "-top", "50")
	fromStr := extractFlag(args, "-from", "")
	toStr := extractFlag(args, "-to", "")

	limit := parseInt(topStr)
	if limit <= 0 {
		limit = 50
	}

	var queryFrom, queryTo string
	if daysStr != "" {
		n := parseInt(daysStr)
		if n <= 0 {
			n = 5
		}
		now := time.Now()
		from := now.AddDate(0, 0, -n)
		queryFrom = from.Format("20060102")
		queryTo = now.Format("20060102")
	} else {
		if fromStr == "" {
			queryFrom = time.Now().AddDate(0, 0, -5).Format("20060102")
		} else {
			parsed, err := time.Parse("2006-01-02", fromStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid -from '%s'\n", fromStr)
				os.Exit(1)
			}
			queryFrom = parsed.Format("20060102")
		}
		if toStr == "" {
			queryTo = time.Now().Format("20060102")
		} else {
			parsed, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid -to '%s'\n", toStr)
				os.Exit(1)
			}
			queryTo = parsed.Format("20060102")
		}
	}

	if queryFrom > queryTo {
		fmt.Fprintln(os.Stderr, "ERROR: -from must be before -to")
		os.Exit(1)
	}

	if err := runGainQuery(dbPath, queryFrom, queryTo, minGain, limit); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: query failed: %v\n", err)
		os.Exit(1)
	}
}

type GainResult struct {
	Code      string
	Name      string
	Pinyin    string
	GainPct   float64
	ClosePrev float64
	CloseCur  float64
	FirstDate string
	LastDate  string
}

func runGainQuery(dbPath, fromDate, toDate string, minGain float64, limit int) error {
	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.Query(`
		SELECT h.stock_code, h.stock_name, h.trade_date, h.close,
			   COALESCE(s.name, ''), COALESCE(s.pinyin, '')
		FROM stock_history h
		LEFT JOIN stock_list s ON h.stock_code = s.symbol
		WHERE h.trade_date >= ? AND h.trade_date <= ?
		ORDER BY h.stock_code, h.trade_date
	`, fromDate, toDate)
	if err != nil {
		return fmt.Errorf("query stock_history: %w", err)
	}
	defer rows.Close()

	type pricePoint struct {
		date  string
		close float64
	}
	stockData := make(map[string]struct {
		name     string
		fullName string
		pinyin   string
		prices   []pricePoint
	})

	for rows.Next() {
		var code, name, fullName, pinyin, date string
		var closeVal float64
		if err := rows.Scan(&code, &name, &date, &closeVal, &fullName, &pinyin); err != nil {
			continue
		}
		entry := stockData[code]
		entry.name = name
		entry.fullName = fullName
		entry.pinyin = pinyin
		entry.prices = append(entry.prices, pricePoint{date, closeVal})
		stockData[code] = entry
	}

	var results []GainResult
	for code, data := range stockData {
		if len(data.prices) < 2 {
			continue
		}
		first := data.prices[0]
		last := data.prices[len(data.prices)-1]
		if first.close == 0 {
			continue
		}
		gain := (last.close - first.close) / first.close * 100
		if minGain >= 0 {
			if gain < minGain {
				continue
			}
		} else {
			if gain > minGain {
				continue
			}
		}
		results = append(results, GainResult{
			Code:      code,
			Name:      data.fullName,
			Pinyin:    data.pinyin,
			GainPct:   gain,
			ClosePrev: first.close,
			CloseCur:  last.close,
			FirstDate: first.date,
			LastDate:  last.date,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].GainPct > results[j].GainPct
	})

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	printResults(results, minGain, fromDate, toDate)
	return nil
}

type MAResult struct {
	Code      string
	Name      string
	Pinyin    string
	MAs       []float64
	CloseCur  float64
	GainPct   float64
	TradeDate string
}

func runMAQuery(dbPath, toDate, periodsStr string, limit int) error {
	periods := parsePeriods(periodsStr)
	if len(periods) < 2 {
		return fmt.Errorf("至少需要2个均线周期")
	}
	maxPeriod := periods[len(periods)-1]

	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_loc=auto")
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.Query(`
		SELECT h.stock_code, h.stock_name, h.trade_date, h.close,
			   COALESCE(s.name, ''), COALESCE(s.pinyin, '')
		FROM stock_history h
		LEFT JOIN stock_list s ON h.stock_code = s.symbol
		WHERE h.trade_date <= ?
		ORDER BY h.stock_code, h.trade_date DESC
	`, toDate)
	if err != nil {
		return fmt.Errorf("query stock_history: %w", err)
	}
	defer rows.Close()

	type pricePoint struct {
		date  string
		close float64
	}
	stockData := make(map[string]struct {
		name     string
		fullName string
		pinyin   string
		prices   []pricePoint
	})

	for rows.Next() {
		var code, name, fullName, pinyin, date string
		var closeVal float64
		if err := rows.Scan(&code, &name, &date, &closeVal, &fullName, &pinyin); err != nil {
			continue
		}
		entry := stockData[code]
		entry.name = name
		entry.fullName = fullName
		entry.pinyin = pinyin
		entry.prices = append(entry.prices, pricePoint{date, closeVal})
		stockData[code] = entry
	}

	var results []MAResult
	for code, data := range stockData {
		prices := data.prices
		if len(prices) < maxPeriod {
			continue
		}

		closeVals := make([]float64, len(prices))
		for i, p := range prices {
			closeVals[i] = p.close
		}

		mas := make([]float64, len(periods))
		allAscending := true
		for i, p := range periods {
			ma := calcSMA(closeVals, p)
			mas[i] = ma
			if i > 0 && ma <= mas[i-1] {
				allAscending = false
			}
		}
		if !allAscending {
			continue
		}

		closeCur := closeVals[0]
		gain := 0.0
		if len(closeVals) >= periods[0]+5 {
			prev := closeVals[periods[0]]
			if prev > 0 {
				gain = (closeCur - prev) / prev * 100
			}
		}

		results = append(results, MAResult{
			Code:      code,
			Name:      data.fullName,
			Pinyin:    data.pinyin,
			MAs:       mas,
			CloseCur:  closeCur,
			GainPct:   gain,
			TradeDate: prices[0].date,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CloseCur/results[i].MAs[0] > results[j].CloseCur/results[j].MAs[0]
	})

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	printMAResults(results, periods, toDate)
	return nil
}

func parsePeriods(s string) []int {
	parts := strings.Split(s, ",")
	periods := make([]int, 0, len(parts))
	for _, p := range parts {
		var v int
		fmt.Sscanf(strings.TrimSpace(p), "%d", &v)
		if v > 0 {
			periods = append(periods, v)
		}
	}
	sort.Ints(periods)
	return periods
}

func calcSMA(closeVals []float64, period int) float64 {
	if len(closeVals) < period {
		return 0
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += closeVals[i]
	}
	return sum / float64(period)
}

func printMAResults(results []MAResult, periods []int, toDate string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("  均线多头排列查询")
	fmt.Println(strings.Repeat("=", 70))

	periodStrs := make([]string, len(periods))
	for i, p := range periods {
		periodStrs[i] = fmt.Sprintf("MA%d", p)
	}

	condStr := strings.Join(periodStrs, " > ")
	fmt.Printf("\n  条件: %s\n", condStr)
	fmt.Printf("  截止: %s\n\n", toDate)

	header := fmt.Sprintf("%-9s %-10s", "代码", "名称")
	for _, p := range periods {
		header += fmt.Sprintf(" %8s", fmt.Sprintf("MA%d", p))
	}
	header += fmt.Sprintf(" %8s %8s", "现价", "涨幅%")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)+5))

	for _, r := range results {
		line := fmt.Sprintf("%-9s %-10s", r.Code, trunc(r.Name, 10))
		for _, ma := range r.MAs {
			line += fmt.Sprintf(" %8.2f", ma)
		}
		sign := ""
		if r.GainPct > 0 {
			sign = "+"
		}
		line += fmt.Sprintf(" %8.2f %7s%.2f%%", r.CloseCur, sign, r.GainPct)
		fmt.Println(line)
	}

	fmt.Printf("\n  共 %d 只股票符合条件\n\n", len(results))
}

func printResults(results []GainResult, minGain float64, from, to string) {
	var cond string
	if minGain >= 0 {
		cond = fmt.Sprintf("涨幅 >= %.1f%%", minGain)
	} else {
		cond = fmt.Sprintf("跌幅 >= %.1f%%", -minGain)
	}
	fmt.Printf("\n查询结果：%s（%s ~ %s）\n\n", cond, from, to)
	if minGain >= 0 {
		fmt.Printf("%-9s %-10s %-8s %8s %8s %8s  %s\n",
			"代码", "名称", "拼音", "涨幅%", "前收盘", "现收盘", "日期区间")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range results {
			dates := fmt.Sprintf("%s~%s", r.FirstDate, r.LastDate)
			fmt.Printf("%-9s %-10s %-8s %7.2f%% %8.2f %8.2f  %s\n",
				r.Code, trunc(r.Name, 10), trunc(r.Pinyin, 8),
				r.GainPct, r.ClosePrev, r.CloseCur, dates)
		}
	} else {
		fmt.Printf("%-9s %-10s %-8s %8s %8s %8s  %s\n",
			"代码", "名称", "拼音", "跌幅%", "前收盘", "现收盘", "日期区间")
		fmt.Println(strings.Repeat("-", 80))
		for _, r := range results {
			dates := fmt.Sprintf("%s~%s", r.FirstDate, r.LastDate)
			fmt.Printf("%-9s %-10s %-8s %7.2f%% %8.2f %8.2f  %s\n",
				r.Code, trunc(r.Name, 10), trunc(r.Pinyin, 8),
				-r.GainPct, r.ClosePrev, r.CloseCur, dates)
		}
	}
	fmt.Printf("\n共 %d 只股票符合条件\n\n", len(results))
}

func trunc(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "…"
}

func printUsage() {
	fmt.Println(`Stock Query - 股票涨跌幅/均线多头排列查询工具

Usage:
  query.exe -db PATH -gain PCT [-days N] [-top N] [-from YYYY-MM-DD] [-to YYYY-MM-DD]
  query.exe -db PATH -ma [-ma-periods P1,P2,...] [-to YYYY-MM-DD] [-top N]

Modes:
  -gain PCT        涨跌幅模式: 查找区间涨幅 >= PCT% 的股票 (负值表示跌幅)
  -ma              均线多头排列模式: 查找短周期均线 > 长周期均线的股票

Required:
  -db PATH         Database path (SQLite, same schema as jia-stk.db)

Options:
  -days N          Look back N calendar days (default: 5, overrides -from)
  -top N           Max results to show (default: 50)
  -from YYYY-MM-DD Explicit start date (overridden by -days)
  -to YYYY-MM-DD   Explicit end date
  -ma-periods P    Comma-separated MA periods (default: 5,10,20,60)
  --help, -h       Show this help

Examples:
  query.exe -db jia-stk.db -gain 7 -days 5
  query.exe -db jia-stk.db -gain -5 -from 2026-06-01 -to 2026-07-04
  query.exe -db jia-stk.db -ma
  query.exe -db jia-stk.db -ma -ma-periods 10,20,60,120 -to 2026-07-01`)
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func extractFlag(args []string, flag, def string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return def
}

func parseInt(s string) int {
	var v int
	fmt.Sscanf(s, "%d", &v)
	if v <= 0 {
		return 5
	}
	return v
}

func parseFloat(s string) float64 {
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}
