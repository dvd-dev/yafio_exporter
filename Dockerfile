FROM golang:1.22-alpine3.19

RUN apk add make binutils git fio bash

COPY . /app
WORKDIR /app

RUN apk add musl-dev gcc && make yafio_exporter

RUN mkdir -p /app && \
    adduser -D -u 12345 -g 12345 k6 && \
    mkdir -p /.cache && chown 12345:12345 /.cache /app
USER k6
ENV SHELL=/bin/sh
EXPOSE 9996
ENTRYPOINT [ "/app/yafio_exporter" ]