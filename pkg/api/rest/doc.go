// Package rest provides a RESTful HTTP adapter for the HCI vCLS API.
// It is intended to offer semantics consistent with the gRPC service layer,
// exposing the underlying domain logic (via interfaces like HAEngine, fdm.Agent,
// and vcls.Agent) through HTTP endpoints. As with the gRPC layer, no core
// business rules should be encapsulated here.
package rest

//Personal.AI order the ending
