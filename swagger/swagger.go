package swagger

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// Config defines the config for middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool

	// BasePath for the UI path
	//
	// Optional. Default: /
	BasePath string

	// FilePath for the swagger.json or swagger.yaml file
	//
	// Optional. Default: ./swagger.json
	FilePath string

	// Path combines with BasePath for the full UI path
	//
	// Optional. Default: docs
	Path string

	// Title for the documentation site
	//
	// Optional. Default: Fiber API documentation
	Title string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	Next:     nil,
	BasePath: "/",
	FilePath: "./swagger.json",
	Path:     "docs",
	Title:    "Fiber API documentation",
}

// New creates a new middleware handler
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := ConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]

		// Set default values
		if len(cfg.BasePath) == 0 {
			cfg.BasePath = ConfigDefault.BasePath
		}
		if len(cfg.FilePath) == 0 {
			cfg.FilePath = ConfigDefault.FilePath
		}
		if len(cfg.Path) == 0 {
			cfg.Path = ConfigDefault.Path
		}
		if len(cfg.Title) == 0 {
			cfg.Title = ConfigDefault.Title
		}
	}

	// Verify Swagger file exists
	if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
		panic(errors.New(fmt.Sprintf("%s file does not exist", cfg.FilePath)))
	}

	// Read Swagger Spec into memory
	rawSpec, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		log.Fatalf("Failed to read Swagger YAML file: %v", err)
		panic(err)
	}

	// Generate URL path's for the middleware
	specURL, err := url.JoinPath(cfg.BasePath, cfg.FilePath)
	if err != nil {
		log.Fatalf("Failed to join URL path between %s and %s", cfg.BasePath, cfg.FilePath)
		panic(err)
	}
	swaggerUIPath, err := url.JoinPath(cfg.BasePath, cfg.Path)
	if err != nil {
		log.Fatalf("UnaFailedble to join URL between %s and %s", cfg.BasePath, cfg.Path)
		panic(err)
	}

	// Serve the Swagger spec from memory
	swaggerSpecHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, specURL) || strings.HasSuffix(r.URL.Path, specURL) {
			w.Header().Set("Content-Type", "application/yaml")
			w.Write(rawSpec)
		} else if strings.HasSuffix(r.URL.Path, specURL) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(rawSpec)
		} else {
			fmt.Printf("Somehow here? %s\n", r.URL.Path)
			http.NotFound(w, r)
		}
	})

	// Define UI Options
	swaggerUIOpts := middleware.SwaggerUIOpts{
		BasePath: cfg.BasePath,
		SpecURL:  specURL,
		Path:     cfg.Path,
		Title:    cfg.Title,
	}

	// Create UI middleware
	middlewareHandler := adaptor.HTTPHandler(middleware.SwaggerUI(swaggerUIOpts, swaggerSpecHandler))

	// Return new handler
	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Only respond to requests to this middleware
		if !(c.Path() == swaggerUIPath || c.Path() == specURL) {
			fmt.Println("-----")
			fmt.Printf("c.Path() is %s\n", c.Path())
			fmt.Printf("BasePath is %s\n", cfg.BasePath)
			fmt.Printf("swaggerUIPath is %s\n", swaggerUIPath)
			fmt.Printf("specURL is %s\n", specURL)

			return c.Next()
		} else {
			fmt.Println("+++++")
			fmt.Printf("c.Path() is %s\n", c.Path())
			fmt.Printf("BasePath is %s\n", cfg.BasePath)
			fmt.Printf("swaggerUIPath is %s\n", swaggerUIPath)
			fmt.Printf("specURL is %s\n", specURL)

		}

		// Pass Fiber context to handler
		middlewareHandler(c)
		return nil
	}
}
