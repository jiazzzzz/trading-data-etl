package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	sinaURL  = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=100&sort=code&asc=1&node=hs_a&symbol=&_s_r_a=init"
	kLineURL = "https://ifzq.gtimg.cn/appstock/app/fqkline/get?param=%s,day,%s,%s,241,qfq"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func getExchangeAndID(code string) (prefix string, exchID int) {
	if strings.HasPrefix(code, "6") {
		return "sh", 1
	} else if strings.HasPrefix(code, "0") || strings.HasPrefix(code, "3") {
		return "sz", 0
	} else if strings.HasPrefix(code, "8") || strings.HasPrefix(code, "4") || strings.HasPrefix(code, "9") {
		return "bj", 0
	}
	return "", -1
}

func setCommonHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
}

func fetchHTTP(url, referer string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	setCommonHeaders(req, referer)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func fetchJSON(url string, target interface{}) error {
	return fetchHTTP(url, "https://quote.eastmoney.com/", target)
}

func fetchSinaJSON(url string, target interface{}) error {
	return fetchHTTP(url, "https://finance.sina.com.cn", target)
}

type sinaItem struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

type sinaDailyItem struct {
	Symbol        string  `json:"symbol"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Trade         float64 `json:"trade,string"`
	PriceChange   float64 `json:"pricechange"`
	ChangePercent float64 `json:"changepercent"`
	Buy           float64 `json:"buy,string"`
	Sell          float64 `json:"sell,string"`
	Settlement    float64 `json:"settlement,string"`
	Open          float64 `json:"open,string"`
	High          float64 `json:"high,string"`
	Low           float64 `json:"low,string"`
	Volume        int64   `json:"volume"`
	Amount        int64   `json:"amount"`
	TickTime      string  `json:"ticktime"`
	Per           float64 `json:"per"`
	Pb            float64 `json:"pb"`
	Mktcap        float64 `json:"mktcap"`
	Nmc           float64 `json:"nmc"`
	TurnoverRatio float64 `json:"turnoverratio"`
}

func fetchStockList() ([]StockInfo, error) {
	var stocks []StockInfo
	page := 1
	for {
		url := fmt.Sprintf(sinaURL, page)
		var items []sinaItem
		if err := fetchSinaJSON(url, &items); err != nil {
			return nil, fmt.Errorf("sina page %d: %w", page, err)
		}
		if len(items) == 0 {
			break
		}
		for _, item := range items {
			sym := item.Code
			name := item.Name
			if sym == "" || name == "" {
				continue
			}
			var exch string
			if strings.HasPrefix(sym, "6") {
				exch = "SH"
			} else if strings.HasPrefix(sym, "0") || strings.HasPrefix(sym, "3") {
				exch = "SZ"
			} else if strings.HasPrefix(sym, "8") || strings.HasPrefix(sym, "4") || strings.HasPrefix(sym, "9") {
				exch = "BJ"
			} else {
				continue
			}
			stocks = append(stocks, StockInfo{
				TsCode: sym + "." + exch,
				Symbol: sym,
				Name:   name,
			})
		}
		page++
		if page > 200 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return stocks, nil
}

// fetchSinaConceptBoards parses the HQNodes tree and returns all concept boards.
func fetchSinaConceptBoards() ([]SinaConceptBoard, error) {
	url := "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodes"
	var raw []interface{}
	if err := fetchSinaJSON(url, &raw); err != nil {
		return nil, fmt.Errorf("fetch node tree: %w", err)
	}
	boards := extractConceptBoards(raw)
	// Deduplicate by ID
	seen := make(map[string]bool)
	var out []SinaConceptBoard
	for _, b := range boards {
		if !seen[b.ID] {
			seen[b.ID] = true
			out = append(out, b)
		}
	}
	return out, nil
}

var nodeIDRe = regexp.MustCompile(`^(chgn_\d+|gn_\w+)$`)

func extractConceptBoards(node []interface{}) []SinaConceptBoard {
	var boards []SinaConceptBoard
	for _, v := range node {
		switch arr := v.(type) {
		case []interface{}:
			if len(arr) >= 3 {
				if name, ok := arr[0].(string); ok {
					if id, ok := arr[2].(string); ok && nodeIDRe.MatchString(id) {
						boards = append(boards, SinaConceptBoard{ID: id, Name: name})
					}
				}
			}
			boards = append(boards, extractConceptBoards(arr)...)
		}
	}
	return boards
}

// fetchSinaBoardStocks returns stocks for a Sina concept board node.
func fetchSinaBoardStocks(nodeID string) ([]StockResult, error) {
	baseURL := "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=100&sort=code&asc=1&node=%s&symbol=&_s_r_a=init"
	var all []StockResult
	page := 1
	for {
		url := fmt.Sprintf(baseURL, page, nodeID)
		var items []sinaItem
		if err := fetchSinaJSON(url, &items); err != nil {
			return nil, fmt.Errorf("node %s page %d: %w", nodeID, page, err)
		}
		if len(items) == 0 {
			break
		}
		for _, item := range items {
			all = append(all, StockResult{Symbol: item.Code, Name: item.Name})
		}
		page++
		if page > 50 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return all, nil
}

func fetchDailyData() ([]DailyRecord, string, error) {
	var records []DailyRecord
	now := time.Now()
	tickTime := now.Format("15:04:05")
	dumpTime := now.Format("2006-01-02 15:04:05")
	tradeDate := now.Format("20060102")

	page := 1
	for {
		url := fmt.Sprintf(sinaURL, page)
		var items []sinaDailyItem
		if err := fetchSinaJSON(url, &items); err != nil {
			return nil, "", fmt.Errorf("page %d: %w", page, err)
		}
		if len(items) == 0 {
			break
		}
		if page == 1 {
			logInfo("Total stocks: %d", len(items)*10)
		}
		for _, item := range items {
			rec := DailyRecord{
				Symbol:        item.Symbol,
				Code:          parseInt64(item.Code),
				Name:          item.Name,
				Trade:         item.Trade,
				PriceChange:   item.PriceChange,
				ChangePercent: item.ChangePercent,
				Buy:           item.Buy,
				Sell:          item.Sell,
				Settlement:    item.Settlement,
				Open:          item.Open,
				High:          item.High,
				Low:           item.Low,
				Volume:        item.Volume,
				Amount:        item.Amount,
				TickTime:      tickTime,
				Per:           item.Per,
				Pb:            item.Pb,
				Mktcap:        item.Mktcap,
				Nmc:           item.Nmc,
				TurnoverRatio: item.TurnoverRatio,
				DumpTime:      dumpTime,
			}
			records = append(records, rec)
		}
		page++
		time.Sleep(200 * time.Millisecond)
	}
	return records, tradeDate, nil
}

func fetchKLine(stockCode string, start, end string) ([]KLineData, error) {
	prefix, _ := getExchangeAndID(stockCode)
	if prefix == "" {
		return nil, fmt.Errorf("unknown exchange for %s", stockCode)
	}
	sd := fmt.Sprintf("%s-%s-%s", start[:4], start[4:6], start[6:])
	ed := fmt.Sprintf("%s-%s-%s", end[:4], end[4:6], end[6:])
	symbol := prefix + stockCode
	url := fmt.Sprintf(kLineURL, symbol, sd, ed)

	body, err := fetchRaw(url, "https://quote.eastmoney.com/")
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if code, ok := result["code"].(float64); ok && code != 0 {
		return nil, nil
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	stockData, ok := data[symbol].(map[string]interface{})
	if !ok {
		return nil, nil
	}
	klines, ok := stockData["qfqday"].([]interface{})
	if !ok || len(klines) == 0 {
		return nil, nil
	}
	var out []KLineData
	for _, kl := range klines {
		parts, ok := kl.([]interface{})
		if !ok || len(parts) < 6 {
			continue
		}
		d := strings.ReplaceAll(fmt.Sprint(parts[0]), "-", "")
		out = append(out, KLineData{
			Date:   d,
			Open:   toFloat64(parts[1]),
			Close:  toFloat64(parts[2]),
			High:   toFloat64(parts[3]),
			Low:    toFloat64(parts[4]),
			Volume: int64(toFloat64(parts[5])),
		})
	}
	return out, nil
}

func fetchRaw(url, referer string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", referer)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		return parseFloat(val)
	default:
		return 0
	}
}
