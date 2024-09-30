ARG GOVERSION=1.22
ARG ALPINE_VERSION=3.19

FROM golang:${GOVERSION} AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -a -o venator .

FROM alpine:${ALPINE_VERSION} AS ca-certificates
RUN apk add --no-cache ca-certificates

FROM scratch
WORKDIR /app
COPY --from=builder /app/venator .
COPY --from=ca-certificates /etc/ssl/certs/* /etc/ssl/certs/
ENTRYPOINT ["./venator"]