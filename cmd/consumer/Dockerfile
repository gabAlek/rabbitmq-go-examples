FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY . .

# downloading dependencies
RUN go mod download

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
RUN go build -ldflags="-s -w" -o consumer ./cmd/consumer/consumer.go

FROM scratch

COPY --from=builder ["/app/consumer", "/"]

ENTRYPOINT [ "/consumer" ]

