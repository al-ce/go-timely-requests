package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// scheduleDailyJob schedules a job to run once a day at midnight
func scheduleDailyJob(ctx context.Context, method, url string, ch chan<- string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		// Calculate time until next midnight
		waitTime := getTimeUntilNextMidnight()
		log.Printf("%s %s: next job scheduled in %v", method, url, waitTime)

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

			// Use context for the request so it can be canceled
			req = req.WithContext(ctx)

			resp, err := client.Do(req)
			if err != nil {
				ch <- fmt.Sprintf("Error executing request: %v", err)
				continue
			}

			ch <- fmt.Sprintf("%s %s %s", method, url, resp.Status)
			resp.Body.Close()

			// Check if context was canceled during request execution
			if ctx.Err() != nil {
				return
			}
		}
	}
}

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
