package main

import (
	"fmt"
	"os"

	"github.com/David2024patton/Mimir/internal/config"
	"github.com/David2024patton/Mimir/internal/server"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		runDaemon()
		return
	}
	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Println("mimir", version)
	case "daemon", "serve":
		runDaemon()
	case "auth":
		fmt.Println("mimir auth: login/logout/list (E1.3 - coming next)")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(2)
	}
}

func runDaemon() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("mimir %s starting daemon on port %d...\n", version, cfg.Port)
	if err := server.Serve(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "daemon error: %v\n", err)
		os.Exit(1)
	}
}
