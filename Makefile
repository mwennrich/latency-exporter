GO111MODULE := on
DOCKER_TAG := $(or ${GIT_TAG_NAME}, latest)

all: latency-exporter

.PHONY: latency-exporter
latency-exporter:
	go build -o bin/latency-exporter *.go
	strip bin/latency-exporter

.PHONY: dockerimages
dockerimages: 
	docker build -t mwennrich/latency-exporter:${DOCKER_TAG} .

.PHONY: dockerpush
dockerpush:
	docker push mwennrich/latency-exporter:${DOCKER_TAG}

.PHONY: clean
clean:
	rm -f bin/*
