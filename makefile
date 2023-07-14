.PHONY: profile

run:
	LOCAL=true go run app.go

view-profile-cpu:
	go tool pprof -http=localhost:8080 cpu.prof

view-profile-mem:
	go tool pprof -http=localhost:8080 mem.prof

bench-profile-cpu:
	LOCAL=true go test -bench=. -cpuprofile=cpu.prof

bench-profile-mem:
	LOCAL=true go test -bench=. -memprofile=mem.prof

bench:
	LOCAL=true go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

cpu: bench-profile-cpu view-profile-cpu

mem: bench-profile-mem view-profile-mem
