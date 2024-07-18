package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/harshinsecurity/mantramatch/internal/config"
	"github.com/harshinsecurity/mantramatch/internal/service"
	"github.com/schollz/progressbar/v3"
)

var (
	configFile   string
	verbose      bool
	silent       bool
	timeout      int
	listFile     string
	listServices bool
	initConfig   bool
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
	flag.BoolVar(&listServices, "ls", false, "List supported services")
	flag.BoolVar(&initConfig, "init-config", false, "Initialize default configuration file")
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
	fmt.Fprintf(os.Stderr, "  mantramatch -ls\n")
	fmt.Fprintf(os.Stderr, "  mantramatch -init-config\n")
}

func main() {
	if initConfig {
		err := config.CreateDefaultConfig(configFile)
		if err != nil {
			fmt.Printf("Error creating default configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Default configuration file created at: %s\n", configFile)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: Configuration file not found at %s\n", configFile)
			fmt.Println("Run 'mantramatch -init-config' to create a default configuration file.")
			os.Exit(1)
		}
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if listServices {
		printSupportedServices(cfg)
		return
	}

	if listFile != "" {
		processKeyList(cfg)
	} else if len(flag.Args()) == 1 {
		processKey(cfg, flag.Args()[0])
	} else {
		flag.Usage()
		os.Exit(1)
	}
}

func printSupportedServices(cfg *config.Config) {
	fmt.Println("Supported services:")
	for _, service := range cfg.Services {
		fmt.Printf("- %s\n", service.Name)
	}
}

func processKey(cfg *config.Config, apiKey string) {
	matchedServices := service.MatchServices(cfg.Services, apiKey)
	if len(matchedServices) == 0 {
		if !silent {
			fmt.Printf("%s : invalid\n", apiKey)
			fmt.Println("No matching services found for the given API key.")
			fmt.Println(strings.Repeat("-", 40))
		}
		return
	}

	results := verifyKeys(matchedServices, apiKey)
	printResults(results, apiKey, matchedServices)
}

func processKeyList(cfg *config.Config) {
	file, err := os.Open(listFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var keys []string
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	bar := progressbar.Default(int64(len(keys)))

	for _, apiKey := range keys {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(key string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			processKey(cfg, key)
			bar.Add(1)
		}(apiKey)
	}

	wg.Wait()
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

func printResults(results map[string]bool, apiKey string, services []config.Service) {
	foundValid := false
	for _, s := range services {
		valid := results[s.Name]
		status := "invalid"
		if valid {
			status = "valid"
			foundValid = true
		}
		fmt.Printf("%s : %s\n", apiKey, status)

		if !silent && s.Note != "" {
			fmt.Printf("Note: %s\n", s.Note)
		}

		if !silent {
			fmt.Println(strings.Repeat("-", 40))
		}
	}

	if !foundValid && !silent {
		fmt.Println("No valid services found for this API key.")
		fmt.Println("This could mean the key is invalid, expired, or not supported by MantraMatch.")
		fmt.Println(strings.Repeat("-", 40))
	}
}
