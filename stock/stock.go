package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FinnhubResponse struct {
	CurrentPrice  float64 `json:"c"`
	HighPrice     float64 `json:"h"`
	LowPrice      float64 `json:"l"`
	OpenPrice     float64 `json:"o"`
	PreviousClose float64 `json:"pc"`
	Change        float64 `json:"d"`
	PercentChange float64 `json:"dp"`
	Timestamp     int64   `json:"t"`
}

var emptyRes FinnhubResponse

func FetchStockPrice(stockName, apiKey string) (float64, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", stockName, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error fetching stock data for %s: %v", stockName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received non-200 status code for %s: %d", stockName, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body for %s: %v", stockName, err)
	}

	var data FinnhubResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON for %s: %v", stockName, err)
	}

	if data == emptyRes {
		return 0, fmt.Errorf("got empty response for %s: %v", stockName, err)
	}

	return data.CurrentPrice, nil
}
