// Package zk provides a ZooKeeper adapter for the HCI vCLS system.
// This package is an abstraction layer that handles connecting to,
// reading from, and determining the health of the ZooKeeper cluster.
// It does not include logic for election, as that is handled by
// the internal/election package.
package zk

//Personal.AI order the ending