bench:
	go test -v -run=XXX -bench=. -benchtime 5s -benchmem
graph:
	 go test -run=XXX -bench . -cpuprofile cpu.prof
	 go tool pprof -svg go-rbush.test cpu.prof > cpu1.svg
