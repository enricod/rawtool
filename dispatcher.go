package main

import "fmt"

// WorkQueueWorkRequestChan A buffered channel that we can send work requests on.
var WorkQueueWorkRequestChan = make(chan WorkRequest, 100)

// WorkerQueueChanChan vedi http://tleyden.github.io/blog/2013/11/23/understanding-chan-chans-in-go/
// passiamo una richiesta che contiene un canale in nui il ricevente scriverà
var WorkerQueueChanChan chan chan WorkRequest

// StartDispatcher avviamo tutti i worker
func StartDispatcher(nworkers int) {

	WorkerQueueChanChan = make(chan chan WorkRequest, nworkers)

	// creiamo n worker queue
	for i := 0; i < nworkers; i++ {
		fmt.Println("Starting worker ", i+1)
		worker := NewWorker(i+1, WorkerQueueChanChan)
		worker.Start()
	}

	go func() {
		for {
			select {
			// WorkQueueWorkRequestChan è creato dal Collector
			case workReq := <-WorkQueueWorkRequestChan:
				// attende work request passate dal Collector
				fmt.Println("Received work request ", workReq.SourceFileName)
				go func() {
					worker := <-WorkerQueueChanChan

					fmt.Println("Dispatching work request ", workReq.SourceFileName)
					worker <- workReq
				}()
			}
		}
	}()
}
