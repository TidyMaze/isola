.PHONY: profile

run:
	LOCAL=true go run app.go

profile:
	go tool pprof -http=localhost:8080 cpu.prof
