FROM golang:1.21-alpine3.18 as builder

RUN apk add make binutils git

COPY . /app
WORKDIR /app

RUN apk add musl-dev gcc && make yafio_exporter

# NOTE(dvd): For some reason, go binary has to be present.
FROM alpine:3.18
RUN apk add fio

RUN mkdir -p /app && \
    adduser -D -u 12345 -g 12345 k6 && \
    mkdir -p /.cache && chown 12345:12345 /.cache /app

COPY --from=builder /app/yafio_exporter /app
USER k6
ENV SHELL=/bin/sh

EXPOSE 9996

ENTRYPOINT [ "/app/yafio_exporter" ]