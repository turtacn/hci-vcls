# API Protobuf Definitions

This directory contains the Protocol Buffers (protobuf) definitions for the HCI vCLS services.

## Generating Go Code

To generate the corresponding Go gRPC and protobuf code, run the following command from the root of the project:

```bash
make proto
```

## Generated Artifacts

The generated `.pb.go` and `_grpc.pb.go` files will be placed in the same directory as the `.proto` file (`pkg/api/proto`).

## Package Structure

The proto package is `hcivcls` and the `go_package` option is set to `github.com/turtacn/hci-vcls/pkg/api/proto` to align with the Go module's structure.

<!-- Personal.AI order the ending -->