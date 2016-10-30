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

test: $(patsubst %,%.test,$(PACKAGES))

%.test:
	@[ -d .coverage ] || mkdir .coverage
	go test -v -race -covermode=atomic -coverprofile=.coverage/$(subst /,-,$*).out $*
	@[ ! -e .coverage/$(subst /,-,$*).out ] || go tool cover -html=.coverage/$(subst /,-,$*).out -o .coverage/$(subst /,-,$*).html
