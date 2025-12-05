# GoLand - Echo Framework Quick Start

Create Echo web applications in DevArch using GoLand with container-based Go toolchain.

## Prerequisites

- GoLand installed
- DevArch containers running:
  ```bash
  ./scripts/service-manager.sh start database proxy backend
  ```

## 1. Create Project

**Initialize Go module:**

```bash
podman exec -it go bash
cd /app
mkdir my-echo-app
cd my-echo-app
go mod init github.com/yourname/my-echo-app
go get -u github.com/labstack/echo/v4
go get -u github.com/labstack/echo/v4/middleware
exit
```

**Open in GoLand:**
1. File → Open → `/home/fhcadmin/projects/devarch/apps/my-echo-app`

## 2. Setup `public/` Structure

```
apps/my-echo-app/
├── public/              # Static assets
│   ├── css/
│   ├── js/
│   └── images/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   ├── models/
│   └── middleware/
├── templates/
│   └── index.html
├── go.mod
└── go.sum
```

## 3. Configure Go SDK

1. Settings → Go → GOROOT → "..." → "Add SDK..."
2. Docker Compose:
   - Configuration file: `/home/fhcadmin/projects/devarch/compose/backend/go.yml`
   - Service: `go`
3. Click "OK"

## 4. Create Echo Application

**Create `cmd/server/main.go`:**

```go
package main

import (
    "net/http"
    "html/template"
    "io"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

// Template renderer
type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Template renderer
    t := &Template{
        templates: template.Must(template.ParseGlob("templates/*.html")),
    }
    e.Renderer = t

    // Serve static files
    e.Static("/static", "public")

    // Routes
    e.GET("/", func(c echo.Context) error {
        return c.Render(http.StatusOK, "index.html", map[string]interface{}{
            "title": "Echo Application",
        })
    })

    e.GET("/api/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]interface{}{
            "status":  "ok",
            "message": "Echo is running",
        })
    })

    e.GET("/api/users", func(c echo.Context) error {
        return c.JSON(http.StatusOK, []map[string]interface{}{
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
        })
    })

    // Start server
    e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}
```

**Create `templates/index.html`:**

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <h1>Welcome to Echo</h1>
    <p>Running in DevArch container</p>
</body>
</html>
```

**Install dependencies:**

```bash
podman exec -it go bash
cd /app/my-echo-app
go mod tidy
```

## 5. Development Workflow

**Start Echo server:**

```bash
podman exec -it go bash
cd /app/my-echo-app
go run cmd/server/main.go
```

**Server runs on:** http://localhost:8400

## 6. Database Integration (GORM + PostgreSQL)

**Install dependencies:**

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

**Create `internal/database/db.go`:**

```go
package database

import (
    "fmt"
    "log"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        getEnv("DB_HOST", "postgres"),
        getEnv("DB_USER", "postgres"),
        getEnv("DB_PASSWORD", "admin1234567"),
        getEnv("DB_NAME", "my_echo_db"),
        getEnv("DB_PORT", "5432"),
    )

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }

    log.Println("Database connected")
    return nil
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**Create `internal/models/user.go`:**

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Username  string         `json:"username" gorm:"unique;not null"`
    Email     string         `json:"email" gorm:"unique;not null"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
```

**Create database:**

```bash
podman exec -it postgres bash
psql -U postgres -c "CREATE DATABASE my_echo_db;"
exit
```

**Create `internal/handlers/user.go`:**

```go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/yourname/my-echo-app/internal/database"
    "github.com/yourname/my-echo-app/internal/models"
)

func GetUsers(c echo.Context) error {
    var users []models.User
    database.DB.Find(&users)
    return c.JSON(http.StatusOK, users)
}

func CreateUser(c echo.Context) error {
    user := new(models.User)
    if err := c.Bind(user); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    database.DB.Create(&user)
    return c.JSON(http.StatusCreated, user)
}

func GetUser(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var user models.User

    if err := database.DB.First(&user, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
    }

    return c.JSON(http.StatusOK, user)
}

func UpdateUser(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var user models.User

    if err := database.DB.First(&user, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
    }

    if err := c.Bind(&user); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    database.DB.Save(&user)
    return c.JSON(http.StatusOK, user)
}

func DeleteUser(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var user models.User

    if err := database.DB.Delete(&user, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
    }

    return c.JSON(http.StatusOK, map[string]string{"message": "User deleted"})
}
```

**Update `main.go`:**

```go
package main

import (
    "log"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/yourname/my-echo-app/internal/database"
    "github.com/yourname/my-echo-app/internal/models"
    "github.com/yourname/my-echo-app/internal/handlers"
)

func main() {
    // Connect to database
    if err := database.Connect(); err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto-migrate
    database.DB.AutoMigrate(&models.User{})

    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // API routes
    e.GET("/api/users", handlers.GetUsers)
    e.POST("/api/users", handlers.CreateUser)
    e.GET("/api/users/:id", handlers.GetUser)
    e.PUT("/api/users/:id", handlers.UpdateUser)
    e.DELETE("/api/users/:id", handlers.DeleteUser)

    e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}
```

