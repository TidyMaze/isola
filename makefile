.PHONY: profile

run:
	LOCAL=true go run app.go

view-profile-cpu:
	go tool pprof -http=localhost:8080 cpu.prof

view-profile-mem:
	go tool pprof -http=localhost:8080 mem.prof

bench-profile-cpu:
	go test -bench=. -cpuprofile=cpu.prof

bench-profile-mem:
	go test -bench=. -memprofile=mem.prof
