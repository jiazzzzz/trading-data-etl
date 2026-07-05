package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Test igu888 with 3 concurrent requests
	for i := 0; i < 5; i++ {
		code := "002876"
		if i%2 == 0 { code = "600519" }
		
		start := time.Now()
		u := "https://igu888.com/ticai/" + code + ".html"
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("Referer", "https://igu888.com/")
		
		resp, err := client.Do(req)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("[%d] %s ERR after %v: %v\n", i, code, elapsed, err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("[%d] %s OK after %v, body=%d\n", i, code, elapsed, len(body))
		}
	}
}
