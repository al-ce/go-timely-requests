# go-timely-requests

Runs once-daily simple HTTP requests from read from a `.tsv` file.

Each row in the tsv should be in the format:

```
{method}\t{url}\t{hour}\t{minute}\t{second}\t{JSON}
```

Separated by actual tab characters.

The first five fields are mandatory. The sixth field must be a valid JSON on a single line, with no tabs in it.


## Example

```tsv
PUT	http://localhost:8080/topics/rotate	6	30	10
GET	http://localhost:3001/ping	6	30	18
POST	https://httpbin.org/post	7	18	5	{"name":"bob","age":"55"}
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
[JOB] 2025/05/15 02:18:05 POST https://httpbin.org/post: next job scheduled in 23h59m59.272370789s
[JOB] 2025/05/15 02:18:05 POST https://httpbin.org/post 200 OK {
 "args": {},
 "data": "{\"age\":\"55\",\"name\":\"bob\"}",
 "files": {},
 "form": {},
  # etc.
}
^C # attempt graceful shutdown
[JOB] 2025/05/15 01:30:21 stopping jobrunner...
[JOB] 2025/05/15 01:30:21 graceful shutdown
```

