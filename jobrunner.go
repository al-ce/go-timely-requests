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

func scheduledRequest(ctx context.Context, method, url string, ch chan<- string, d time.Duration) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ch <- fmt.Sprintf("stopping request (%s %s)", method, url)
		case <-ticker.C:
			req, err := http.NewRequest(method, url, nil)
			if err != nil {
				ch <- fmt.Sprintf("Error creating request: %v", err)
				continue
			}

			resp, err := client.Do(req)
			defer resp.Body.Close()
			if err != nil {
				ch <- fmt.Sprintf("Error executing request: %v", err)
				continue
			}

			ch <- fmt.Sprintf("%s %s %s", method, url, resp.Status)
		}
	}
}

func main() {
	log.SetPrefix("[JOB]")

	ctx, cancel := context.WithCancel(context.Background())
	chStopSignal := make(chan os.Signal, 1)
	signal.Notify(chStopSignal, os.Interrupt)

	var wg sync.WaitGroup

	wg.Add(1)
	rotateTopicsChan := make(chan string)
	go scheduledRequest(
		ctx,
		"GET",
		"http://localhost:8080/topics/rotate",
		rotateTopicsChan,
		time.Second*5,
	)

	go func() {
		defer wg.Done()
		for {
			select {
			case rotate := <-rotateTopicsChan:
				log.Println(rotate)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-chStopSignal
	log.Println("Stopping jobrunner...")
	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("graceful shutdown")
	case <-time.After(3 * time.Second):
		log.Println("forcequitting")
	}
}
