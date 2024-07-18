# MantraMatch

MantraMatch is an open-source tool designed to identify and verify API keys across multiple services. It uses regex patterns to match API keys and attempts to verify them against their respective services.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Adding New Services](#adding-new-services)
- [Contributing](#contributing)
- [License](#license)

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
- `-ls`: List supported services

Examples:
```
mantramatch -verbose -timeout=15 your_api_key_here
mantramatch -silent -list=keys.txt
mantramatch -ls
```

Output format:
```
<API-KEY> : <valid/invalid>
Note: <note text if available>
----------------------------------------
```

## Configuration

MantraMatch uses a YAML configuration file to define services, their regex patterns, and verification endpoints. The default location for this file is `~/.config/mantramatch/config.yaml`.

Each service in the configuration file should include:
- `name`: Name of the service
- `regex`: Regex pattern to match the API key
- `verify_url`: URL to verify the API key
- `verify_method`: HTTP method for verification (GET, POST, etc.)
- `headers`: Any headers required for the verification request
- `validation`: Validation criteria for the response
- `note` (optional): Additional information about the service or API key

Example configuration entry:
```yaml
- name: "Example Service"
  regex: "^[a-zA-Z0-9]{32}$"
  verify_url: "https://api.example.com/verify"
  verify_method: "GET"
  headers:
    "Authorization": "Bearer %s"
  validation:
    status_code: 200
    success_indicator:
      type: "json_key_exists"
      key: "success"
  note: "This is an optional note for this service."
```

## Adding New Services

To add a new service to MantraMatch:

1. Open the `config.yaml` file.
2. Add a new service entry following the structure outlined in the [Configuration](#configuration) section.
3. Ensure the regex pattern accurately matches the API key format for the service.
4. Provide the correct verification URL and method.
5. Specify the appropriate headers, if any.
6. Define the validation criteria, including the success indicator type.
7. Add a note if there's any additional information users should know about the service or API key.
8. Test the new service configuration thoroughly.

## Contributing

Contributions to MantraMatch are welcome! Here's how you can contribute:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them with a clear commit message.
4. Push your changes to your fork.
5. Submit a pull request to the main repository.

Please ensure your code adheres to the existing style and include tests for new features.

## License

MantraMatch is released under the MIT License. See the [LICENSE](LICENSE) file for details.