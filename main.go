package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

type stockValue struct {
	Symbol string  `json:"symbol"`
	Value  float64 `json:"value"`
}

type StockQuote struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	PercentChange float64 `json:"percent_change"`
}

func (i StockQuote) MarshalBinary() ([]byte, error) {
	return json.Marshal(i)
}

func fetchStockValue(symbol string, apiKey string) (*StockQuote, error) {
	url := fmt.Sprintf("https://apidojo-yahoo-finance-v1.p.rapidapi.com/market/v2/get-quotes?region=US&symbols=%s", strings.ToUpper(symbol))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-RapidAPI-Key", apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		errMsg := ""
		if errorResponse, ok := data["error"].(map[string]interface{}); ok {
			errMsg = errorResponse["message"].(string)
		} else {
			errMsg = "Unknown error"
		}
		return nil, errors.New(fmt.Sprintf("Error response from Yahoo Finance API: %s", errMsg))
	}
	result := data["quoteResponse"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})
	price, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", result["regularMarketPrice"].(float64)), 64)
	change, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", result["regularMarketChangePercent"].(float64)), 64)
	return &StockQuote{
		Symbol:        symbol,
		Price:         price,
		PercentChange: change,
	}, nil
}

func main() {
	apiKey := ""
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("REDIS_HOST environment variable not set")
	}

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
			value, err := fetchStockValue(symbol, apiKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = client.Set(symbol, value.Price, 0).Err()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			ttlDuration := 12 * time.Hour
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
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
