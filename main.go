package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type stockValue struct {
	Symbol string  `json:"symbol"`
	Value  float64 `json:"value"`
}

type StockQuote struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

func (i StockQuote) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func fetchStockValue(symbol string) (*StockQuote, error) {
	searchTerm := fmt.Sprintf("%s price", symbol)

	// Create a new context
	ctx, cancel := chromedp.NewContext(context.Background())

	// Uncomment for visual browser mode for debugging...
	//ctx, cancel := chromedp.NewExecAllocator(context.Background(), append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false))...)
	defer cancel()

	// Enable logging
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set up timeout
	ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Run the browser automation
	var stockPriceHTML string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.google.com/"),
		chromedp.WaitVisible(`textarea[aria-label="Search"]`, chromedp.ByQuery),
		chromedp.SetValue(`textarea[aria-label="Search"]`, searchTerm, chromedp.ByQuery),
		chromedp.Submit(`form[action="/search"]`, chromedp.ByQuery),
		chromedp.WaitVisible(`div[data-attrid="Price"]`, chromedp.ByQuery),
		chromedp.OuterHTML(`div[data-attrid="Price"]`, &stockPriceHTML, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Extract stock price using regex
	regex := regexp.MustCompile(`<span.*?>([0-9.]+)</span>`)
	match := regex.FindStringSubmatch(stockPriceHTML)

	var stockPrice float64
	if len(match) > 1 {
		stockPriceStr := match[1]
		stockPrice, err = strconv.ParseFloat(stockPriceStr, 64)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Stock price:", stockPriceStr)
	} else {
		fmt.Println("Stock price not found")
	}

	return &StockQuote{
		Symbol: symbol,
		Price:  stockPrice,
	}, nil
}

func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is not set", key)
	}
	return value
}

func main() {
	stockPrices := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stock_price",
			Help: "Latest stock price for the given symbol",
		},
		[]string{"symbol"},
	)
	prometheus.MustRegister(stockPrices)
	http.Handle("/metrics", promhttp.Handler())

	redisHost := getEnvVar("REDIS_HOST")

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	http.HandleFunc("/stock/", func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Path[len("/stock/"):]
		if symbol == "brew" {
			http.Error(w, "I'm a teapot", http.StatusTeapot)
			return
		}

		val, err := client.Get(symbol).Float64()
		if err == redis.Nil {
			value, err := fetchStockValue(symbol)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = client.Set(symbol, value.Price, 0).Err()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ttlDuration := 24 * time.Hour
			err = client.Expire(symbol, ttlDuration).Err()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			val = value.Price
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		stockValue := &stockValue{
			Symbol: symbol,
			Value:  val,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stockValue)
		stockPrices.WithLabelValues(symbol).Set(val)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
