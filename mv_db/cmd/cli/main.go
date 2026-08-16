package main

import (
	"bufio"
	"fmt"
	"log"
	"mv_db/internal/config"
	"mv_db/network"
	"os"
	"strings"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Can't load config.yaml (%v). Use default values.", err)
		cfg = config.NewDefaultConfig()
	}
	logger.Info("Config is loaded:", zap.Object("config", cfg))

	client := network.NewTCPClient(cfg.Network.Address)

	if err := client.Connect(); err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to mvdb. Type your commands (SET, GET, DEL):")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			break
		}

		if line == "" {
			continue
		}

		resp, err := client.Send(line)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		fmt.Print(resp)
	}
}
