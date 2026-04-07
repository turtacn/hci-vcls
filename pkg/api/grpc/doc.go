// Package grpc provides the gRPC transport layer implementation for the HCI vCLS API.
// It acts solely as an adapter, adapting gRPC requests into calls on the core
// domain objects (like HAEngine, fdm.Agent, vcls.Agent) and mapping the
// domain results back to the gRPC response formats. Core decision-making logic
// should not reside in this package.
package grpc

