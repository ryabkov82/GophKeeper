package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ryabkov82/gophkeeper/internal/pkg/logger"
	"github.com/ryabkov82/gophkeeper/internal/server"
	"github.com/ryabkov82/gophkeeper/internal/server/config"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	printBuildInfo()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("logger initialize failed: %v", err)
	}

	fmt.Printf("gRPC addr: %s\nDB DSN: %s\n", cfg.GRPCServerAddr, cfg.DBConnect)

	server.StartServer(logger.Log, cfg)

}

func printBuildInfo() {
	// Set default value "N/A" if variables are empty
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}
