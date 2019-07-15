
all:
	go build

install:
	go install

test:
	go test -v -cover -race
	go test -v -run=xxx -test.bench=. -test.benchmem -cpuprofile profile_cpu.out

# The following requires graphviz
viz:
	go tool pprof -svg profile_cpu.out > profile_cpu.svg

# The following requires github.com/uber/go-torch and github.com/brendangregg/FlameGraph
# Here it is installed to $(HOME)/software/FlameGraph
torch:
	PATH=$(PATH):$(HOME)/software/FlameGraph go-torch -b profile_cpu.out -f profile_cpu.torch.svg

heap:
	go build -gcflags '-m -m'
	@# two more m's are possible but its too verbose
	@# -race is better for load and integration tests
