package main

import (
  "encoding/json"
  "fmt"
  "io/ioutil"
  "net/http"
  "strconv"
)

type Trade struct {
  Timestamp int
  Side int
  Price float64
  Amount float64
}

func GetPairs(httpClient *http.Client, filter string) [][2]string {
  res, err := httpClient.Get("https://api.binance.com/api/v3/exchangeInfo")

  if err != nil{
    panic(err)
  }

  var data map[string]interface{}

  body, err := ioutil.ReadAll(res.Body)

  if err := json.Unmarshal(body, &data); err != nil {
    panic(err)
  }

  symbols := data["symbols"].([]interface{})

  filteredSymbols := [][2]string{}

  for i := 0; i < len(symbols); i++ {
    symbol := symbols[i].(map[string]interface{})

    quoteAsset := symbol["quoteAsset"].(string)
    status := symbol["status"].(string)
    if (quoteAsset == filter) && (status == "TRADING") {
      baseAsset := symbol["baseAsset"].(string)
      filteredSymbols = append(filteredSymbols, [2]string{baseAsset, quoteAsset})
    }
  }

  return filteredSymbols
}

func GetRecentPairs(httpClient *http.Client, pair [2]string) []Trade {
  url := fmt.Sprintf("https://api.binance.com/api/v3/aggTrades?symbol=%s&limit=%d", pair[0] + pair[1], tf)

  res, err := httpClient.Get(url)

  if err != nil{
    panic(err)
  }

  var data []interface{}

  body, err := ioutil.ReadAll(res.Body)

  if err := json.Unmarshal(body, &data); err != nil {
    panic(err)
  }

  trades := make([]Trade, tf)

  for i := 0; i < len(data); i++ {
    item := data[i].(map[string]interface{})

    price, err := strconv.ParseFloat(item["p"].(string), 64)
    if err != nil{
      panic(err)
    }

    amount, err := strconv.ParseFloat(item["q"].(string), 64)
    if err != nil{
      panic(err)
    }

    side := 1
    if item["m"].(bool) {
      side = -1
    }

    trades[i] = Trade{
      Timestamp: int(item["T"].(float64)),
      Side: side,
      Price: price,
      Amount: amount,
    }
  }

  return trades
}
