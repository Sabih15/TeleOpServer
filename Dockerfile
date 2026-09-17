FROM golang:1.24-alpine AS builder

RUN apk add --no-cache bash
ENV SHELL=/bin/bash

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/api

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
