FROM golang:1.15.3-buster as builder

WORKDIR /go/src/github.com/KzmnbrS/golang_test

COPY go.mod ./

COPY go.sum ./

RUN go mod download

EXPOSE 3000

COPY src ./src

RUN go build -o app ./src

FROM debian:buster-20201012

WORKDIR /opt/app

RUN mkdir /opt/persist && mkdir /opt/persist/images

RUN apt-get update && apt-get -y install sqlite && apt clean

COPY ./schema.sql ./

RUN cat schema.sql | sqlite3 /opt/persist/images.db

COPY --from=builder /go/src/github.com/KzmnbrS/golang_test/app ./app

ENTRYPOINT ./app

