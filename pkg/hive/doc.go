// pckage hive provides the infrastructure for building dolphin applications from modular
// components (cells)

// this package is implemented using the uber/dig library, which provides the dependency injection for
// objects in the hive. It is similar to uber/fix, but adds an opinionated approach to configuration
// Example
// for a runnable example see pkg/hive/example
// example$ go run .
// go run . --dot-graph | dot -Tx11
package hive
