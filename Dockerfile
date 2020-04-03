# Stage 1
FROM golang:1.13.9-alpine3.11 as builder

ARG BUILD_TOKEN

# Add git
RUN apk update && \
    apk add git && \
    apk add openssl-dev && \
    apk add gcc && \
    apk add libc-dev

RUN mkdir $GOPATH/src/github.com
RUN mkdir $GOPATH/src/github.com/YOVO-LABS
RUN mkdir $GOPATH/src/github.com/YOVO-LABS/workflow

ADD . $GOPATH/src/github.com/YOVO-LABS/workflow/
WORKDIR $GOPATH/src/github.com/YOVO-LABS/workflow

RUN export GOPRIVATE=github.com/YOVO-LABS/*

# RUN ls
RUN go get ./...
RUN go build cmd/main.go

# Stage 2
FROM alpine:3.11

RUN apk add --update \
    curl \
    ca-certificates \
    ffmpeg=4.2.1-r3

COPY --from=builder /go/bin/cmd /
EXPOSE 4000

CMD ["./cmd"]

# docker run -d --name transcoder --env-file=.env -p 4000:4000 asia.gcr.io/chrome-weft-229408/transcoder:v1
