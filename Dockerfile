FROM golang:1.22-alpine AS build

ENV CGO_ENABLED=0

RUN apk add make binutils git fio bash

COPY . /app
WORKDIR /app

RUN go build -o yafio_exporter
RUN strip yafio_exporter

# NOTE(dvd): For some reason, go binary has to be present.
#FROM alpine:3.19
#RUN apk add fio

USER 65534:65534

#COPY --from=build /app/yafio_exporter .

EXPOSE 9996

ENTRYPOINT [ "/app/yafio_exporter" ]
