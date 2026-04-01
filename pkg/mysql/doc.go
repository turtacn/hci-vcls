// Package mysql provides a MySQL adapter for the HCI vCLS system.
// This package is an abstraction layer that handles connecting to,
// reading from, and writing to the MySQL cluster.
// It manages boot tokens for HA operations to prevent duplicate boots
// in the event of a split-brain scenario, and is responsible for
// the state of VM configurations as they pertain to HA.
package mysql

//Personal.AI order the ending
