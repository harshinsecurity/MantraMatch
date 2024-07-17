package service

import (
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

	return isValidResponse(service, resp.StatusCode, resp.Header, body, verbose)
}

func createRequest(service config.Service, apiKey string) (*http.Request, error) {
	url := strings.ReplaceAll(service.VerifyURL, "%s", apiKey)
	req, err := http.NewRequest(service.VerifyMethod, url, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range service.Headers {
		req.Header.Add(key, strings.ReplaceAll(value, "%s", apiKey))
	}

	return req, nil
}

func isValidResponse(service config.Service, statusCode int, headers http.Header, body []byte, verbose bool) bool {
	if statusCode != service.Validation.StatusCode {
		logError(fmt.Sprintf("%s returned unexpected status code: %d", service.Name, statusCode), verbose)
		return false
	}

	if service.Validation.ContentType != "" && !strings.HasPrefix(headers.Get("Content-Type"), service.Validation.ContentType) {
		logError(fmt.Sprintf("%s returned unexpected content type: %s", service.Name, headers.Get("Content-Type")), verbose)
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
		regex := regexp.MustCompile(service.Validation.SuccessIndicator.Value)
		return regex.Match(body)
	case "header_exists", "header_value":
		return validateHeaderResponse(service, headers, verbose)
	default:
		logError(fmt.Sprintf("Unknown validation type for %s: %s", service.Name, service.Validation.SuccessIndicator.Type), verbose)
		return false
	}
}

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

func validateHeaderResponse(service config.Service, headers http.Header, verbose bool) bool {
	value := headers.Get(service.Validation.SuccessIndicator.Key)
	if service.Validation.SuccessIndicator.Type == "header_exists" {
		return value != ""
	}
	return value == service.Validation.SuccessIndicator.Value
}

func logError(message string, verbose bool) {
	if verbose {
		log.Println(message)
	}
}