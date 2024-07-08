package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/harshinsecurity/mantramatch/internal/config"
)

// MatchServices finds all services that match the given API key
func MatchServices(services []config.Service, apiKey string) []config.Service {
	var matches []config.Service
	for _, service := range services {
		regex := regexp.MustCompile(service.Regex)
		if regex.MatchString(apiKey) {
			matches = append(matches, service)
		}
	}
	return matches
}

// VerifyKey checks if the given API key is valid for the specified service
func VerifyKey(service config.Service, apiKey string, timeout int, verbose bool) bool {
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	req, err := createRequest(service, apiKey)
	if err != nil {
		logError(fmt.Sprintf("Error creating request for %s: %v", service.Name, err), verbose)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		logError(fmt.Sprintf("Error making request to %s: %v", service.Name, err), verbose)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logError(fmt.Sprintf("Error reading response from %s: %v", service.Name, err), verbose)
		return false
	}

	return isValidResponse(service, resp.StatusCode, body, verbose)
}

// createRequest creates an http.Request for the given service and API key
func createRequest(service config.Service, apiKey string) (*http.Request, error) {
	req, err := http.NewRequest(service.VerifyMethod, service.VerifyURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range service.Headers {
		req.Header.Add(key, fmt.Sprintf(value, apiKey))
	}

	return req, nil
}

// isValidResponse checks if the API response indicates a valid key
func isValidResponse(service config.Service, statusCode int, body []byte, verbose bool) bool {
	if statusCode != http.StatusOK {
		logError(fmt.Sprintf("%s returned non-200 status code: %d", service.Name, statusCode), verbose)
		return false
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logError(fmt.Sprintf("Error parsing JSON response from %s: %v", service.Name, err), verbose)
		return false
	}

	if service.SuccessKey != "" {
		_, hasSuccessKey := result[service.SuccessKey]
		return hasSuccessKey
	}

	_, hasError := result["error"]
	return !hasError
}

// logError logs an error message if verbose mode is enabled
func logError(message string, verbose bool) {
	if verbose {
		log.Println(message)
	}
}
