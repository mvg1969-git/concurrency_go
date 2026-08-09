package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"mv_db/network"
)

func main() {
	addr := "127.0.0.1:3223"
	client := network.NewTCPClient(addr)

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
