BINARIES=bin/v1tov2

GOFILES=$(shell find . -type f -name '*.go')

DOCKERFILES=$(shell find . -type f -name 'Dockerfile')

BUILDKIT_HOST="tcp://0.0.0.0:1234"

.DEFAULT: all
all: v1tov2

binaries: $(BINARIES)

bin/%: build
	@echo "+ $@"

buildkit:
	@echo "+ $@"
	@docker inspect buildkit > /dev/null || docker run -d --name buildkit --privileged -p 1234:1234 tonistiigi/buildkit:latest --debug --addr ${BUILDKIT_HOST}

build: buildkit $(GOFILES)
	@echo "+ $@"
	@go get -u github.com/moby/buildkit/cmd/buildctl
	@go run cmd/buildkit-lemming/main.go | BUILDKIT_HOST=${BUILDKIT_HOST} buildctl build --local src=. --exporter local --exporter-opt output=./bin

lemming: bin/lemming $(DOCKERFILES)
	@echo "+ $@"
	@BUILDKIT_HOST=${BUILDKIT_HOST} buildctl build --frontend=dockerfile.v0 --local context=. --local dockerfile=cmd/lemming --exporter=docker --exporter-opt name=docker/lemming | docker load

lemmingd: bin/lemmingd $(DOCKERFILES)
	@echo "+ $@"
	@BUILDKIT_HOST=${BUILDKIT_HOST} buildctl build --frontend=dockerfile.v0 --local context=. --local dockerfile=cmd/lemmingd --exporter=docker --exporter-opt name=docker/lemmingd | docker load

lemmingctl: bin/lemmingctl $(DOCKERFILES)
	@echo "+ $@"
	@BUILDKIT_HOST=${BUILDKIT_HOST} buildctl build --frontend=dockerfile.v0 --local context=. --local dockerfile=cmd/lemmingctl --exporter=docker --exporter-opt name=docker/lemmingctl | docker load

clean:
	@echo "+ $@"
	@docker rm -f buildkit
	@rm -rf "./bin"

vendor:
	@echo "+ $@"
	@./hack/update-vendor

.PHONY: buildkit build lemming lemmingd lemmingctl clean vendor

