package docs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"sigs.k8s.io/yaml"

	"go.mau.fi/whatsmeow/api/internal/version"
)

//go:embed openapi.yaml
var specYAML []byte

//go:embed openapi.json
var specJSONRaw []byte

var (
	jsonOnce sync.Once
	specJSON []byte
	jsonErr  error
)

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi" yaml:"openapi"`
	Info       map[string]interface{} `json:"info" yaml:"info"`
	Servers    []Server               `json:"servers" yaml:"servers"`
	Paths      map[string]interface{} `json:"paths" yaml:"paths"`
	Components map[string]interface{} `json:"components" yaml:"components"`
}

type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
}

type Config struct {
	BaseURL string
}

func generateServers(baseURL string) []Server {
	servers := []Server{}

	servers = append(servers, Server{
		URL:         baseURL,
		Description: "API Server",
	})

	if !strings.Contains(baseURL, "localhost") && !strings.Contains(baseURL, "127.0.0.1") {
		servers = append(servers, Server{
			URL:         "http://localhost:8080",
			Description: "Local development",
		})
	}

	return servers
}

func generateDynamicSpec(baseURL string) ([]byte, []byte, error) {
	var spec OpenAPISpec
	if err := yaml.Unmarshal(specYAML, &spec); err != nil {
		return nil, nil, fmt.Errorf("failed to parse base spec: %w", err)
	}

	// Inject dynamic version from VERSION file
	if spec.Info != nil {
		spec.Info["version"] = version.String()
	}

	spec.Servers = generateServers(baseURL)

	yamlBytes, err := yaml.Marshal(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	jsonBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return yamlBytes, jsonBytes, nil
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <title>Zé da API Docs</title>
    <link rel="icon" type="image/x-icon" href="/favicon.ico">
    <link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png">
    <link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png">
    <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png">
    <link rel="manifest" href="/manifest.json">
    <meta name="theme-color" content="#4593c6">
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css"/>
    <style>
      body { margin: 0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/docs/openapi.json',
          dom_id: '#swagger-ui',
          presets: [SwaggerUIBundle.presets.apis],
          layout: 'BaseLayout'
        });
      };
    </script>
  </body>
</html>`

func YAMLHandler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := cfg.BaseURL
		if baseURL == "" {
			scheme := "http"
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			host := r.Host
			baseURL = fmt.Sprintf("%s://%s", scheme, host)
		}

		yamlBytes, _, err := generateDynamicSpec(baseURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate spec: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(yamlBytes)
	})
}

func JSONHandler(cfg Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseURL := cfg.BaseURL
		if baseURL == "" {
			scheme := "http"
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			host := r.Host
			baseURL = fmt.Sprintf("%s://%s", scheme, host)
		}

		_, jsonBytes, err := generateDynamicSpec(baseURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to generate spec: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jsonBytes)
	})
}

func UIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerUIHTML))
	})
}

func LegacyYAMLHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(specYAML)
	})
}

func LegacyJSONHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonOnce.Do(func() {
			if len(specJSONRaw) > 0 {
				specJSON = specJSONRaw
				return
			}
			specJSON, jsonErr = yaml.YAMLToJSON(specYAML)
		})
		if jsonErr != nil {
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(specJSON)
	})
}
