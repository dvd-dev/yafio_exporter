DOCKER_TAG := $(or ${GIT_TAG_NAME}, latest)
#GOFLAGS := "-ldflags \"-extldflags '-static' -linkmode 'external'\" -tags musl,netgo,osusergo"
#GOFLAGS := "-ldflags=-w -ldflags=-s"
BUILD_ARGS := -trimpath -ldflags='-extldflags=-static'
GOOS := linux
GOARCH := amd64

all: yafio_exporter


.PHONY: yafio_exporter
yafio_exporter:
				echo "GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GOFLAGS)"
				CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o yafio_exporter $(BUILD_ARGS)

.PHONY: dockerimages
dockerimages:
				docker build -t dvd-dev/yafio_exporter:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
				docker push dvd-dev/yafio_exporter:${DOCKER_TAG}

.PHONY: clean
clean:
				rm -f yafio_exporter