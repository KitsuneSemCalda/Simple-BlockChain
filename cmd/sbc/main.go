package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"KitsuneSemCalda/SBC/internal/structures"
)

func main() {
	blockchain := structures.NewBlockchain()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Simple Blockchain CLI")
	fmt.Println("Commands: add <bpm>, print, validate, length, quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}
		parts := strings.Fields(input)
		command := parts[0]
		switch command {
		case "add":
			if len(parts) < 2 {
				fmt.Println("Usage: add <bpm>")
				continue
			}
			bpm, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Error: BPM must be a number")
				continue
			}
			blockchain.AddBlock(bpm)
			fmt.Printf("Block #%d added successfully!\n", blockchain.Length()-1)
		case "print":
			blockchain.Print()
		case "validate":
			if blockchain.IsValid() {
				fmt.Println("Blockchain is valid!")
			} else {
				fmt.Println("Blockchain is INVALID!")
			}
		case "length":
			fmt.Printf("Blockchain length: %d\n", blockchain.Length())
		case "quit", "exit":
			fmt.Println("Goodbye!")
			break
		default:
			fmt.Printf("Unknown command: %s\n", command)
			fmt.Println("Available commands: add, print, validate, length, quit")
		}
	}
}
