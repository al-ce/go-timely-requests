// Runs daily http requests at scheduled times

// References:
// https://www.calhoun.io/using-select-in-go/
// https://www.calhoun.io/signals-via-context/

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

const GRACE_PERIOD = 10

// Job represents an http request to be run at the scheduled time
type Job struct {
	id     int
	method string
	url    string
	hour   int
	minute int
	second int
	data   string
}

func main() {
	log.SetPrefix("[JOB] ")

	// Get file path from args
	if len(os.Args) != 2 {
		log.Fatalf("Usage: jobrunner [FILE]")
	}
	path := os.Args[1]

	// Read list of jobs to schedule
	jobs, err := readJobs(path)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	// Create cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	chStopSignal := make(chan os.Signal, 1)
	signal.Notify(chStopSignal, os.Interrupt)

	// Create a channel to receive job responses
	resultChan := make(chan string, len(jobs))

	var wg sync.WaitGroup
	// Start all jobs
	for _, j := range jobs {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			scheduleDailyJob(ctx, job, resultChan)
		}(j)
		log.Printf("scheduled: %v", j)
	}

	// Log results as they are sent to the result channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case result := <-resultChan:
				log.Println(result)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Block until signal interrupt
	<-chStopSignal
	log.Println("stopping jobrunner...")
	cancel()

	// Attempt to block program exit until wait group counter is zero
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("graceful shutdown")
	case <-time.After(GRACE_PERIOD * time.Second): // don't wait around forever!
		log.Println("forcequitting")
	}
}
