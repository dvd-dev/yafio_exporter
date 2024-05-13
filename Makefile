GO111MODULE := on
DOCKER_TAG := $(or ${GIT_TAG_NAME}, latest)

all: yafio_exporter

.PHONY: yafio_exporter
yafio_exporter:
    mkdir -p bin/
	go build -o bin/yafio_exporter
	strip bin/yafio_exporter

.PHONY: dockerimages
dockerimages:
	docker build -t dvd-dev/yafio_exporter:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push dvd-dev/yafio_exporter:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*
