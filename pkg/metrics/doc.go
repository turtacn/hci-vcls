// Package metrics implements an interface and specific implementations for capturing
// operational metrics for the HCI vCLS system.
//
// This package includes definitions for labels and constant metrics names, allowing
// developers to rely on a consistent set of identifiers. It implements a fully-featured
// Prometheus version for use in production environments and a No-op implementation
// for testing and local development where metric collection isn't necessary.
package metrics

//Personal.AI order the ending
