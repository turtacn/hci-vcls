// Package ha provides the High Availability Engine for HCI vCLS.
// This package is responsible for making HA decisions based on the current
// cluster view and executing them.
//
// The package is broken down into three primary responsibilities:
//
//  1. **Evaluator**: Uses cache, fdm agent, and the current cluster view
//     to decide if a VM should be restarted, where it should be restarted,
//     and at what priority.
//  2. **Batch Executor**: Takes a set of decisions from the evaluator and
//     manages the task of actually restarting the VMs across multiple nodes,
//     managing the active tasks and retries.
//  3. **HA Engine**: The main entry point that wires the evaluator and
//     batch executor, ensuring that HA decisions are only evaluated and
//     executed by the elected leader, confirming and releasing boot tokens
//     in MySQL as necessary.
package ha

// Personal.AI order the ending
