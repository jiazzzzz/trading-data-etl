package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type BasicInfo struct {
	Code      string
	Name      string
	Region    string
	Exchange  string
	MarketCap float64
	PE        float64
	ChangePct float64
	Price     float64
}

var reConcepts = regexp.MustCompile(`<div class='tit'><span>要点一：</span><span>所属板块</span></div><p class='content1'>([^<]+)</p>`)
var reNewsItem = regexp.MustCompile(`【(\d{4}-\d{2}-\d{2})】([^【]+)`)
var reAction = regexp.MustCompile(`【公告日期】(\d{4}-\d{2}-\d{2})【类别】([^【]+)(?:【简介】([^【]*))?`)

func main() {
	code := ""
	region := ""
	for i, arg := range os.Args[1:] {
		if arg == "-code" && i+1 < len(os.Args[1:]) {
			code = cleanCode(os.Args[i+2])
		}
	}
	for i, arg := range os.Args[1:] {
		if arg == "-region" && i+1 < len(os.Args[1:]) {
			region = os.Args[i+2]
		}
	}

	if code == "" || hasFlag(os.Args[1:], "--help") || hasFlag(os.Args[1:], "-h") {
		printUsage()
		if code == "" {
			os.Exit(1)
		}
		return
	}

	info := fetchBasicInfo(code)
	if info.Region == "" && region != "" {
		info.Region = region
	}
	concepts := fetchConcepts(code)
	news, actions := fetchNews(code)

	printReport(info, concepts, news, actions)
}

func cleanCode(s string) string {
	s = strings.ToUpper(s)
	s = strings.TrimPrefix(s, "SH")
	s = strings.TrimPrefix(s, "SZ")
	s = strings.TrimPrefix(s, "BJ")
	s = strings.TrimSuffix(s, ".SH")
	s = strings.TrimSuffix(s, ".SZ")
	s = strings.TrimSuffix(s, ".BJ")
	if matched, _ := regexp.MatchString(`^\d{6}$`, s); !matched {
		return ""
	}
	return s
}

func tencentPrefix(code string) string {
	switch {
	case strings.HasPrefix(code, "6"):
		return "sh"
	case strings.HasPrefix(code, "0"), strings.HasPrefix(code, "3"):
		return "sz"
	case strings.HasPrefix(code, "8"), strings.HasPrefix(code, "4"):
		return "bj"
	}
	return "sh"
}

