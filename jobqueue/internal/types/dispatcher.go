package types

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Dispatcher struct {
	JobQueue   chan Job
	Workers    []*Worker
	MaxWorkers int
	MinWorkers int
	ScalerTick time.Duration
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func NewDispatcher(min, max int, scalerTick time.Duration) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{
		MinWorkers: min,
		MaxWorkers: max,
		JobQueue:   make(chan Job),
		Workers:    make([]*Worker, 0),
		ScalerTick: scalerTick,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (d *Dispatcher) Run() {
	for i := 0; i < d.MinWorkers; i++ {
		d.addWorker()
	}

	go d.scaler()

	go func() {
		for {
			select {
			case job := <-d.JobQueue:
				d.dispatch(job)
			case <-d.ctx.Done():
				return
			}
		}
	}()
}

func (d *Dispatcher) addWorker() {
	id := len(d.Workers) + 1
	worker := NewWorker(id)
	d.Workers = append(d.Workers, worker)
	worker.Start(&d.wg)
	fmt.Printf("Added Worker %d\n", worker.ID)
}

func (d *Dispatcher) removeWorker() {
	if len(d.Workers) == 0 {
		return
	}

	worker := d.Workers[len(d.Workers)-1]
	d.Workers = d.Workers[:len(d.Workers)-1]
	worker.Stop()
	fmt.Printf("Removed Worker %d\n", worker.ID)
}

func (d *Dispatcher) scaler() {
	ticker := time.NewTicker(d.ScalerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			queueLen := len(d.JobQueue)
			workerCount := len(d.Workers)

			if queueLen > workerCount && workerCount < d.MaxWorkers {
				d.addWorker()
			} else if queueLen == 0 && workerCount > d.MinWorkers {
				d.removeWorker()
			}

			d.mu.Unlock()
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) dispatch(job Job) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.Workers) == 0 {
		fmt.Println("No workers available")
		return
	}

	index := job.ID % len(d.Workers)
	d.Workers[index].JobChan <- job
}

func (d *Dispatcher) Stop() {
	d.cancel()
	d.mu.Lock()
	for _, w := range d.Workers {
		w.Stop()
	}
	d.mu.Unlock()
	d.wg.Wait()
}
