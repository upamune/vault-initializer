FROM golang:1.11.0
WORKDIR /go/src/app
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep \
    && dep ensure -vendor-only
RUN CGO_ENABLE=0 GOOS=linux go build -ldflags="-s -w" -o vault-initializer -v .

FROM gcr.io/distroless/base
COPY --from=0 /go/src/app/vault-initializer .
ENTRYPOINT ["/vault-initializer"]

