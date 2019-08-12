#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR:= $(ROOTDIR)/_output
PACKAGE_PATH:=github.com/solo-io/valet
SOURCES := $(shell find . -name "*.go" | grep -v test.go | grep -v '\.\#*')
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	# TAGGED_VERSION := $(shell git describe --tags)
	# This doesn't work in CI, need to find another way...
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)
LDFLAGS := "-X github.com/solo-io/valet/cli/version.Version=$(VERSION)"

#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

#----------------------------------------------------------------------------------
# Build
#----------------------------------------------------------------------------------

.PHONY: build
build:
	go build -ldflags=$(LDFLAGS) -o _output/valet -v main.go

.PHONY: build-linux
build-linux:
	GO111MODULE=on CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -o _output/valet-linux-amd64 -v main.go

#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
docker-build: build-linux
	docker build -t gcr.io/solo-test-236622/valet:$(VERSION) .

docker-push: docker-build
	docker push gcr.io/solo-test-236622/valet:$(VERSION)