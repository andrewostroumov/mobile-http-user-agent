FROM golang:1.17.5-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o mobile-http-user-agent

FROM alpine:3.12

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/mobile-http-user-agent /build/.version /build/.revision /app/

CMD ["./mobile-http-user-agent", "--rev.ver-path=.version", "--rev.rev-path=.revision"]