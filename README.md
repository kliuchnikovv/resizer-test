# Field Materials Backend Interview skeleton code

## Build & Run Server locally
```
go run .
```

## Run a sample request against the server
Default variant without any parameters
```
curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize
```

Sync variant - same as previous
```
curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize
```

Async variant
```
curl -X POST -H "Content-Type: application/json" -d @req.json http://localhost:8080/v1/resize?async=true
```

Now in your browser, you can check one of the returned urls!


P.S.
1) Since there are no requirements for handling queue overflows, we simply fail the request if there’s no available slot.
2) For the same reason, retrieval is implemented via periodic cache checks instead of events.
