PACKAGES = $(shell go list ./... | grep -v /vendor/)

# Same as $(PACKAGES) except we get directory paths. We exclude the first line
# because it contains the top level directory which contains /vendor/
PACKAGE_DIRS=$(shell go list -f '{{ .Dir }}' ./... | egrep -v /vendor/ | tail -n +2)

SOURCES = $(shell for f in $(PACKAGES); do ls $$GOPATH/src/$$f/*.go; done)
EXTRA_DEPENDENCIES = \
    github.com/golang/lint/golint \
    github.com/tools/godep \
    github.com/alecthomas/gometalinter

check: deps vet lint test

deps:
	go get $(EXTRA_DEPENDENCIES)
	gometalinter --install > /dev/null

lint:
	gometalinter --vendor --disable-all --enable=deadcode --enable=errcheck --enable=goimports \
	--enable=gocyclo --enable=golint --enable=gosimple --enable=misspell \
	--enable=unconvert --enable=unused --enable=varcheck --enable=interfacer \
	./...

fmt:
	gofmt -w -s $(SOURCES)
	goimports -w $(SOURCES)

vet:
	go vet $(PACKAGES)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -check.v $(PACKAGES)
