# DevArch Gin (Go) Template

Gin framework template configured to serve static files from `public/` directory.

## Quick Start

```bash
# Initialize Go module
go mod init your-app

# Install Gin
go get -u github.com/gin-gonic/gin

# Run server
go run main.go
```

## main.go Configuration

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()

    // Serve static files from public/
    r.Static("/assets", "./public/assets")
    r.StaticFile("/", "./public/index.html")

    // API routes
    api := r.Group("/api")
    {
        api.GET("/hello", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "message": "Hello from Gin API",
            })
        })
    }

    // Start server
    r.Run(":8400")
}
```

## Structure

```
gin-app/
├── public/              # WEB ROOT
│   ├── index.html
│   └── assets/
├── main.go
├── go.mod
└── README.md
```

## DevArch Integration

### Port: 8400-8499 (Go range)
### Domain: Configure via Nginx Proxy Manager
### Container: golang service

## Documentation

See: https://gin-gonic.com/docs/
