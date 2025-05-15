package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// scheduleDailyJob schedules a job to run once a day at midnight
func scheduleDailyJob(ctx context.Context, job Job, ch chan<- string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		// Calculate time until next midnight
		waitTime := getDurationToNextJobRun(job)
		log.Printf("%s %s: next job scheduled in %v", job.method, job.url, waitTime)

		// Timer will send to its channel on the next midnight
		timer := time.NewTimer(waitTime)

		select {
		case <-ctx.Done(): // Listen for context cancels after os.Interrupt signal
			timer.Stop()
			ch <- fmt.Sprintf("stopping request (%s %s)", job.method, job.url)
			return
		case <-timer.C: // Make the request at the scheduled time
			req, err := http.NewRequest(job.method, job.url, strings.NewReader(job.data))
			if err != nil {
				ch <- fmt.Sprintf("Error creating request: %v", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			// Use context for the request so it can be canceled
			req = req.WithContext(ctx)

			// Make the request
			resp, err := client.Do(req)
			if err != nil {
				ch <- fmt.Sprintf("Error executing request: %v", err)
				continue
			}

			// Read response
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				continue
			}
			defer resp.Body.Close()

			// Format data to JSON
			var respData bytes.Buffer
			err = json.Indent(&respData, respBody, "", " ")
			if err != nil {
				log.Println(
					fmt.Sprintf(
						"Error formatting JSON: %s %s %s %v",
						job.method,
						job.url,
						job.data,
						err,
					),
				)
			}

			// Send response to channel for logging
			ch <- fmt.Sprintf(
				"%s %s %s %v",
				job.method,
				job.url,
				resp.Status,
				respData.String(),
			)

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
	nextMidnight := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		job.hour,
		job.minute,
		job.second,
		0,
		time.UTC,
	)
	if now.After(nextMidnight) {
		// If we've already passed midnight today, schedule for tomorrow
		nextMidnight = nextMidnight.Add(24 * time.Hour)
	}
	return nextMidnight.Sub(now)
}
