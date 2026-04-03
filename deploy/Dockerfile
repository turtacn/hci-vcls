FROM golang:1.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/hci-vcls ./cmd/hci-vcls

FROM gcr.io/distroless/base-debian11

COPY --from=builder /app/hci-vcls /hci-vcls
ENTRYPOINT ["/hci-vcls"]