func exchangeName(code string) string {
	switch {
	case strings.HasPrefix(code, "6"):
		return "SH"
	case strings.HasPrefix(code, "0"), strings.HasPrefix(code, "3"):
		return "SZ"
	case strings.HasPrefix(code, "8"), strings.HasPrefix(code, "4"):
		return "BJ"
	}
	return "SH"
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func fetchBasicInfo(code string) BasicInfo {
	// Try Tencent API first
	info := fetchFromTencent(code)
	if info.Name != "" {
		return info
	}
	// Fallback: push2 East Money API
	info = fetchFromPush2(code)
	if info.Name != "" {
		return info
	}
	return BasicInfo{Code: code, Name: code, Exchange: exchangeName(code)}
}

func fetchFromTencent(code string) BasicInfo {
	prefix := tencentPrefix(code)
	url := fmt.Sprintf("http://qt.gtimg.cn/q=%s%s", prefix, code)
	body := httpGetGB2312(url, "https://qt.gtimg.cn/")
	if body == nil {
		return BasicInfo{}
	}

	// Parse: v_{prefix}{code}="field1~field2~...";
	raw := string(body)
	eqIdx := strings.IndexByte(raw, '"')
	if eqIdx < 0 {
		return BasicInfo{}
	}
	endIdx := strings.LastIndexByte(raw, '"')
	if endIdx <= eqIdx {
		return BasicInfo{}
	}
	content := raw[eqIdx+1 : endIdx]
	fields := strings.Split(content, "~")
	if len(fields) < 46 {
		return BasicInfo{}
	}

	info := BasicInfo{
		Code:     fields[2],
		Name:     fields[1],
		Exchange: exchangeName(code),
	}

	// Field indices (0-based):
	// [3] = current price
	// [4] = yesterday close
	// [32] = change percent (e.g., "-5.19")
	// [39] = PE
	// [44] = 流通市值(亿)
	// [45] = 总市值(亿)

	if v, err := strconv.ParseFloat(fields[3], 64); err == nil {
		info.Price = v
	}
	if v, err := strconv.ParseFloat(fields[32], 64); err == nil {
		info.ChangePct = v
	}
	if v, err := strconv.ParseFloat(fields[39], 64); err == nil {
		info.PE = v
	}
	if v, err := strconv.ParseFloat(fields[45], 64); err == nil {
		info.MarketCap = v * 1e8 // fields are in 亿(yi)
	}

	return info
}

func fetchFromPush2(code string) BasicInfo {
	secID := "0." + code
	if strings.HasPrefix(code, "6") {
		secID = "1." + code
	} else if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") {
		secID = "2." + code
	}
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/stock/get?secid=%s&fields=f57,f58,f107,f128,f140,f170,f173", secID)
	body := httpGet(url, "https://quote.eastmoney.com/")
	if body == nil {
		return BasicInfo{}
	}

	var resp struct {
		Data *struct {
			Code     string  `json:"f57"`
			Name     string  `json:"f58"`
			Region   string  `json:"f128"`
			MktCap   float64 `json:"f140"`
			PE       float64 `json:"f173"`
			ChangeP  float64 `json:"f170"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil || resp.Data == nil {
		return BasicInfo{}
	}

	info := BasicInfo{
		Code:      resp.Data.Code,
		Name:      resp.Data.Name,
		Region:    resp.Data.Region,
		Exchange:  exchangeName(code),
		MarketCap: resp.Data.MktCap * 1e8,
		PE:        resp.Data.PE,
		ChangePct: resp.Data.ChangeP * 0.01,
	}
	return info
}

func fetchConcepts(code string) []string {
	url := fmt.Sprintf("https://igu888.com/ticai/%s.html", code)
	// Try raw bytes first (the page might already be UTF-8 despite charset=gb2312)
	body := httpGet(url, "https://igu888.com/")
	if body == nil {
		return nil
	}
	html := string(body)
	matches := reConcepts.FindStringSubmatch(html)
	if len(matches) >= 2 {
		return strings.Fields(matches[1])
	}
	// Fallback: try GB2312 decoding
	body = httpGetGB2312(url, "https://igu888.com/")
	if body == nil {
		return nil
	}
	html = string(body)
	matches = reConcepts.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil
	}
	return strings.Fields(matches[1])
}

type NewsItem struct {
	Date string
	Text string
}

type ActionItem struct {
	Date   string
	Type   string
	Detail string
}

func fetchNews(code string) ([]NewsItem, []ActionItem) {
	url := fmt.Sprintf("https://igu888.com/stocknews/%s.html", code)
	body := httpGet(url, "https://igu888.com/")
	if body == nil {
		return nil, nil
	}
	html := string(body)
	news := extractNews(html)
	actions := extractActions(html)
	if len(news) == 0 && len(actions) == 0 {
		// Try with GB2312 decoding
		body = httpGetGB2312(url, "https://igu888.com/")
		if body != nil {
			html = string(body)
			news = extractNews(html)
			actions = extractActions(html)
		}
	}
	return news, actions
}

func extractNews(html string) []NewsItem {
	matches := reNewsItem.FindAllStringSubmatch(html, -1)
	if len(matches) > 15 {
		matches = matches[:15]
	}
	var items []NewsItem
	seen := map[string]bool{}
	for _, m := range matches {
		key := m[1] + m[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		items = append(items, NewsItem{Date: m[1], Text: strings.TrimSpace(m[2])})
	}
	return items
}

func extractActions(html string) []ActionItem {
	matches := reAction.FindAllStringSubmatch(html, -1)
	var items []ActionItem
	seen := map[string]bool{}
	for _, m := range matches {
		key := m[1] + m[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		detail := ""
		if len(m) >= 4 {
			detail = strings.TrimSpace(m[3])
		}
		items = append(items, ActionItem{
			Date:   m[1],
			Type:   strings.TrimSpace(m[2]),
			Detail: detail,
		})
	}
	return items
}

func httpGet(urlStr, referer string) []byte {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

func httpGetGB2312(urlStr, referer string) []byte {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	decoder := simplifiedchinese.GBK.NewDecoder()
	reader := transform.NewReader(resp.Body, decoder)
	body, _ := io.ReadAll(reader)
	return body
}

func printReport(info BasicInfo, concepts []string, news []NewsItem, actions []ActionItem) {
	fmt.Println(strings.Repeat("=", 58))
	fmt.Printf("  %s (%s.%s) - 炒作点分析\n", info.Name, info.Code, info.Exchange)
	fmt.Println(strings.Repeat("=", 58))

	fmt.Println("\n  ▎基础信息")
	fmt.Println("  " + strings.Repeat("-", 54))
	var arrow string
	if info.ChangePct > 0 {
		arrow = "▲"
	} else if info.ChangePct < 0 {
		arrow = "▼"
	}
	if info.Price > 0 {
		changeStr := fmt.Sprintf("%.2f", info.ChangePct)
		fmt.Printf("  %-12s %.2f (%s%s%%)\n", "价格:", info.Price, arrow, changeStr)
	}
	if info.Region != "" {
		fmt.Printf("  %-12s %s\n", "板块:", info.Region)
	}
	if info.MarketCap > 0 {
		fmt.Printf("  %-12s %.2f亿\n", "总市值:", info.MarketCap/1e8)
	}
	if info.PE > 0 {
		fmt.Printf("  %-12s %.2f\n", "市盈率:", info.PE)
	}

	if len(concepts) > 0 {
		fmt.Println("\n  ▎概念题材 (" + fmt.Sprintf("%d项", len(concepts)) + ")")
		fmt.Println("  " + strings.Repeat("-", 54))
		const lineLen = 4
		for i := 0; i < len(concepts); i += lineLen {
			end := i + lineLen
			if end > len(concepts) {
				end = len(concepts)
			}
			fmt.Println("   " + strings.Join(concepts[i:end], "  |  "))
		}
	}

	if len(news) > 0 {
		fmt.Println("\n  ▎近期动态")
		fmt.Println("  " + strings.Repeat("-", 54))
		count := 0
		for _, n := range news {
			if count >= 8 {
				break
			}
			text := n.Text
			if len([]rune(text)) > 50 {
				text = string([]rune(text)[:47]) + "..."
			}
			fmt.Printf("  [%s] %s\n", n.Date, text)
			count++
		}
	}

	if len(actions) > 0 {
		fmt.Println("\n  ▎近期资本运作 (潜在炒作点)")
		fmt.Println("  " + strings.Repeat("-", 54))
		for i, a := range actions {
			if i >= 5 {
				break
			}
			fmt.Printf("  [%s] %s\n", a.Date, a.Type)
			if a.Detail != "" {
				detail := a.Detail
				if len([]rune(detail)) > 60 {
					detail = string([]rune(detail)[:57]) + "..."
				}
				fmt.Printf("        %s\n", detail)
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 58))
	fmt.Println("  数据来源: 腾讯(行情) + 爱股网(概念题材)")
	fmt.Println("  免责声明: 以上信息仅供参考，不构成投资建议")
	fmt.Println(strings.Repeat("=", 58))
}

func printUsage() {
	fmt.Println(`Stock Hype - 股票炒作点分析工具

Usage:
  hype.exe -code STOCK_CODE [-region REGION]

Required:
  -code STOCK_CODE  Stock code, e.g. 600519, 002876

Options:
  -region REGION    Override region name (if auto-detect fails)
  --help, -h        Show this help

Examples:
  hype.exe -code 002876
  hype.exe -code 600519 -region "贵州板块"`)
}
