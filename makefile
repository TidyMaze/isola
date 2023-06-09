.PHONY: profile

run:
	go run app.go

profile:
	go tool pprof -http=localhost:8080 cpu.prof
