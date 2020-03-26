# Command variables
GO_BUILD 	= go build
GO_PLUGIN 	= go build -buildmode=plugin
INSTALL 	= /usr/bin/install
MKDIR 		= mkdir -p
RM 		= rm
CP 		= cp
DOCKER_COMPOSE ?= docker-compose
DOCKER_COMPOSE_EXEC ?= docker-compose exec -T

# Optimization build processes
#CPUS ?= $(shell nproc)
#MAKEFLAGS += --jobs=$(CPUS)

OS = $(shell uname -s | tr 'A-Z' 'a-z')
ARCH = amd64

ifeq ($(shell uname -m), x86_64)
	ARCH = amd64
endif

# Project variables
PROJECT_PKG ?= github.com/uthng/container-injector
PROJECT_PATH ?= $(GOPATH)/src/go/$(PROJECT_PKG)
PROJECT_BIN_DIR ?= bin
PROJECT_PLUGIN_DIR ?= plugins

# Compilation variables
PROJECT_BUILD_SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')
PROJECT_BUILD_OSARCH = darwin/amd64 linux/amd64
PROJECT_BUILD_TARGET = container-injector

all: clean build

# Build targets multiple platforms
build: clean
	for osarch in $(PROJECT_BUILD_OSARCH); do \
		OS=`echo $$osarch | cut -d"/" -f1`; \
		ARCH=`echo $$osarch | cut -d"/" -f2`; \
		echo "Compiling $(PROJECT_BUILD_TARGET) for "$$OS"_"$$ARCH"..." ; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags="-s -w" -o $(PROJECT_BIN_DIR)"/"$$OS"_"$$ARCH"/"$(PROJECT_BUILD_TARGET); \
	done

optimize:
	for osarch in $(PROJECT_BUILD_OSARCH); do \
		OS=`echo $$osarch | cut -d"/" -f1`; \
		ARCH=`echo $$osarch | cut -d"/" -f2`; \
		echo "Optimizing $(PROJECT_BUILD_TARGET) for "$$OS"_"$$ARCH"..." ; \
		upx --brute $(PROJECT_BIN_DIR)"/"$$OS"_"$$ARCH"/"$(PROJECT_BUILD_TARGET) ; \
		for plugin in $(PROJECT_BUILD_PLUGINS); do \
		echo "Optimizing "$$plugin".so for "$$OS"_"$$ARCH"..." ; \
			upx --brute $(PROJECT_BIN_DIR)"/"$$OS"_"$$ARCH"/"$(PROJECT_PLUGIN_DIR)"/"$$plugin".so" ; \
		done; \
	done

test-unit:
# Use flag -p 1 to force not to run test in parallel because of
# the presence of different secrets/auths in diffrent tests
	go test -count 1 -p 1 -v -tags=unit -cover ./...

docker-test-unit: docker-stop
	$(DOCKER_COMPOSE) up -d
# Use flag -p 1 to force not to run test in parallel because of
# the presence of different secrets/auths in diffrent tests
	go test -count 1 -p 1 -v -tags=unit -cover ./...
	$(DOCKER_COMPOSE) down

linters:
	golangci-lint run ./...

fmt:
	gofmt -s -l -w $(PROJECT_BUILD_SRCS)

deps:
	go get -u github.com/mitchellh/gox

clean:
	-$(RM) -rf bin

docker-start:
	$(DOCKER_COMPOSE) up -d

docker-stop:
	$(DOCKER_COMPOSE) down

distclean:

install:

.PHONY: all build optimize distclean clean fmt deps install test-unit docker-test-unit docker-stop docker-start lint
