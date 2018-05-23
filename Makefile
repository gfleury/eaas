# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
LINTER_ARGS_SLOW = \
        -j 4 --enable-gc -s vendor -e '.*/vendor/.*' --vendor --enable=misspell --enable=gofmt --enable=goimports --enable=unused \
        --disable=dupl --disable=gocyclo --disable=errcheck --disable=golint --disable=interfacer --disable=gas \
        --disable=structcheck --disable=gotype --disable=gotypex --deadline=60m --tests

LINTER_ARGS = \
        $(LINTER_ARGS_SLOW) --disable=staticcheck --disable=unused --disable=gosimple

PKGS = $$(go list ./... | grep -v /vendor/)


ifeq ($(GOROOT),)
.EXPORT_ALL_VARIABLES:
eval("gimme stable")
endif

all: build

build: main.go
	CGO_ENABLED=0 go build -a -installsuffix cgo --ldflags '-w' -o eaas

clean:
	rm -f debug debug.test eaas

test:
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

