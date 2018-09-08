FROM golang:1.11.0-alpine3.8
WORKDIR /go/src/app
COPY . .
RUN apk --no-cache add git \
    && go get -u github.com/golang/dep/cmd/dep \
    && dep ensure -vendor-only
RUN CGO_ENABLE=0 GOOS=linux go build -ldflags="-s -w" -o vault-initializer -v .

FROM alpine:3.8
RUN apk --no-cache add ca-certificates openssl
COPY --from=0 /go/src/app/vault-initializer .
ENTRYPOINT ["/vault-initializer"]
