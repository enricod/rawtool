package main

import (
	"fmt"
	"time"
)

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueueChanChan chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:                  id,
		WorkRequestChan:     make(chan WorkRequest),
		WorkerQueueChanChan: workerQueueChanChan,
		QuitChan:            make(chan bool)}

	return worker
}

// Worker kk
type Worker struct {
	ID                  int
	WorkRequestChan     chan WorkRequest
	WorkerQueueChanChan chan chan WorkRequest
	QuitChan            chan bool
}

// Start This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueueChanChan <- w.WorkRequestChan

			select {
			case workRequest := <-w.WorkRequestChan:
				// Receive a work request.
				fmt.Printf("  worker%d: Received work request, delaying for %f seconds\n", w.ID, workRequest.Delay.Seconds())

				time.Sleep(workRequest.Delay)
				fmt.Printf("    worker%d: Hello, %s!\n", w.ID, workRequest.SourceFileName)

			case <-w.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
