LINTER_ARGS_SLOW = \
        -j 4 --enable-gc -s vendor -e '.*/vendor/.*' --vendor --enable=misspell --enable=gofmt --enable=goimports --enable=unused \
        --disable=dupl --disable=gocyclo --disable=errcheck --disable=golint --disable=interfacer --disable=gas \
        --disable=structcheck --disable=gotype --disable=gotypex --deadline=60m --tests

LINTER_ARGS = \
        $(LINTER_ARGS_SLOW) --disable=staticcheck --disable=unused --disable=gosimple

PKGS = $$(go list ./... | grep -v /vendor/)

ifeq ($(GOPATH),)
IGNORE := $(shell bash -c "eval `gimme stable` export GOPATH=`pwd`; env | sed 's/=/:=/' | sed 's/^/export /' > makeenv")                         
include makeenv
endif

all: build

build: main.go
	CGO_ENABLED=0 go build -a -installsuffix cgo --ldflags '-w' -o eaas

clean:
	rm -f debug debug.test eaas

test:
	echo $(GOPATH)
	go clean 
	go test -check.v

metalint:
	@if [ -z $$(go version | grep -o 'go1.5') ]; then \
			go get -u github.com/alecthomas/gometalinter; \
			gometalinter --install; \
			go install $(PKGS); \
			go test -i $(PKGS); \
			gometalinter $(LINTER_ARGS) ./...; \
	fi

race:
	go test $(GO_EXTRAFLAGS) -race -i $(PKGS)
	go test $(GO_EXTRAFLAGS) -race $(PKGS)

fmt:
	gofmt -l -w ./

