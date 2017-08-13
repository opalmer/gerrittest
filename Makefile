PACKAGES = $(shell go list . | grep -v /vendor/)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./... | grep -v /vendor/)
SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/kardianos/govendor \
    github.com/golang/lint/golint

check: deps docker build test lint

docker:
	$(MAKE) -C docker build

build:
	go build cmd/gerrittest.go

lint:
	golint -set_exit_status $(PACKAGES)

deps:
	go get $(EXTRA_DEPENDENCIES)
	dep ensure

fmt:
	goimports -w $(SOURCES)
	go fmt ./...

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)
