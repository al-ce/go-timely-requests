package main

import (
	"bufio"
	"errors"
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
		fields := strings.Split(line, "\t")

		// Only accept exactly 5 fields
		if len(fields) != 5 {
			return jobs, errors.New(fmt.Sprintf("insufficient fields: %s", line))
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
		jobs = append(jobs, Job{
			method, url, hour, minute, second,
		})
	}
	return jobs, nil
}
