PACKAGES = $(shell go list ./... | grep -v /vendor/)

# Same as $(PACKAGES) except we get directory paths. We exclude the first line
# because it contains the top level directory which contains /vendor/
PACKAGE_DIRS=$(shell go list -f '{{ .Dir }}' ./... | egrep -v /vendor/ | tail -n +2)

SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/golang/lint/golint \
    github.com/golang/dep/cmd/dep \
    github.com/wadey/gocovmerge

check: deps docker build test coverage lint

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
	go fmt $(PACKAGES)
	goimports -w $(SOURCES)

test:
	go test -race -v $(PACKAGES)

# coverage runs the tests to collect coverage but does not attempt to look
# for race conditions.
coverage: $(patsubst %,%.coverage,$(PACKAGES))
	@rm -f .gocoverage/cover.txt
	gocovmerge .gocoverage/*.out > coverage.txt
	go tool cover -html=coverage.txt -o .gocoverage/index.html

%.coverage:
	@[ -d .gocoverage ] || mkdir .gocoverage
	go test -covermode=count -coverprofile=.gocoverage/$(subst /,-,$*).out $* -v
