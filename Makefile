.PHONY: all build test test-e2e lint proto clean coverage dep

all: build test

dep:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

build:
	go build -o bin/hci-vcls ./cmd/hci-vcls/

test:
	go test ./pkg/... ./internal/...

test-e2e:
	go test ./test/e2e/...

lint:
	export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	golangci-lint run

proto:
	export PATH="$$PATH:$$(go env GOPATH)/bin" && \
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