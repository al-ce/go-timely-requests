# go-timely-requests

Runs once-daily HTTP requests from read from a `.tsv` file.

Each row in the tsv should be in the format

```
{method}\t{url}\t{hour}\t{minute}\t{second}
```


## Example

```tsv
PUT	http://localhost:8080/topics/rotate	6	30	10
GET	http://localhost:3001/ping	6	30	18
```



```bash
❯ go run . jobs.tsv
[JOB] 2025/05/15 01:30:05 scheduled: {PUT http://localhost:8080/topics/rotate 6 30 10}
[JOB] 2025/05/15 01:30:05 scheduled: {GET http://localhost:3001/ping 6 30 18}
[JOB] 2025/05/15 01:30:05 PUT http://localhost:8080/topics/rotate: next job scheduled in 4.569317909s
[JOB] 2025/05/15 01:30:05 GET http://localhost:3001/ping: next job scheduled in 12.569299507s
[JOB] 2025/05/15 01:30:10 PUT http://localhost:8080/topics/rotate 200 OK
[JOB] 2025/05/15 01:30:10 PUT http://localhost:8080/topics/rotate: next job scheduled in 23h59m59.96376295s
[JOB] 2025/05/15 01:30:18 GET http://localhost:3001/ping 200 OK
[JOB] 2025/05/15 01:30:18 GET http://localhost:3001/ping: next job scheduled in 23h59m59.99417753s
^C[JOB] 2025/05/15 01:30:21 stopping jobrunner...
[JOB] 2025/05/15 01:30:21 graceful shutdown
```

