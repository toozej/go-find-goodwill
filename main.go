// Package main provides the entry point for the go-find-goodwill application.
//
// This application serves as a template for Go projects, demonstrating
// best practices for CLI applications using cobra, logrus, and environment
// configuration management.
package main

import cmd "github.com/toozej/go-find-goodwill/cmd/go-find-goodwill"

// main is the entry point of the go-find-goodwill application.
// It delegates execution to the cmd package which handles all
// command-line interface functionality.
func main() {
	cmd.Execute()
}
