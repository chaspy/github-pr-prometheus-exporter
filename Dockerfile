FROM golang:1.22.6 as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY ./main.go  ./

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build \
    -o /go/bin/github-pr-prometheus-exporter \
    -ldflags '-s -w'

FROM alpine:3.22.1 as runner

COPY --from=builder /go/bin/github-pr-prometheus-exporter /app/github-pr-prometheus-exporter

RUN adduser -D -S -H exporter
USER exporter

ENTRYPOINT ["/app/github-pr-prometheus-exporter"]
