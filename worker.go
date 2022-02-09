package main

import (
  // "fmt"
  "net/http"
	"sync"
)

type Job struct {
  Pair [2]string
}

type Worker struct {
  Job chan Job
  WorkerQueue chan chan Job
  Waiter *sync.WaitGroup
  Client *http.Client
}

func InitWorker(workerQueue chan chan Job, wg *sync.WaitGroup, httpClient *http.Client) Worker {
	return Worker {
		Job: make(chan Job),
		WorkerQueue: workerQueue,
    Waiter: wg,
    Client: httpClient,
	}
}

func (s *Worker) Start() {
	go func() {
		for {
			s.WorkerQueue <- s.Job

			job := <-s.Job

      trades := GetRecentPairs(s.Client, job.Pair)
      allTradesLock.Lock()
      allTrades = append(allTrades, trades)
      symbolsMap[len(allTrades)-1] = job.Pair[0] + job.Pair[1]
      allTradesLock.Unlock()

			s.Waiter.Done()
		}
	}()
}
