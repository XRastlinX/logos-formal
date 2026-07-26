// Status: PROPOSED
// authority_effect: NONE
// Command cyxd runs the CyExchange Inbox/Outbox daemon.
package main

import (
	"fmt"
	"os"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cyxd <command>")
		fmt.Println("Commands: init, serve, route")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		InitCommand()
	case "serve":
		ServeCommand()
	case "route":
		RouteCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// Global adapters (stubs for PROPOSED state)
var globalInbox = inbox.NewAdapter()
