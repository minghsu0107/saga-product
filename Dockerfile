FROM golang:1.15.2 AS builder

RUN mkdir -p /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build-linux

FROM alpine:3.12

RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /app/server /app/config.yml ./
RUN apk add --no-cache bash

ENTRYPOINT ["./server"]