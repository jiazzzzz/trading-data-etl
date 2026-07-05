package main

type StockInfo struct {
	TsCode string `json:"ts_code"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Area   string `json:"area"`
	Industry string `json:"industry"`
	ListDate string `json:"list_date"`
	Pinyin string `json:"pinyin"`
}

type DailyRecord struct {
	Symbol        string  `json:"symbol"`
	Code          int64   `json:"code"`
	Name          string  `json:"name"`
	Trade         float64 `json:"trade"`
	PriceChange   float64 `json:"pricechange"`
	ChangePercent float64 `json:"changepercent"`
	Buy           float64 `json:"buy"`
	Sell          float64 `json:"sell"`
	Settlement    float64 `json:"settlement"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Volume        int64   `json:"volume"`
	Amount        int64   `json:"amount"`
	TickTime      string  `json:"ticktime"`
	Per           float64 `json:"per"`
	Pb            float64 `json:"pb"`
	Mktcap        float64 `json:"mktcap"`
	Nmc           float64 `json:"nmc"`
	TurnoverRatio float64 `json:"turnoverratio"`
	DumpTime      string  `json:"dump_time"`
}

type KLineData struct {
	Date   string
	Open   float64
	Close  float64
	High   float64
	Low    float64
	Volume int64
	Amount float64
}

// Sina concept board
type SinaConceptBoard struct {
	ID   string
	Name string
}
