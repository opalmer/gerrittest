PACKAGES = $(shell go list . | grep -v /vendor/)
PACKAGE_DIRS = $(shell go list -f '{{ .Dir }}' ./... | grep -v /vendor/)
SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/kardianos/govendor \
    github.com/golang/lint/golint

check: deps docker build test

docker:
	$(MAKE) -C docker build

build:
	go build cmd/gerrittest.go

deps:
	go get $(EXTRA_DEPENDENCIES)
	govendor sync
	rm -rf $(GOPATH)/src/github.com/docker/docker/vendor
	rm -rf vendor/github.com/docker/docker/vendor

fmt:
	goimports -w $(SOURCES)
	go fmt ./...

test:
	go test -race -v -check.v ./...
