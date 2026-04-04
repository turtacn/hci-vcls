// Package cache implements the cache manager for the HCI vCLS system.
// This package is responsible for caching metadata related to VMs, such as
// Compute, Network, Storage, and HA state. It synchronizes this metadata
// from various sources (like CFS or MySQL) to local disk stores to
// enable fast HA evaluation even in the event of network partition or ZK
// degradation. It only holds the metadata and manages the lifecycle of
// the cache entries; it does not make HA decisions.
package cache

// Personal.AI order the ending
