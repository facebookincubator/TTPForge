# TTPForge/network

The `network` package is a part of the TTPForge.

---

## Table of contents

- [Functions](#functions)
- [Installation](#installation)
- [Usage](#usage)
- [Tests](#tests)
- [Contributing](#contributing)
- [License](#license)

---

## Functions

### DisableHTTPProxy()

```go
DisableHTTPProxy()
```

DisableHTTPProxy unsets the "http_proxy" environment variable.
This disables the HTTP proxy for network operations in the current process.

---

### DisableHTTPSProxy()

```go
DisableHTTPSProxy()
```

DisableHTTPSProxy unsets the "https_proxy" and "http_proxy"
environment variables. This disables the HTTPS proxy for network
operations in the current process.

Parameters:

proxy: A string representing the URL of the proxy server to be used for HTTPS connections.

---

### DisableNoProxy()

```go
DisableNoProxy()
```

DisableNoProxy unsets the "no_proxy" environment variable.
This clears the list of domains excluded from being
proxied for network operations in the current process.

---

### EnableHTTPProxy(string)

```go
EnableHTTPProxy(string)
```

EnableHTTPProxy sets the "http_proxy" environment variable
to the provided proxy string. This enables an HTTP proxy
for network operations in the current process.

Parameters:

proxy: A string representing the URL of the proxy server to be used for HTTP connections.

---

### EnableHTTPSProxy(string)

```go
EnableHTTPSProxy(string)
```

EnableHTTPSProxy sets the environment variables "https_proxy"
and "http_proxy" to the provided proxy string.
This enables an HTTPS proxy for network operations in the current process.

Parameters:

proxy: A string representing the URL of the proxy server to be used for HTTPS connections.

---

### EnableNoProxy(string)

```go
EnableNoProxy(string)
```

EnableNoProxy sets the "no_proxy" environment variable to
the provided domains string. This excludes the specified domains
from being proxied for network operations in the current process.

Parameters:

domains: A string representing a comma-separated list of domain names
to be excluded from proxying.

---

## Installation

To use the TTPForge/network package, you first need to install it.
Follow the steps below to install via go get.

```bash
go get github.com/ttpforge/facebookincubator/network
```

---

## Usage

After installation, you can import the package in your Go project
using the following import statement:

```go
import "github.com/ttpforge/facebookincubator/network"
```

---

## Tests

To ensure the package is working correctly, run the following
command to execute the tests for `TTPForge/network`:

```bash
go test -v
```

---

## Contributing

Pull requests are welcome. For major changes,
please open an issue first to discuss what
you would like to change.

---

## License

This project is licensed under the MIT
License - see the [LICENSE](../LICENSE)
file for details.
