// jobrunner runs http requests daily at midnight

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


// Job represents an http request with the given method and url to be run daily
// at the given hour and second
type Job struct {
	method string
	url    string
	hour   int
	minute int
	second int
}

func main() {
	log.SetPrefix("[JOB] ")

	// Read list of jobs to schedule
	jobs, err := readJobs("jobs.tsv")
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
			scheduleDailyJob(ctx, job.method, job.url, resultChan)
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
