package types

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Worker struct {
	ID      int
	JobChan chan Job
	Quit    chan struct{}
}

func NewWorker(id int) *Worker {
	return &Worker{
		ID:      id,
		JobChan: make(chan Job),
		Quit:    make(chan struct{}),
	}
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case job := <-w.JobChan:
				fmt.Printf("Worker %d processing job %d\n", w.ID, job.ID)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
			case <-w.Quit:
				fmt.Printf("Stopping Worker ID %d\n", w.ID)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.Quit)
}
