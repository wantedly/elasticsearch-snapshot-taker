FROM golang:1.11-alpine3.7

RUN apk add -U ca-certificates curl git gcc musl-dev make
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 \
		&& chmod +x /usr/local/bin/dep

RUN mkdir -p $GOPATH/src/github.com/wantedly/elasticsearch/snapshot
WORKDIR $GOPATH/src/github.com/wantedly/elasticsearch/snapshot

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only -v

ADD . ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -o /snapshot


FROM alpine:3.7
COPY --from=0 /etc/ssl /etc/ssl
COPY --from=0 /snapshot /snapshot
ENTRYPOINT ["/snapshot"]

