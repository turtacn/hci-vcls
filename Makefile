.PHONY: all build test test-e2e lint proto clean coverage

all: build test

build:
	go build -o bin/hci-vcls ./cmd/hci-vcls/

test:
	go test ./pkg/... ./internal/...

test-e2e:
	go test ./test/e2e/...

lint:
	golangci-lint run

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/api/proto/hcivcls.proto

clean:
	rm -rf bin/
	rm -f pkg/api/proto/*.pb.go

coverage:
	go test -coverprofile=coverage.out ./pkg/... ./internal/...
	go tool cover -html=coverage.out

# Personal.AI order the ending