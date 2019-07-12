

all:
	go build

install:
	go install

test:
	go test -v -cover -race
	@# -race is better for load and integration tests
