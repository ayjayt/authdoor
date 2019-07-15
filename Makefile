
all:
	go build

install:
	go install

test:
	go test -v -cover -race
	go test -v -run=xxx -test.bench=. -test.benchmem -cpuprofile profile_cpu.out

heap:
	go build -gcflags '-m -m'
	@# two more m's are possible but its too verbose
	@# -race is better for load and integration tests
