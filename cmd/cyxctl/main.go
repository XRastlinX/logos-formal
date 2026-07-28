// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cyxctl <command> [args]")
		fmt.Println("Commands: send, inspect-node, verify-receipt, conformance, reliability")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "send":
		SendCommand(os.Args[2:])
	case "inspect-node":
		InspectCommand(os.Args[2:])
	case "verify-receipt":
		VerifyCommand(os.Args[2:])
	case "conformance":
		ConformanceCommand()
	case "reliability":
		if err := ReliabilityCommand(); err != nil {
			fmt.Printf("[FAIL] Reliability suite: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
