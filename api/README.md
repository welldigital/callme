# API

## Jobs

Allows the scheduling of a job in the future (or the past, in which case it will execute immediately, but mess up the delay metrics).

## POST `:8080/job/`

### Request

```
Content-Type: application/json
```

```json
{ 
	"when": "2000-01-01T00:00:00Z", 
	"arn": "example.com", 
	"payload": "test_payload"
}
```

### Reponse

```
Content-Type: application/json
```

#### Success

```json
{"JobID":1,"ScheduleID":null,"When":"2000-01-01T00:00:00Z","ARN":"example.com","Payload":"test_payload"}
```

#### Failures

```json
{"err":"message"}
```

### Possible statuses

* 201 (Status created): success
* 400 (Bad request): malformed or unreadable body
* 422 (Unprocessable entity): invalid JSON
* 422 (Unprocessable entity): tried to update an existing job
* 422 (Unprocessable entity): tried to add a job against a schedule
* 422 (Unprocessable entity): required fields were not present or didn't pass validation
* 500 (Internal server error): Unable to start a job
* 500 (Unable to start a job)

### Example

```bash
curl --header "Content-Type: application/json" -d '{"when": "2000-01-01T00:00:00Z", "arn": "example.com", "payload":"test_payload"}' http://localhost:8080/job
```