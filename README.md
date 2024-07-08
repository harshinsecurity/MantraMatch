# MantraMatch

MantraMatch is an open-source tool designed to identify and verify API keys across multiple services. It uses regex patterns to match API keys and attempts to verify them against their respective services.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Areas for Improvement](#areas-for-improvement)
- [Contributing](#contributing)
- [License](#license)

## Installation

To install MantraMatch, you need to have Go installed on your system (version 1.16 or later). Then, you can install it directly using `go install`:

```
go install github.com/harshinsecurity/mantramatch@latest
```

This command will download the source code, compile it, and install the `mantramatch` binary in your `$GOPATH/bin` directory. Make sure this directory is in your system's PATH.

After installation, you need to set up the configuration file:

1. Create a directory for the configuration:
   ```
   mkdir -p ~/.config/mantramatch
   ```

2. Download the default configuration file:
   ```
   curl -o ~/.config/mantramatch/config.yaml https://raw.githubusercontent.com/harshinsecurity/mantramatch/main/configs/config.yaml
   ```

   Or, if you prefer to use `wget`:
   ```
   wget -O ~/.config/mantramatch/config.yaml https://raw.githubusercontent.com/harshinsecurity/mantramatch/main/configs/config.yaml
   ```

## Usage

After installation, you can use MantraMatch from the command line:

```
mantramatch [options] <api-key>
```

Options:
- `-config`: Path to the configuration file (default: ~/.config/mantramatch/config.yaml)
- `-verbose`: Enable verbose output
- `-timeout`: Timeout for HTTP requests in seconds (default: 10)

Example:
```
mantramatch -verbose -timeout=15 your_api_key_here
```

## How It Works

1. MantraMatch reads the configuration file that contains regex patterns and verification endpoints for various services.
2. It matches the input API key against these patterns to identify potential services.
3. For each matched service, it attempts to verify the API key by making a request to the service's verification endpoint.
4. The tool reports which services (if any) the API key is valid for.

## Configuration

MantraMatch uses a YAML configuration file to define services, their regex patterns, and verification endpoints. The default location for this file is `~/.config/mantramatch/config.yaml`.

You can add or modify services in this file. Each service entry should include:
- `name`: Name of the service
- `regex`: Regex pattern to match the API key
- `verify_url`: URL to verify the API key
- `verify_method`: HTTP method for verification (GET, POST, etc.)
- `headers`: Any headers required for the verification request

Example configuration entry:
```yaml
- name: "Example Service"
  regex: "^[a-zA-Z0-9]{32}$"
  verify_url: "https://api.example.com/verify"
  verify_method: "GET"
  headers:
    "Authorization": "Bearer %s"
```

## Areas for Improvement

1. **More Services**: Add support for additional API services.
2. **Better Error Handling**: Improve error messages and handling for different types of failures.
3. **Concurrent Verification**: Implement concurrent verification for faster results with multiple services.
4. **GUI**: Develop a graphical user interface for easier use.
5. **API Integration**: Create an API that other tools can integrate with.
6. **Periodic Updates**: Implement a system to periodically update the regex patterns and verification endpoints.


### Reporting Issues

If you find a bug or have a suggestion for improvement, please open an issue on the GitHub repository. Provide as much detail as possible, including steps to reproduce any bugs.