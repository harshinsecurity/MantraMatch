// scripts/install.go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting user home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".config", "mantramatch")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		fmt.Printf("Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	srcConfig, err := os.Open("configs/config.yaml")
	if err != nil {
		fmt.Printf("Error opening source config file: %v\n", err)
		os.Exit(1)
	}
	defer srcConfig.Close()

	dstConfig, err := os.Create(filepath.Join(configDir, "config.yaml"))
	if err != nil {
		fmt.Printf("Error creating destination config file: %v\n", err)
		os.Exit(1)
	}
	defer dstConfig.Close()

	_, err = io.Copy(dstConfig, srcConfig)
	if err != nil {
		fmt.Printf("Error copying config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Config file installed successfully!")
}