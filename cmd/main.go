package main

import (
	"fmt"
	"log"
	"time"

	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/a-dev-mobile/reverse-proxy-home/internal/config"
	"github.com/a-dev-mobile/reverse-proxy-home/internal/logging"

	"golang.org/x/exp/slog"
)

func main() {
	cfg, logger := initializeApp()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(logger, cfg, w, r)
	})

	startHTTPServer(cfg, mux, logger)
	startHTTPSServer(cfg, mux, logger)
}

func handleRequest(logger *slog.Logger, cfg *config.Config, w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	host := strings.Split(r.Host, ":")[0]
	logger.Info("Request received", "host", host, "path", r.URL.Path, "method", r.Method)

	if targetURL, ok := cfg.ProxyConfig.Redirects[host]; ok {
		logger.Info("Redirect found", "targetURL", targetURL) // Добавлено для диагностики
		proxyURL, err := url.Parse(targetURL)
		if err != nil {
			logger.Error("Failed to parse proxy URL", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("Proxy error", "error", err, "status_code", http.StatusBadGateway)
			http.Error(w, "Proxy Error", http.StatusBadGateway)
		}
		proxy.ServeHTTP(w, r)
		logger.Info("Proxy request", "host", host, "path", r.URL.Path, "status_code", http.StatusOK, "duration_ms", time.Since(startTime).Milliseconds())
		return
	}

	http.NotFound(w, r)
	logger.Info("Request not found", "host", host, "path", r.URL.Path, "status_code", http.StatusNotFound, "duration_ms", time.Since(startTime).Milliseconds())
}
func startHTTPServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) {
	go func() {
		portStr := fmt.Sprintf(":%d", cfg.ProxyConfig.HTTPPort)
		logger.Info("Starting HTTP server", "port", portStr)
		if err := http.ListenAndServe(portStr, handler); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()
}

func startHTTPSServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) {
	portStr := fmt.Sprintf(":%d", cfg.ProxyConfig.HTTPSPort)
	logger.Info("Starting HTTPS server", "port", portStr)
	if err := http.ListenAndServeTLS(portStr, cfg.ProxyConfig.CertFile, cfg.ProxyConfig.KeyFile, handler); err != nil {
		logger.Error("HTTPS server failed", "error", err)
	}
}

func initializeApp() (*config.Config, *slog.Logger) {

	cfg := getConfigOrFail()

	lg := logging.SetupLogger(cfg)

	return cfg, lg
}

func getConfigOrFail() *config.Config {

	cfg, err := config.LoadConfig("../config", "config.yaml")

	if err != nil {
		log.Fatalf("Error loading config: %s", err)

	}

	return cfg
}
