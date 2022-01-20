FROM golang:1.15-alpine as builder

RUN apk update && \
    apk add bash git openssh alpine-sdk libgit2-dev=1.1.0-r2

WORKDIR /code

COPY . /code
RUN mkdir -p ./bin && \
    go build -o ./bin ./...

FROM alpine:3.14

RUN apk update && \
    apk add libgit2=1.1.0-r2

COPY --from=builder /code/bin/* /opt/resource/