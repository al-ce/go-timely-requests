package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// scheduleDailyJob schedules a job to run once a day
func scheduleDailyJob(ctx context.Context, job Job, ch chan<- string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		// Calculate time until next job run
		waitTime := getDurationToNextJobRun(job)
		log.Printf("%s %s: next job scheduled in %v", job.method, job.url, waitTime)

		// Timer will send to its channel at the next scheduled time
		timer := time.NewTimer(waitTime)

		select {
		case <-ctx.Done(): // Listen for context cancels after os.Interrupt signal
			timer.Stop()
			ch <- fmt.Sprintf("stopping request (%s %s)", job.method, job.url)
			return
		case <-timer.C: // Make the request at the scheduled time
			ch <- makeRequest(client, ctx, job)

			// Check if context was canceled during request execution
			if ctx.Err() != nil {
				return
			}
		}
	}
}

// getDurationToNextJobRun calculates the duration until the next scheduled job run
func getDurationToNextJobRun(job Job) time.Duration {
	now := time.Now().UTC()
	nextJobRun := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		job.hour,
		job.minute,
		job.second,
		0,
		time.UTC,
	)
	if now.After(nextJobRun) {
		// If we've already passed the schedlued time today, schedule for tomorrow
		nextJobRun = nextJobRun.Add(24 * time.Hour)
	}
	return nextJobRun.Sub(now)
}
