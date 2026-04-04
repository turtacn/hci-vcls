// Package witness implements the Witness mechanism for HCI vCLS.
// This package is responsible for confirming the failure of nodes
// in the event of a potential split-brain scenario. It polls a
// list of external witnesses to determine whether a given node is
// truly down or if there is a network partition, ensuring the HA
// operations do not erroneously restart a VM that is still active.
package witness

// Personal.AI order the ending
