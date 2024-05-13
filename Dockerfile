FROM golang:1.22-alpine AS build

ENV CGO_ENABLED=0

RUN apk add make binutils git fio bash

COPY . /app
WORKDIR /app

RUN go build -o yafio_exporter &&\
    strip yafio_exporter && \
    adduser -D -u 12345 -g 12345 k6 && \
    mkdir -p /.cache && chown 12345:12345 /.cache

# NOTE(dvd): For some reason, go binary has to be present.
#FROM alpine:3.19
#RUN apk add fio

USER k6
ENV SHELL=/bin/sh

#COPY --from=build /app/yafio_exporter .

EXPOSE 9996

ENTRYPOINT [ "/app/yafio_exporter" ]
