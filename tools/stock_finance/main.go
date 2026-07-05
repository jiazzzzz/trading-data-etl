package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type FinRecord struct {
	ReportDate  string
	ReportType  string
	EPS         float64
	BPS         float64
	ROE         float64
	Revenue     float64
	NetProfit   float64
	RevYoY      float64
	ProfitYoY   float64
	GrossMargin float64
	NetMargin   float64
	DebtRatio   float64
	CurRatio    float64
	QuickRatio  float64
	OpCashFlow  float64
	TotalAssets float64
}

type RiskItem struct {
	Date   string
	Title  string
	Level  string
	Column string
}

var riskKeywords = []struct {
	word    string
	level   string
	display string
}{
	{"立案", "CRITICAL", "证监会立案调查"},
	{"强制退市", "CRITICAL", "强制退市"},
	{"终止上市", "CRITICAL", "终止上市"},
	{"行政处罚", "CRITICAL", "行政处罚"},
	{"公开谴责", "WARNING", "公开谴责"},
	{"通报批评", "WARNING", "通报批评"},
	{"监管警示", "WARNING", "监管警示"},
	{"监管措施", "WARNING", "监管措施"},
	{"纪律处分", "WARNING", "纪律处分"},
	{"罚款", "WARNING", "罚款"},
	{"出具警示函", "WARNING", "警示函"},
	{"停牌", "NOTICE", "停牌"},
	{"退市风险", "WARNING", "退市风险"},
	{"风险提示", "NOTICE", "风险提示"},
	{"整改", "NOTICE", "整改"},
	{"调查", "NOTICE", "调查"},
	{"资金占用", "WARNING", "资金占用"},
	{"违规担保", "WARNING", "违规担保"},
	{"财务造假", "CRITICAL", "财务造假"},
	{"虚假记载", "CRITICAL", "虚假记载"},
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || hasFlag(args, "--help") || hasFlag(args, "-h") {
		printUsage()
		return
	}

	code := extractFlag(args, "-code", "")
	if code == "" {
		fmt.Fprintln(os.Stderr, "ERROR: -code is required")
		os.Exit(1)
	}

	code = cleanCode(code)
	if code == "" {
		fmt.Fprintln(os.Stderr, "ERROR: invalid stock code")
		os.Exit(1)
	}

	showRisk := hasFlag(args, "-risk")

	topStr := extractFlag(args, "-top", "4")
	top := 4
	fmt.Sscanf(topStr, "%d", &top)
	if top <= 0 || top > 20 {
		top = 4
	}

	records, err := fetchFinance(code, top)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if len(records) > 0 {
		exchange := getExchange(code)
		name := records[0].ReportType
		displayName := name
		if name != "" {
			displayName = name
		}
		fmt.Printf("\n%s (%s.%s) - 财务数据\n\n", displayName, code, exchange)
		printFinanceTable(records)
	} else {
		fmt.Println("No financial data found for this stock.")
	}

	risks, err := fetchRiskInfo(code, showRisk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  [WARN] risk check failed: %v\n", err)
		return
	}

	if len(risks) > 0 {
		printRiskSummary(risks)
		if showRisk {
			fmt.Println()
			printRiskDetail(risks)
		}
	} else {
		fmt.Println("\n  [OK] 该股暂未发现近期风险预警公告")
	}
}

func cleanCode(code string) string {
	code = strings.ToUpper(code)
	code = strings.TrimPrefix(code, "SH")
	code = strings.TrimPrefix(code, "SZ")
	code = strings.TrimPrefix(code, "BJ")
	code = strings.TrimSuffix(code, ".SH")
	code = strings.TrimSuffix(code, ".SZ")
	code = strings.TrimSuffix(code, ".BJ")

	match, _ := regexp.MatchString(`^\d{6}$`, code)
	if !match {
		return ""
	}
	return code
}

func getExchange(code string) string {
	if strings.HasPrefix(code, "6") {
		return "SH"
	} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
		return "SZ"
	} else if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		return "BJ"
	}
	return "SH"
}

