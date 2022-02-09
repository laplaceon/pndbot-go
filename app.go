package main

import (
  "fmt"
  "net/http"
	"sync"
  "time"
)

const (
  tf int = 528
  delay time.Duration = 1 * time.Second
)

var (
  allTrades [][]Trade
  symbolsMap map[int]string
  allTradesLock *sync.Mutex
)

func InitWorkers(numWorkers int, jobQueue chan Job, wg *sync.WaitGroup, httpClient *http.Client) {
	workerQueue := make(chan chan Job, numWorkers)

	for i := 0; i < numWorkers; i++ {
		worker := InitWorker(workerQueue, wg, httpClient)
		worker.Start()
	}

  go func() {
    for {
      select {
      case job := <-jobQueue:
        go func() {
          worker := <-workerQueue
          worker <- job
        }()
      }
    }
  }()
}

func main() {
  fmt.Println("Started pndbot")

  httpClient := http.Client{}

  filteredSymbols := GetPairs(&httpClient, "ETH")
  fmt.Println("Retrieved trading pairs")

  jobQueue := make(chan Job, 1000)
	allTradesLock = &sync.Mutex{}
	wg := sync.WaitGroup{}

  InitWorkers(4, jobQueue, &wg, &httpClient)

  clf := InitClassifier()

  for {
    select {
			case <-time.After(delay):
        // Reset trades first
        allTrades = [][]Trade{}
        symbolsMap = make(map[int]string)

        for i := 0; i < len(filteredSymbols); i++ {
          jobQueue <- Job{filteredSymbols[i]}
          wg.Add(1)
          if i == 1 {
            break
          }
        }

        wg.Wait()
        fmt.Println("Fetched all recent trades for filtered symbols")

        clf.Predict(allTrades)
		}
  }
}
