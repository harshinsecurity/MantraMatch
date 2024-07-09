# MantraMatch

MantraMatch is an open-source tool designed to identify and verify API keys across multiple services. It uses regex patterns to match API keys and attempts to verify them against their respective services.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
  - [Service Configuration](#service-configuration)
  - [Validation Types](#validation-types)
- [How It Works](#how-it-works)
- [Contributing](#contributing)
  - [Adding New Services](#adding-new-services)
- [Areas for Improvement](#areas-for-improvement)
- [Reporting Issues](#reporting-issues)

## Installation

To install MantraMatch, you need to have Go installed on your system (version 1.16 or later). Then, you can install it directly using `go install`:

```
go install github.com/harshinsecurity/mantramatch@latest
```

This command will download the source code, compile it, and install the `mantramatch` binary in your `$GOPATH/bin` directory. Make sure this directory is in your system's PATH.

## Usage

After installation, you can use MantraMatch from the command line:

```
mantramatch [options] <api-key>
```

Options:
- `-config`: Path to the configuration file (default: ~/.config/mantramatch/config.yaml)
- `-verbose`: Enable verbose output
- `-silent`: Show only verified API keys and services
- `-timeout`: Timeout for HTTP requests in seconds (default: 10)
- `-list`: Path to file containing list of API keys

Examples:
```
mantramatch -verbose -timeout=15 your_api_key_here
mantramatch -silent -list=keys.txt
```

## Configuration

MantraMatch uses a YAML configuration file to define services, their regex patterns, and verification endpoints. The default location for this file is `~/.config/mantramatch/config.yaml`.

### Service Configuration

Each service in the configuration file should include:

```yaml
- name: "Service Name"
  regex: "^regex_pattern_here$"
  verify_url: "https://api.example.com/verify"
  verify_method: "GET"
  headers:  # Optional
    "Authorization": "Bearer %s"
  validation:
    status_code: 200
    content_type: "application/json"  # Optional
    success_indicator:
      type: "json_key_exists"
      key: "user_id"

# Example of a service without headers
- name: "Simple API"
  regex: "^[A-Za-z0-9]{32}$"
  verify_url: "https://api.example.com/verify?key=%s"
  verify_method: "GET"
  validation:
    status_code: 200
    success_indicator:
      type: "status_code_only"
```

Note: The `headers` field is optional. For services that don't require headers, you can omit this field. In such cases, you might need to include the API key in the `verify_url` (as shown in the "Simple API" example).

### Validation Types

The `success_indicator` supports the following types:

1. `status_code_only`: Only checks the HTTP status code.
2. `json_key_exists`: Checks if a specific key exists in the JSON response.
3. `json_key_value`: Checks if a specific key in the JSON response has a particular value.
4. `contains_string`: Checks if the response body contains a specific string.
5. `regex_match`: Checks if the response body matches a given regular expression.
6. `header_exists`: Checks if a specific header exists in the response.
7. `header_value`: Checks if a specific header in the response has a particular value.

## How It Works

1. MantraMatch reads the configuration file that contains regex patterns and verification endpoints for various services.
2. It matches the input API key against these patterns to identify potential services.
3. For each matched service, it attempts to verify the API key by making a request to the service's verification endpoint.
4. The tool reports which services (if any) the API key is valid for.

## Contributing

Contributions to MantraMatch are welcome! Here's how you can contribute:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them with a clear commit message.
4. Push your changes to your fork.
5. Submit a pull request to the main repository.

### Adding New Services

To add a new service to MantraMatch:

1. Open the `config.yaml` file.
2. Add a new service entry following the structure outlined in the [Service Configuration](#service-configuration) section.
3. Ensure the regex pattern accurately matches the API key format for the service.
4. Provide the correct verification URL and method.
5. Specify the appropriate headers, if any.
6. Define the validation criteria, including the success indicator type.
7. Test the new service configuration thoroughly.
8. Submit a pull request with your changes.

Please ensure your code adheres to the existing style and include tests for new features.

## Areas for Improvement

1. **More Services**: Add support for additional API services.
2. **Better Error Handling**: Improve error messages and handling for different types of failures.
3. **GUI**: Develop a graphical user interface for easier use.
4. **API Integration**: Create an API that other tools can integrate with.
5. **Periodic Updates**: Implement a system to periodically update the regex patterns and verification endpoints.

## Reporting Issues

If you find a bug or have a suggestion for improvement, please open an issue on the GitHub repository. Provide as much detail as possible, including steps to reproduce any bugs.