## 7. Validation

**Install validator:**

```bash
go get github.com/go-playground/validator/v10
```

**Add validation to models:**

```go
type User struct {
    ID       uint   `json:"id" gorm:"primaryKey"`
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
}
```

**Create validator middleware:**

```go
type CustomValidator struct {
    validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

// In main.go
e.Validator = &CustomValidator{validator: validator.New()}
```

## 8. Middleware

**Custom middleware:** `internal/middleware/auth.go`

```go
package middleware

import (
    "github.com/labstack/echo/v4"
)

func RequireAuth() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := c.Request().Header.Get("Authorization")
            if token == "" {
                return echo.ErrUnauthorized
            }
            // Validate token...
            return next(c)
        }
    }
}
```

**Use middleware:**

```go
api := e.Group("/api")
api.Use(middleware.RequireAuth())
```

## 9. Debugging

**Start with Delve:**

```bash
dlv debug ./cmd/server/main.go --headless --listen=0.0.0.0:2345 --api-version=2
```

**GoLand debugger:**

1. Run → Edit Configurations → "+" → Go Remote
2. Host: `localhost`, Port: `8402`
3. Set breakpoints, click "Debug"

## 10. Testing

**Create `internal/handlers/user_test.go`:**

```go
package handlers

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    if assert.NoError(t, GetUsers(c)) {
        assert.Equal(t, http.StatusOK, rec.Code)
    }
}

func TestCreateUser(t *testing.T) {
    e := echo.New()
    userJSON := `{"username":"test","email":"test@example.com"}`
    req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(userJSON))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    if assert.NoError(t, CreateUser(c)) {
        assert.Equal(t, http.StatusCreated, rec.Code)
    }
}
```

**Run tests:**

```bash
go test ./...
go test -v -cover ./internal/handlers
```

## 11. Error Handling

**Custom error handler:**

```go
func customHTTPErrorHandler(err error, c echo.Context) {
    code := http.StatusInternalServerError
    message := "Internal Server Error"

    if he, ok := err.(*echo.HTTPError); ok {
        code = he.Code
        message = he.Message.(string)
    }

    c.JSON(code, map[string]interface{}{
        "error": map[string]interface{}{
            "code":    code,
            "message": message,
        },
    })
}

// In main.go
e.HTTPErrorHandler = customHTTPErrorHandler
```

## 12. Request Binding and Validation

**Struct binding:**

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
}

func CreateUser(c echo.Context) error {
    req := new(CreateUserRequest)
    if err := c.Bind(req); err != nil {
        return err
    }
    if err := c.Validate(req); err != nil {
        return err
    }
    // Process request...
}
```

## 13. Swagger Documentation

**Install:**

```bash
go get -u github.com/swaggo/echo-swagger
go get -u github.com/swaggo/swag/cmd/swag
```

**Add annotations:**

```go
// GetUsers godoc
// @Summary Get all users
// @Description Get list of users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Router /api/users [get]
func GetUsers(c echo.Context) error {
    // ...
}
```

**Generate docs and add route:**

```bash
swag init -g cmd/server/main.go
```

```go
import echoSwagger "github.com/swaggo/echo-swagger"

e.GET("/swagger/*", echoSwagger.WrapHandler)
```

## 14. Configure nginx-proxy-manager

1. Open http://localhost:81
2. Add Proxy Host:
   - Domain: `my-echo-app.test`
   - Forward to: `go:8080`
3. SSL: Request certificate
4. Click "Save"

**Update /etc/hosts:**
```bash
sudo sh -c 'echo "127.0.0.1 my-echo-app.test" >> /etc/hosts'
```

## 15. Graceful Shutdown

```go
import (
    "context"
    "os"
    "os/signal"
    "time"
)

func main() {
    e := echo.New()
    // ... setup

    go func() {
        if err := e.Start("0.0.0.0:8080"); err != nil && err != http.ErrServerClosed {
            e.Logger.Fatal("shutting down the server")
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := e.Shutdown(ctx); err != nil {
        e.Logger.Fatal(err)
    }
}
```

## Port Allocation

- **8400**: Echo application
- **8401**: Metrics
- **8402**: Delve debugger
- **8403**: pprof

## Troubleshooting

**Issue:** Template not found
- Verify path in `template.ParseGlob("templates/*.html")`
- Check templates directory exists
- Use absolute path if needed

**Issue:** Static files not serving
- Check `e.Static("/static", "public")`
- Verify public/ directory exists
- Ensure working directory correct

**Issue:** Database connection fails
- Use container name (`postgres`)
- Verify credentials in environment
- Check postgres container running

**Issue:** Middleware not executing
- Verify middleware registered before routes
- Check middleware returns `next(c)`
- Use `e.Use()` for global middleware

## Next Steps

- Add JWT authentication with `github.com/golang-jwt/jwt`
- Implement rate limiting
- Add Prometheus metrics
- Setup request ID middleware
- Configure CORS properly
- Add request/response logging
- Implement API versioning
- Setup database migrations with golang-migrate
