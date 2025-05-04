package main

import (
	"time"

	"github.com/Sparker0i/go-practice/jobqueue/internal/types"
)

func main() {
	dispatcher := types.NewDispatcher(2, 10, 2*time.Second)
	dispatcher.Run()

	for i := range 50 {
		dispatcher.JobQueue <- types.Job{
			ID: i,
		}
		time.Sleep(300 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	dispatcher.Stop()
}
