all: test vet build
#all: test vet fmt lint build

test:
	go test ./cmd/gen

vet:
	go vet ./cmd/gen

fmt:
	go list -f '{{.Dir}}' ./cmd/gen | grep -v /vendor/ | xargs -L1 gofmt -l
	test -z $$(go list -f '{{.Dir}}' ./cmd/gen | grep -v /vendor/ | xargs -L1 gofmt -l)

lint:
	go list ./cmd/gen | grep -v /vendor/ | xargs -L1 golint -set_exit_status

build:
	mkdir bin
	go build -o bin/gen ./cmd/gen
