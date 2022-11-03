FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY . .

# downloading dependencies
RUN go mod download

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
RUN go build -ldflags="-s -w" -v -o producer ./cmd/producer/producer.go 

FROM scratch

COPY --from=builder ["/app/producer", "/"]

ENTRYPOINT [ "/producer" ]

