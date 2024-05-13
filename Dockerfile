FROM golang:1.20-alpine AS build

ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN apk add make binutils git

COPY . /app
WORKDIR /app


RUN go build -o yafio_exporter
RUN strip yafio_exporter

FROM alpine:3.17

RUN apk add fio

WORKDIR /
USER 65534:65534

COPY --from=build /app/yafio_exporter .

EXPOSE 9996

ENTRYPOINT [ "./yafio_exporter" ]
