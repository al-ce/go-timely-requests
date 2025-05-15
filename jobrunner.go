// jobrunner runs http requests daily at midnight

// References:
// https://www.calhoun.io/using-select-in-go/
// https://www.calhoun.io/signals-via-context/

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

const GRACE_PERIOD = 10

// getTimeUntilNextMidnight calculates the time until the next midnight
func getTimeUntilNextMidnight() time.Duration {
	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if now.After(nextMidnight) {
		// If we've already passed midnight today, schedule for tomorrow
		nextMidnight = nextMidnight.Add(24 * time.Hour)
	}
	return nextMidnight.Sub(now)
}

// scheduleDailyJob schedules a job to run once a day at midnight
func scheduleDailyJob(ctx context.Context, method, url string, ch chan<- string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		// Calculate time until next midnight
		waitTime := getTimeUntilNextMidnight()
		log.Printf("Next job scheduled in %v", waitTime)

		// Timer will send to its channel on the next midnight
		timer := time.NewTimer(waitTime)

		select {

		case <-ctx.Done(): // Listen for context cancels after os.Interrupt signal
			timer.Stop()
			ch <- fmt.Sprintf("stopping request (%s %s)", method, url)
			return

		case <-timer.C: // Execute the job at midnight
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				ch <- fmt.Sprintf("Error creating request: %v", err)
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				ch <- fmt.Sprintf("Error executing request: %v", err)
				continue
			}

			ch <- fmt.Sprintf("%s %s %s", method, url, resp.Status)
			resp.Body.Close()
		}
	}
}

func main() {
	log.SetPrefix("[JOB] ")

	// Create cancellable context for clean shutdown
	ctx, cancel := context.WithCancel(context.Background())
	chStopSignal := make(chan os.Signal, 1)
	signal.Notify(chStopSignal, os.Interrupt)

	var wg sync.WaitGroup

	// Add topics/rotate job
	wg.Add(1)
	rotateTopicsChan := make(chan string)
	go scheduleDailyJob(
		ctx,
		"PUT",
		"http://localhost:8080/topics/rotate",
		rotateTopicsChan,
	)

	go func() {
		defer wg.Done()
		for {
			select {
			case rotate := <-rotateTopicsChan:
				log.Println(rotate)
			case <-ctx.Done():  // Stops loop after cancel()
				return
			}
		}
	}()

	// Block until signal interrupt
	<-chStopSignal
	log.Println("Stopping jobrunner...")
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
	case <-time.After(GRACE_PERIOD * time.Second):  // don't wait around forever!
		log.Println("forcequitting")
	}
}
