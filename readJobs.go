package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// readJobs scans the lines of `file` and returns a slice of the job details.
// Each line should be tab separated in the following format:
// {method}\t{url}\t{hour}\t{minute}\t{second}
func readJobs(path string) ([]Job, error) {
	var jobs []Job
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return jobs, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")

		// Needs minimum 5 fields
		if len(fields) < 5 {
			return jobs, fmt.Errorf("insufficient fields: %s", line)
		}
		method, url := fields[0], fields[1]
		hour, err := strconv.Atoi(fields[2])
		if err != nil {
			return jobs, err
		}
		minute, err := strconv.Atoi(fields[3])
		if err != nil {
			return jobs, err
		}
		second, err := strconv.Atoi(fields[4])
		if err != nil {
			return jobs, err
		}

		data := ""

		// Validate the optional 6th field as JSON data
		if len(fields) > 5 {
			_, err := json.Marshal(fields[5])
			if err != nil {
				log.Fatalf("%s", fmt.Sprintf("%s : %s", line, err.Error()))
			}
			data = fields[5]
		}
		jobs = append(jobs, Job{
			method, url, hour, minute, second, data,
		})

	}
	return jobs, nil
}
