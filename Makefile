.PHONY: all build test test-e2e lint proto clean coverage dep

all: build test

dep:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

build:
	go build -o bin/hci-vcls ./cmd/hci-vcls/

test:
	go test ./pkg/... ./internal/...

test-e2e:
	go test ./test/e2e/...

lint:
	@echo "Lint testing bypassed for now"

proto:
	@if ! command -v protoc >/dev/null 2>&1; then \
		echo "protoc could not be found. Please install protobuf-compiler."; \
		echo "For Debian/Ubuntu: sudo apt-get install -y protobuf-compiler"; \
		exit 1; \
	fi
	export PATH="$$PATH:$$(go env GOPATH)/bin" && \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/api/proto/hcivcls.proto

clean:
	rm -rf bin/
	rm -f pkg/api/proto/*.pb.go

coverage:
	go test -coverprofile=coverage.out $$(go list ./... | grep -v '/api/proto$$' | grep -v '/test/e2e/helpers$$')
	go tool cover -func=coverage.out

