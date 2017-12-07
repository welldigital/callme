# API

## General

### Possible statuses

* 200 (OK): OK, e.g. item deleted
* 201 (Status created): success
* 400 (Bad request): malformed or unreadable body
* 422 (Unprocessable entity): invalid JSON, failure to validate or invalid operation
* 500 (Internal server error): Internal errors

### Failures

Failures return JSON in the following format.

```json
{"err":"message"}
```

## Jobs

Allows the scheduling of a job in the future (or the past, in which case it will execute immediately, but mess up the delay metrics).

## POST `:8080/job/`

```bash
curl --header "Content-Type: application/json" -d '{"when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload":"test_payload"}' http://localhost:8080/job
```

```json
{"jobId":1,"scheduleId":null,"when":"2000-01-01T00:00:00Z","arn":"example.com","payload":"test_payload"}
```

## GET `:8080/job/{id}`

```bash
curl http://localhost:8080/job/1
```

```json
{"job":{"jobId":1,"scheduleId":null,"when":"2000-01-01T00:00:00Z","arn":"example.com","payload":"test_payload"},"response":{"jobResponseId":0,"jobId":0,"time":"0001-01-01T00:00:00Z","response":"","isError":false,"error":""},"hasJobResponse":false}
```

## POST `:8080/job/{id}/delete

```bash
curl -d {} http://localhost:8080/job/1/delete
```

```json
{"ok":true}
```

## Schedules

Allows a recurring schedule to be setup.

## POST `:8080/schedule/`

```bash
curl --header "Content-Type: application/json" -d '{"arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}' http://localhost:8080/schedule
```

```json
{"scheduleId":2,"from":"0001-01-01T00:00:00Z","arn":"testarn","payload":"testpayload","crontabs":["* * * * *"],"externalId":"testexternalid","by":"testby"}
```

## GET `:8080/schedule/{id}`

```bash
curl http://localhost:8080/schedule/1
```

```json
{"schedule":{"scheduleId":1,"externalId":"testexternalid","by":"testby","arn":"testarn","payload":"testpayload","created":"2017-12-07T17:55:00.37897Z","active":true,"deactivatedDate":"0001-01-01T00:00:00Z"},"crontabs":[{"crontabId":1,"scheduleId":1,"crontab":"* * * * *","previous":"0001-01-01T00:00:00Z","next":"0001-01-01T00:00:00Z","lastUpdated":"0001-01-01T00:00:00Z"}]}
```

## POST `:8080/schedule/{id}/deactivate

```bash
curl -d {} http://localhost:8080/schedule/1/deactivate
```

```json
{"ok":true}
```
