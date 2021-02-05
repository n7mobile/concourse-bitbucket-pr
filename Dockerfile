FROM golang:1.15-alpine as builder

RUN apk update && \
    apk add --no-cache bash git openssh alpine-sdk libgit2-dev

WORKDIR /code

COPY go.* /code/
RUN go mod download

COPY . /code
RUN go build -o ./bin ./...

FROM alpine

RUN apk update && \
    apk add --no-cache libgit2

COPY --from=builder /code/bin/* /opt/resource/