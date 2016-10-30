PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
SOURCES=$(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)

build: deps check test
	go build -o gerrittest cmd/main.go

check: fmt vet lint

deps:
	which golint > /dev/null || go get -u github.com/golang/lint/golint
	which govendor > /dev/null || go get -u github.com/kardianos/govendor
	govendor fetch +missing

fmt:
	go fmt $(PACKAGES)

vet:
	go vet $(PACKAGES)

lint: deps
	(for f in $(SOURCES); do golint -set_exit_status $$f; done)

test:
	./test.sh