func fetchFinance(code string, limit int) ([]FinRecord, error) {
	u, _ := url.Parse("https://datacenter-web.eastmoney.com/api/data/v1/get")
	q := url.Values{}
	q.Set("reportName", "RPT_F10_FINANCE_MAINFINADATA")
	q.Set("columns", "ALL")
	q.Set("filter", fmt.Sprintf(`(SECURITY_CODE="%s")`, code))
	q.Set("pageNumber", "1")
	q.Set("pageSize", fmt.Sprintf("%d", limit+4))
	q.Set("sortTypes", "-1")
	q.Set("sortColumns", "NOTICE_DATE")

	body, err := httpGet(u.String()+"?"+q.Encode(),
		"https://emweb.securities.eastmoney.com/")
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool `json:"success"`
		Result  *struct {
			Data  []map[string]any `json:"data"`
			Count int              `json:"count"`
		} `json:"result"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	if !result.Success || result.Result == nil {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	var records []FinRecord
	for _, d := range result.Result.Data {
		r := FinRecord{
			ReportDate:  getStr(d, "REPORT_DATE"),
			ReportType:  getStr(d, "SECURITY_NAME_ABBR"),
			EPS:         getFlt(d, "EPSJB"),
			BPS:         getFlt(d, "BPS"),
			ROE:         getFlt(d, "ROEJQ"),
			Revenue:     getFlt(d, "TOTALOPERATEREVE"),
			NetProfit:   getFlt(d, "PARENTNETPROFIT"),
			RevYoY:      getFlt(d, "TOTALOPERATEREVETZ"),
			ProfitYoY:   getFlt(d, "PARENTNETPROFITTZ"),
			GrossMargin: getFlt(d, "XSMLL"),
			NetMargin:   getFlt(d, "XSJLL"),
			DebtRatio:   getFlt(d, "ZCFZL"),
			CurRatio:    getFlt(d, "LD"),
			QuickRatio:  getFlt(d, "SD"),
			OpCashFlow:  getFlt(d, "MGJYXJJE"),
		}
		records = append(records, r)
	}

	if len(records) > limit {
		records = records[:limit]
	}

	return records, nil
}

func fetchRiskInfo(code string, fullList bool) ([]RiskItem, error) {
	pageSize := 5
	if fullList {
		pageSize = 30
	}

	u := fmt.Sprintf(
		"https://np-anotice-stock.eastmoney.com/api/security/ann?sr=-1&page_size=%d&page_index=1&ann_type=A&stock_list=%s&f_node=0&s_node=0",
		pageSize, code,
	)

	body, err := httpGet(u, "https://np-anotice-stock.eastmoney.com/")
	if err != nil {
		return nil, err
	}

	var result struct {
		Data *struct {
			List []struct {
				Title      string `json:"title_ch"`
				NoticeDate string `json:"notice_date"`
				Columns    []struct {
					Name string `json:"column_name"`
				} `json:"columns"`
			} `json:"list"`
		} `json:"data"`
		Success int `json:"success"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse ann JSON: %w", err)
	}

	if result.Success != 1 || result.Data == nil {
		return nil, nil
	}

	var risks []RiskItem
	for _, ann := range result.Data.List {
		columns := ""
		for _, c := range ann.Columns {
			if columns != "" {
				columns += "; "
			}
			columns += c.Name
		}

		for _, kw := range riskKeywords {
			if strings.Contains(ann.Title, kw.word) || strings.Contains(columns, kw.word) {
				risks = append(risks, RiskItem{
					Date:   ann.NoticeDate,
					Title:  ann.Title,
					Level:  kw.level,
					Column: kw.display,
				})
				break
			}
		}
	}

	return risks, nil
}

func httpGet(urlStr, referer string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}

