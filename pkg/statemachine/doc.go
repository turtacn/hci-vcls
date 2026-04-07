// Package statemachine implements a state machine for the cluster's degradation
// levels based on a combination of different metrics sampled over time, like the
// health of Zookeeper, CFS, MySQL, and the Fault Domain Manager.
//
// The evaluation of these rules is separated from the machine execution to keep
// the logic pure, easily testable, and loosely coupled.
package statemachine

