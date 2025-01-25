package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sha65536/promstock/stock"
	"github.com/urfave/cli/v2"
)

var stockPriceGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "stock_prom_value",
		Help: "Current stock price with stock label",
	},
	[]string{"stock"},
)

func main() {
	prometheus.MustRegister(stockPriceGauge)

	app := &cli.App{
		Name:  "PromStock",
		Usage: "Expose stock prices as prometheus metrics",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "stocks",
				Usage:   "Comma-separated list of stock symbols (e.g., AAPL,GOOGL,MSFT)",
				EnvVars: []string{"STOCKS"},
			},
			&cli.StringFlag{
				Name:    "api-key",
				Usage:   "Finnhub API key",
				EnvVars: []string{"API_KEY"},
			},
			&cli.IntFlag{
				Name:    "interval",
				Usage:   "Interval in seconds between stock price updates",
				EnvVars: []string{"INTERVAL"},
				Value:   30,
			},
			&cli.IntFlag{
				Name:    "metrics-port",
				Usage:   "Port to expose Prometheus metrics",
				EnvVars: []string{"METRICS_PORT"},
				Value:   8080,
			},
		},
		Action: func(c *cli.Context) error {
			stocks := c.String("stocks")
			apiKey := c.String("api-key")
			interval := c.Int("interval")
			metricsPort := c.Int("metrics-port")

			if stocks == "" {
				return fmt.Errorf("you must provide at least one stock symbol using --stocks or STOCKS")
			}
			if apiKey == "" {
				return fmt.Errorf("you must provide an API key using --api-key or API_KEY")
			}
			if interval <= 0 {
				return fmt.Errorf("interval must be greater than 0")
			}

			stockSymbols := strings.Split(stocks, ",")

			go func() {
				http.Handle("/metrics", promhttp.Handler())
				log.Printf("Starting Prometheus metrics server on :%d\n", metricsPort)
				log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), nil))
			}()

			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			log.Printf("Tracking stocks: %v every %d seconds...\n", stockSymbols, interval)

			for range ticker.C {
				for _, stck := range stockSymbols {
					price, err := stock.FetchStockPrice(stck, apiKey)
					if err != nil {
						log.Printf("Error fetching price for %s: %v\n", stck, err)
						continue
					}
					stockPriceGauge.WithLabelValues(stck).Set(price)
					log.Printf("The current price of %s is $%.2f\n", stck, price)
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
