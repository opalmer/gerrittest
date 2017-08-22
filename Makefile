PACKAGES = $(shell go list ./... | grep -v /vendor/)

# Same as $(PACKAGES) except we get directory paths. We exclude the first line
# because it contains the top level directory which contains /vendor/
PACKAGE_DIRS=$(shell go list -f '{{ .Dir }}' ./... | egrep -v /vendor/ | tail -n +2)

SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/golang/lint/golint \
    github.com/golang/dep/cmd/dep \
    github.com/wadey/gocovmerge
TEST_CMD_PREFIX ?= go test -v
TESTCMD = $(TEST_CMD_PREFIX)

ifdef TEST_SHORT
TESTCMD := $(TESTCMD) -short
endif

check: deps vet docker lint build test coverage

docker:
	$(MAKE) -C docker build

build:
	go build cmd/gerrittest.go

lint:
	golint -set_exit_status $(PACKAGES)

deps:
	go get $(EXTRA_DEPENDENCIES)
	dep ensure

vet:
	go vet $(PACKAGES)

fmt:
	gofmt -w -s $(SOURCES)
	goimports -w $(SOURCES)

test:
	$(TESTCMD) -race $(PACKAGES)

# coverage runs the tests to collect coverage but does not attempt to look
# for race conditions.
coverage: $(patsubst %,%.coverage,$(PACKAGES))
	@rm -f .gocoverage/cover.txt
	gocovmerge .gocoverage/*.out > coverage.txt
	go tool cover -html=coverage.txt -o .gocoverage/index.html

%.coverage:
	@[ -d .gocoverage ] || mkdir .gocoverage
	$(TESTCMD) -covermode=count -coverprofile=.gocoverage/$(subst /,-,$*).out $*

bindata:
	go-bindata -pkg internal -o internal/internal.go internal/commit-msg
	$(MAKE) fmt
