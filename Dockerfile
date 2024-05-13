FROM golang:1.22-alpine AS build

ENV CGO_ENABLED=0

RUN apk add make binutils git fio bash

COPY . /app
WORKDIR /app

RUN go build -o yafio_exporter && strip yafio_exporter && mkdir -p /.cache && chown 65534:65534 /.cache

# NOTE(dvd): For some reason, go binary has to be present.
#FROM alpine:3.19
#RUN apk add fio

USER 65534:65534
ENV SHELL=/bin/sh

#COPY --from=build /app/yafio_exporter .

EXPOSE 9996

ENTRYPOINT [ "/app/yafio_exporter" ]
