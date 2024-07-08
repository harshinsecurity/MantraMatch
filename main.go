package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/harshinsecurity/mantramatch/internal/config"
	"github.com/harshinsecurity/mantramatch/internal/service"
)

var (
	configFile string
	verbose    bool
	timeout    int
)

func init() {
	flag.Usage = usage
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".config", "mantramatch", "config.yaml")
	flag.StringVar(&configFile, "config", defaultConfigPath, "Path to configuration file")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.IntVar(&timeout, "timeout", 10, "Timeout for HTTP requests in seconds")
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "MantraMatch: A tool to identify and verify API keys\n\n")
	fmt.Fprintf(os.Stderr, "Usage: mantramatch [options] <api-key>\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  mantramatch -verbose -timeout=15 your_api_key_here\n\n")
	fmt.Fprintf(os.Stderr, "Before first use, run: go run scripts/install.go\n")
	fmt.Fprintf(os.Stderr, "This will install the default configuration file.\n")
}

func main() {
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	apiKey := flag.Args()[0]

	config, err := internal.LoadConfig(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error: Config file not found.")
			fmt.Println("Please run 'go run scripts/install.go' to install the default config.")
			fmt.Printf("Expected config location: %s\n", configFile)
			os.Exit(1)
		}
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	matchedServices := internal.MatchServices(config.Services, apiKey)

	if len(matchedServices) == 0 {
		fmt.Println("No matching services found for the given API key.")
		return
	}

	fmt.Printf("Potential matches found: %d\n", len(matchedServices))
	results := verifyKeys(matchedServices, apiKey)

	foundValid := false
	for service, valid := range results {
		if valid {
			fmt.Printf("✅ API key is valid for %s\n", service)
			foundValid = true
		} else {
			fmt.Printf("❌ API key is not valid for %s\n", service)
		}
	}

	if !foundValid {
		fmt.Println("\nNo valid services found for this API key.")
		fmt.Println("This could mean the key is invalid, expired, or not supported by MantraMatch.")
	}
}

func verifyKeys(services []internal.Service, apiKey string) map[string]bool {
	results := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, service := range services {
		wg.Add(1)
		go func(s internal.Service) {
			defer wg.Done()
			valid := internal.VerifyKey(s, apiKey, timeout, verbose)
			mu.Lock()
			results[s.Name] = valid
			mu.Unlock()
		}(service)
	}

	wg.Wait()
	return results
}
