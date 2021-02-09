FROM golang:1.15-alpine as builder

RUN apk update && \
    apk add --no-cache bash git openssh alpine-sdk libgit2-dev

WORKDIR /code

COPY . /code
RUN mkdir -p ./bin && \
    go build -o ./bin ./...

FROM alpine

RUN apk update && \
    apk add --no-cache libgit2

COPY --from=builder /code/bin/* /opt/resource/