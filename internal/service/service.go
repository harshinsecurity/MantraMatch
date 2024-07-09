package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
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

	return isValidResponse(service, resp, body, verbose)
}

// createRequest creates an http.Request for the given service and API key
func createRequest(service config.Service, apiKey string) (*http.Request, error) {
	req, err := http.NewRequest(service.VerifyMethod, service.VerifyURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range service.Headers {
		if strings.Contains(value, "%s") {
			req.Header.Add(key, fmt.Sprintf(value, apiKey))
		} else {
			req.Header.Add(key, value)
		}
	}

	// Special handling for Basic Auth
	if authHeader := req.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Basic") {
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(apiKey + ":"))
		req.Header.Set("Authorization", "Basic "+encodedAuth)
	}

	return req, nil
}

// isValidResponse checks if the API response indicates a valid key
func isValidResponse(service config.Service, resp *http.Response, body []byte, verbose bool) bool {
	if resp.StatusCode != service.Validation.StatusCode {
		logError(fmt.Sprintf("%s returned unexpected status code: %d", service.Name, resp.StatusCode), verbose)
		return false
	}

	if service.Validation.ContentType != "" && !strings.HasPrefix(resp.Header.Get("Content-Type"), service.Validation.ContentType) {
		logError(fmt.Sprintf("%s returned unexpected content type: %s", service.Name, resp.Header.Get("Content-Type")), verbose)
		return false
	}

	switch service.Validation.SuccessIndicator.Type {
	case "status_code_only":
		return true
	case "json_key_exists", "json_key_value":
		return validateJSONResponse(service, body, verbose)
	case "contains_string":
		return strings.Contains(string(body), service.Validation.SuccessIndicator.Value)
	case "regex_match":
		re, err := regexp.Compile(service.Validation.SuccessIndicator.Value)
		if err != nil {
			logError(fmt.Sprintf("Invalid regex for %s: %v", service.Name, err), verbose)
			return false
		}
		return re.Match(body)
	case "header_exists", "header_value":
		return validateHeaderResponse(service, resp.Header, verbose)
	default:
		logError(fmt.Sprintf("Unknown validation type for %s: %s", service.Name, service.Validation.SuccessIndicator.Type), verbose)
		return false
	}
}

// validateJSONResponse validates JSON responses
func validateJSONResponse(service config.Service, body []byte, verbose bool) bool {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logError(fmt.Sprintf("Error parsing JSON response from %s: %v", service.Name, err), verbose)
		return false
	}

	value, exists := result[service.Validation.SuccessIndicator.Key]
	if !exists {
		return false
	}

	if service.Validation.SuccessIndicator.Type == "json_key_value" {
		return fmt.Sprintf("%v", value) == service.Validation.SuccessIndicator.Value
	}

	return true
}

// validateHeaderResponse validates header-based responses
func validateHeaderResponse(service config.Service, headers http.Header, verbose bool) bool {
	value := headers.Get(service.Validation.SuccessIndicator.Key)
	if service.Validation.SuccessIndicator.Type == "header_exists" {
		return value != ""
	}
	return value == service.Validation.SuccessIndicator.Value
}

// logError logs an error message if verbose mode is enabled
func logError(message string, verbose bool) {
	if verbose {
		log.Println(message)
	}
}
