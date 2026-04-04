// Package fdm implements the Fault Domain Manager for HCI vCLS.
// This package is responsible for probing the health of nodes (L0/L1/L2),
// determining the status of peers through heartbeats, maintaining the
// local and cluster-wide degradation levels, and handling leader election.
// It detects and informs about failures and degradation but is not
// responsible for actually executing High Availability (HA) decisions.
package fdm

// Personal.AI order the ending
