package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/harshinsecurity/mantramatch/internal/config"
	"github.com/harshinsecurity/mantramatch/internal/service"
)

//go:embed config.yaml
var embeddedFiles embed.FS

var (
	configFile string
	verbose    bool
	silent     bool
	timeout    int
	listFile   string
)

func init() {
	flag.Usage = usage
	homeDir, _ := os.UserHomeDir()
	defaultConfigPath := filepath.Join(homeDir, ".config", "mantramatch", "config.yaml")
	flag.StringVar(&configFile, "config", defaultConfigPath, "Path to configuration file")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&silent, "silent", false, "Show only verified API keys and services")
	flag.IntVar(&timeout, "timeout", 10, "Timeout for HTTP requests in seconds")
	flag.StringVar(&listFile, "list", "", "Path to file containing list of API keys")
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "MantraMatch: A tool to identify and verify API keys\n\n")
	fmt.Fprintf(os.Stderr, "Usage: mantramatch [options] <api-key>\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  mantramatch -verbose -timeout=15 your_api_key_here\n")
	fmt.Fprintf(os.Stderr, "  mantramatch -silent -list=keys.txt\n")
}

func main() {
	if err := ensureConfig(); err != nil {
		fmt.Printf("Error ensuring config: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if listFile != "" {
		processKeyList(cfg, listFile)
	} else if len(flag.Args()) == 1 {
		processKey(cfg, flag.Args()[0])
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func ensureConfig() error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		dir := filepath.Dir(configFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		embeddedConfig, err := embeddedFiles.ReadFile("config.yaml")
		if err != nil {
			return fmt.Errorf("failed to read embedded config: %w", err)
		}

		if err := os.WriteFile(configFile, embeddedConfig, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		if !silent {
			fmt.Printf("Config file created at: %s\n", configFile)
		}
	}
	return nil
}

func processKey(cfg *config.Config, apiKey string) {
	matchedServices := service.MatchServices(cfg.Services, apiKey)
	if len(matchedServices) == 0 {
		if !silent {
			fmt.Println("No matching services found for the given API key.")
		}
		return
	}

	if !silent {
		fmt.Printf("Potential matches found: %d\n", len(matchedServices))
	}

	results := verifyKeys(matchedServices, apiKey)
	printResults(results, apiKey)
}

func processKeyList(cfg *config.Config, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	for scanner.Scan() {
		apiKey := scanner.Text()
		wg.Add(1)
		semaphore <- struct{}{}
		go func(key string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			processKey(cfg, key)
		}(apiKey)
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func verifyKeys(services []config.Service, apiKey string) map[string]bool {
	results := make(map[string]bool)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, svc := range services {
		wg.Add(1)
		go func(s config.Service) {
			defer wg.Done()
			valid := service.VerifyKey(s, apiKey, timeout, verbose)
			mu.Lock()
			results[s.Name] = valid
			mu.Unlock()
		}(svc)
	}

	wg.Wait()
	return results
}

func printResults(results map[string]bool, apiKey string) {
	foundValid := false
	for serviceName, valid := range results {
		if valid {
			if silent {
				fmt.Printf("%s: %s\n", apiKey, serviceName)
			} else {
				fmt.Printf("✅ API key is valid for %s\n", serviceName)
			}
			foundValid = true
		} else if !silent {
			fmt.Printf("❌ API key is not valid for %s\n", serviceName)
		}
	}

	if !foundValid && !silent {
		fmt.Println("\nNo valid services found for this API key.")
		fmt.Println("This could mean the key is invalid, expired, or not supported by MantraMatch.")
	}
}
