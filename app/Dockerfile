FROM golang:1.15.3-buster as builder

WORKDIR /go/src/github.com/KzmnbrS/golang_test

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY src ./src

RUN go build -o app ./src

FROM debian:buster-20201012

WORKDIR /opt/app

COPY --from=builder /go/src/github.com/KzmnbrS/golang_test/app ./app

ENTRYPOINT ./app

