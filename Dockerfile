FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/hci-vcls /app/hci-vcls

EXPOSE 8080 9090

ENTRYPOINT ["/app/hci-vcls"]

# Personal.AI order the ending