# MarketMosaic Common Library for Go

A common library for Go microservices that provides shared functionality and utilities.

## Features

- Standardized logging with multiple backends (Logrus and Zap)
- Common HTTP response formats
- Error handling utilities
- Request validation helpers
- UUID generation utilities

## Installation

```bash
go get https://github.com/Kunal726/market-mosaic-common-lib-go
```

## Usage

### Logging

```go
import "https://github.com/Kunal726/market-mosaic-common-lib-go/pkg/logger"

// Initialize logger
logger.InitLogger()

// Use logger
logger.Info("This is an info message")
logger.Error("This is an error message")
```

### HTTP Responses

```go
import "https://github.com/Kunal726/market-mosaic-common-lib-go/pkg/http"

// Success response
http.Success(c, data)

// Error response
http.Error(c, err)
```

## License

MIT
