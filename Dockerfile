# Stage 1
FROM golang:1.13.0-alpine3.10 as builder

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
RUN go get ./

RUN go build cmd/main.go

# Stage 2
FROM alpine:3.10

RUN apk add --update \
    curl \
    ca-certificates \
    ffmpeg \
    python \
    python-dev \
    py-pip \
    build-base \
    && pip install awscli==$AWSCLI_VERSION --upgrade --user \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

COPY --from=builder /go/bin/main /

EXPOSE 4000

CMD ["./main"]