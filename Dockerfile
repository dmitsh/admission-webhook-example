FROM golang:1.15 as builder

ADD . /build
WORKDIR /build

ENV CGO_ENABLED=0

RUN GO111MODULE=on go get ./...
RUN go build -a -ldflags '-extldflags "-static"' ./cmd/initc
RUN go build -a -ldflags '-extldflags "-static"' ./cmd/webhook

FROM scratch

COPY --from=builder /build/initc   /usr/local/bin/initc
COPY --from=builder /build/webhook /usr/local/bin/webhook