func getStr(data map[string]any, key string) string {
	v, ok := data[key]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func getFlt(data map[string]any, key string) float64 {
	v, ok := data[key]
	if !ok || v == nil {
		return 0
	}
	f, ok := v.(float64)
	if ok {
		return f
	}
	return 0
}

func printFinanceTable(records []FinRecord) {
	sep := strings.Repeat("-", 100)
	fmt.Printf("%-12s %-8s %8s %8s %6s %14s %14s %7s %7s %7s %7s %7s %7s %9s\n",
		"报告期", "类型", "EPS", "BPS", "ROE%",
		"营收", "净利润",
		"营收增%", "利润增%",
		"毛利率%", "净利率%",
		"负债率%", "流动率", "经营CF")
	fmt.Println(sep)

	for _, r := range records {
		reportDate := fmtDate(r.ReportDate)
		reportType := fmtReportType(r.ReportDate, records)

		fmt.Printf("%-12s %-8s %8.2f %8.2f %5.2f%% %14s %14s %6.2f%% %6.2f%% %6.2f%% %6.2f%% %6.2f%% %7.2f %9s\n",
			reportDate, reportType,
			r.EPS, r.BPS, r.ROE,
			fmtF(r.Revenue), fmtF(r.NetProfit),
			r.RevYoY, r.ProfitYoY,
			r.GrossMargin, r.NetMargin,
			r.DebtRatio, r.CurRatio,
			fmtF(r.OpCashFlow),
		)
	}

	if len(records) > 0 {
		fmt.Println(sep)
		fmt.Println("说明：EPS=每股收益  BPS=每股净资产  ROE=净资产收益率  经营CF=每股经营现金流")
	}
}

func printRiskSummary(risks []RiskItem) {
	var critical, warning, notice int
	for _, r := range risks {
		switch r.Level {
		case "CRITICAL":
			critical++
		case "WARNING":
			warning++
		case "NOTICE":
			notice++
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("  风险预警")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("  该股存在 %d 条风险相关公告\n", len(risks))
	if critical > 0 {
		fmt.Printf("  !! 严重: %d项\n", critical)
	}
	if warning > 0 {
		fmt.Printf("  !! 警告: %d项\n", warning)
	}
	if notice > 0 {
		fmt.Printf("  -- 提醒: %d项\n", notice)
	}
	fmt.Println()
}

func printRiskDetail(risks []RiskItem) {
	sep := strings.Repeat("-", 80)
	fmt.Printf("%-12s %-14s %s\n", "公告日期", "风险类型", "公告标题")
	fmt.Println(sep)

	for _, r := range risks {
		levelTag := "  -"
		switch r.Level {
		case "CRITICAL":
			levelTag = "!!严重"
		case "WARNING":
			levelTag = "!!警告"
		case "NOTICE":
			levelTag = "--提醒"
		}

		date := r.Date
		if len(date) >= 10 {
			date = date[:10]
		}

		title := r.Title
		if len([]rune(title)) > 50 {
			title = string([]rune(title)[:47]) + "..."
		}

		fmt.Printf("%-12s %-14s %s\n", date, levelTag, title)
	}
}

func fmtDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func fmtReportType(date string, records []FinRecord) string {
	if len(date) < 7 {
		return ""
	}
	ym := date[:7]
	for _, r := range records {
		if len(r.ReportDate) >= 10 && r.ReportDate[:7] == ym {
			return reportTypeFromDate(r.ReportDate)
		}
	}

	m := date[5:7]
	switch m {
	case "03":
		return "一季报"
	case "06":
		return "中报"
	case "09":
		return "三季报"
	case "12":
		return "年报"
	}
	return ""
}

func reportTypeFromDate(date string) string {
	if len(date) < 10 {
		return ""
	}
	m := date[5:7]
	switch m {
	case "03":
		return "一季报"
	case "06":
		return "中报/半年报"
	case "09":
		return "三季报"
	case "12":
		return "年报"
	}
	return ""
}

func fmtF(v float64) string {
	if v >= 1e8 {
		return fmt.Sprintf("%.2f亿", v/1e8)
	} else if v >= 1e4 {
		return fmt.Sprintf("%.2f万", v/1e4)
	} else if v > 0 {
		return fmt.Sprintf("%.2f", v)
	}
	return "-"
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

func printUsage() {
	fmt.Println(`Stock Finance - 股票财务及风险预警查询工具

Usage:
  finance.exe -code STOCK_CODE [-top N] [-risk]

Required:
  -code STOCK_CODE  Stock code, e.g. 600519, 000001, sh600519

Options:
  -top N            Number of financial periods (default: 4, max: 20)
  -risk             Show detailed risk announcement list
  --help, -h        Show this help

Risk check runs automatically; use -risk for full detail.

Examples:
  finance.exe -code 600519
  finance.exe -code 600811 -risk
  finance.exe -code 000001 -top 8`)
}
