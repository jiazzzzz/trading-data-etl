package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LHBRecord struct {
	SecuCode     string  `json:"SECUCODE"`
	Name         string  `json:"SECURITY_NAME_ABBR"`
	Code         string  `json:"SECURITY_CODE"`
	TradeDate    string  `json:"TRADE_DATE"`
	ClosePrice   float64 `json:"CLOSE_PRICE"`
	ChangeRate   float64 `json:"CHANGE_RATE"`
	BuyTimes     int     `json:"BUY_TIMES"`
	SellTimes    int     `json:"SELL_TIMES"`
	BuyAmt       float64 `json:"BUY_AMT"`
	SellAmt      float64 `json:"SELL_AMT"`
	NetBuyAmt    float64 `json:"NET_BUY_AMT"`
	AccumAmt     float64 `json:"ACCUM_AMOUNT"`
	Ratio        float64 `json:"RATIO"`
	TurnoverRate float64 `json:"TURNOVERRATE"`
	FreeCap      float64 `json:"FREECAP"`
	Explanation  string  `json:"EXPLANATION"`
}

type LHBResp struct {
	Result *struct {
		Data  []LHBRecord `json:"data"`
		Count int         `json:"count"`
		Pages int         `json:"pages"`
	} `json:"result"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func main() {
	date := ""
	topN := 30
	maxCap := 200.0
	minCap := 0.0
	showAll := false
	skipRisk := false
	showHelp := false

	for i, arg := range os.Args[1:] {
		switch {
		case arg == "-date" && i+1 < len(os.Args[1:]):
			date = os.Args[i+2]
		case arg == "-top" && i+1 < len(os.Args[1:]):
			fmt.Sscanf(os.Args[i+2], "%d", &topN)
		case arg == "-maxcap" && i+1 < len(os.Args[1:]):
			fmt.Sscanf(os.Args[i+2], "%f", &maxCap)
		case arg == "-mincap" && i+1 < len(os.Args[1:]):
			fmt.Sscanf(os.Args[i+2], "%f", &minCap)
		case arg == "-all":
			showAll = true
		case arg == "-skiprisk":
			skipRisk = true
		case arg == "--help" || arg == "-h":
			showHelp = true
		}
	}

	if showHelp {
		printUsage()
		return
	}

	if date == "" {
		date = guessLatestTradeDay()
	}

	records := fetchLHB(date)
	if len(records) == 0 {
		fmt.Fprintf(os.Stderr, "未找到龙虎榜数据 (日期: %s)\n", date)
		os.Exit(1)
	}

	if !showAll {
		var filtered []LHBRecord
		for _, r := range records {
			if !skipRisk && isST(r.Name) {
				continue
			}
			if r.NetBuyAmt <= 0 {
				continue
			}
			if r.FreeCap > maxCap && maxCap > 0 {
				continue
			}
			if r.FreeCap < minCap {
				continue
			}
			filtered = append(filtered, r)
		}
		records = filtered
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].NetBuyAmt > records[j].NetBuyAmt
	})

	if topN > 0 && topN < len(records) {
		records = records[:topN]
	}

	printResult(records, date, showAll, skipRisk)
}

func fetchLHB(date string) []LHBRecord {
	total := 0
	allData := make([]LHBRecord, 0, 200)

	for page := 1; ; page++ {
		u, _ := url.Parse("https://datacenter-web.eastmoney.com/api/data/v1/get")
		q := url.Values{}
		q.Set("reportName", "RPT_ORGANIZATION_TRADE_DETAILS")
		q.Set("columns", "ALL")
		q.Set("pageNumber", fmt.Sprintf("%d", page))
		q.Set("pageSize", "100")
		q.Set("sortTypes", "-1")
		q.Set("sortColumns", "NET_BUY_AMT")
		q.Set("source", "WEB")
		q.Set("client", "WEB")
		q.Set("filter", fmt.Sprintf("(TRADE_DATE>='%s')", date))

		body, err := httpGet(u.String()+"?"+q.Encode(), "https://data.eastmoney.com/")
		if err != nil {
			fmt.Fprintf(os.Stderr, "API请求失败: %v\n", err)
			return nil
		}

		var resp LHBResp
		if err := json.Unmarshal(body, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
			return nil
		}

		if !resp.Success {
			fmt.Fprintf(os.Stderr, "API返回错误: %s\n", resp.Message)
			return nil
		}

		if resp.Result == nil || len(resp.Result.Data) == 0 {
			break
		}

		allData = append(allData, resp.Result.Data...)
		total = resp.Result.Count
		totalPages := resp.Result.Pages

		if page >= totalPages {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Fprintf(os.Stderr, "龙虎榜: %d 只上榜 (日期: %s)\n", total, date)
	return allData
}

func isST(name string) bool {
	return strings.HasPrefix(name, "ST") || strings.HasPrefix(name, "*ST") || strings.HasPrefix(name, "S") || strings.HasPrefix(name, "退")
}

func httpGet(urlStr, referer string) ([]byte, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func formatAmt(v float64) string {
	if v >= 1e8 {
		return fmt.Sprintf("%.2f亿", v/1e8)
	}
	if v >= 1e4 {
		return fmt.Sprintf("%.0f万", v/1e4)
	}
	return fmt.Sprintf("%.0f", v)
}

func formatCap(v float64) string {
	if v >= 1 {
		return fmt.Sprintf("%.0f亿", v)
	}
	return fmt.Sprintf("%.1f亿", v)
}

func printResult(records []LHBRecord, date string, showAll, skipRisk bool) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 72))
	fmt.Printf("  机构龙虎榜・净买入排行  %s\n", date)
	fmt.Println(strings.Repeat("=", 72))

	if len(records) == 0 {
		fmt.Println("\n  (无符合条件的记录)")
		fmt.Println()
		return
	}

	fmt.Printf("\n  %-4s %-8s %-10s %8s %12s %6s %6s %12s\n", "序号", "代码", "名称", "涨跌幅", "净买额", "机构买", "机构卖", "流通市值")
	fmt.Println("  " + strings.Repeat("-", 72))
	for i, r := range records {
		chgRate := r.ChangeRate
		if chgRate > 0 && chgRate < 1 {
			chgRate *= 100
		}
		chg := fmt.Sprintf("%+.2f%%", chgRate)
		fmt.Printf("  %-4d %-8s %-10s %8s %12s %6d %6d %12s\n",
			i+1, r.Code, r.Name, chg,
			formatAmt(r.NetBuyAmt), r.BuyTimes, r.SellTimes,
			formatCap(r.FreeCap))
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 72))
	fmt.Printf("  说明: 仅显示机构净买入 > 0 的股票")
	if !skipRisk {
		fmt.Printf(" (已排除ST/*ST/退市)")
	}
	fmt.Printf("\n  数据来源: 东方财富Choice")
	fmt.Println()
}

func guessLatestTradeDay() string {
	now := time.Now()
	weekday := now.Weekday()

	switch weekday {
	case time.Saturday:
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	case time.Sunday:
		return now.AddDate(0, 0, -2).Format("2006-01-02")
	default:
		if now.Hour() < 18 {
			return now.AddDate(0, 0, -1).Format("2006-01-02")
		}
		return now.Format("2006-01-02")
	}
}

func printUsage() {
	name := filepath.Base(os.Args[0])
	fmt.Printf(`%s - 龙虎榜机构买入查询

查询龙虎榜中机构席位买入数据，默认过滤ST/*ST并排除流通市值>200亿的大盘股。

Usage:
  %s [-date YYYY-MM-DD] [-top N] [-maxcap N] [-mincap N] [-all] [-skiprisk]

Options:
  -date YYYY-MM-DD    交易日 (默认:最近交易日)
  -top N              只显示前N只 (默认:30)
  -maxcap N           流通市值上限(亿), 0=不限制 (默认:200)
  -mincap N           流通市值下限(亿) (默认:0)
  -all                显示全部龙虎榜股票(不按净买入过滤)
  -skiprisk           不过滤ST/*ST/退市股票
  --help, -h          显示此帮助

Examples:
  %s                                  # 今日机构龙虎榜
  %s -date 2026-07-03                 # 指定日期
  %s -top 50 -maxcap 300              # 前50只,市值上限300亿
  %s -all -skiprisk                   # 全部龙虎榜(不过滤)
`,
		name, name, name, name, name, name)
}

func init() {
	_ = json.Marshal
}
