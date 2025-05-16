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
)

func makeRequest(client *http.Client, ctx context.Context, job Job) string {
	req, err := http.NewRequest(job.method, job.url, strings.NewReader(job.data))
	if err != nil {
		return fmt.Sprintf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Use context for the request so it can be canceled
	req = req.WithContext(ctx)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error executing request: %v", err)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Error reading response body: %v", err)
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
	return fmt.Sprintf(
		"%s %s %s %v",
		job.method,
		job.url,
		resp.Status,
		respData.String(),
	)
}
