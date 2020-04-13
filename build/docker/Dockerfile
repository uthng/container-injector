##############################################
# STAGE 1: Build container-injector
##############################################

FROM golang:1.14-alpine as builder

RUN apk add --update make git build-base

# Copy files
COPY . /go/src/github.com/uthng/container-injector
WORKDIR /go/src/github.com/uthng/container-injector

# Build static binaries
RUN make deps build

##############################################
# STAGE 2: Build container
##############################################

FROM alpine:3.11

LABEL maintainer="Thanh NGUYEN <thanh@uthng.me>" \
  name="container-injector" \
  description="A Kubernetes mutating webhook server that implements container injection" \
  summary="A Kubernetes mutating webhook server that implements container injection"

RUN apk add --update ca-certificates curl gnupg libcap openssl su-exec iputils

ENV HTTPADDR ":8443"
ENV VERBOSITY 4
ENV TLS_CERT ""
ENV TLS_KEY ""

RUN addgroup container-injector && \
    adduser -S -G container-injector container-injector

RUN mkdir -p /app

COPY --from=builder /go/src/github.com/uthng/container-injector/bin/linux_amd64/container-injector /app/

WORKDIR /app

USER container-injector

ENTRYPOINT ["/app/container-injector"]
