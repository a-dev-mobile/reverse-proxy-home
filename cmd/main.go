package main

import (
	"log"

	"strings"
    "net/http"
    "net/http/httputil"
    "net/url"
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
    host := strings.Split(r.Host, ":")[0]
    logger.Info("Request received", "host", host, "path", r.URL.Path)

    // Проверка наличия конфигурации прокси для хоста
    if targetURL, ok := cfg.ProxyConfig.Redirects[host]; ok {
        proxyURL, err := url.Parse(targetURL)
        if err != nil {
            logger.Error("Failed to parse proxy URL", "error", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        proxy := httputil.NewSingleHostReverseProxy(proxyURL)
        proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
            logger.Error("Proxy error", "error", err)
            http.Error(w, "Proxy Error", http.StatusBadGateway)
        }
        proxy.ServeHTTP(w, r)
        return
    }
    // Обработка стандартных доменов
    switch host {
    case "wayofdt.com":
        handleMainDomain(w, r)
    case "subdomain1.wayofdt.com":
        handleSubdomain1(w, r)
    case "subdomain2.wayofdt.com":
        handleSubdomain2(w, r)
    default:
        http.NotFound(w, r)
    }
}
func startHTTPServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) {
	go func() {
		logger.Info("Starting HTTP server", slog.String("port", cfg.ProxyConfig.HTTPPort))
		if err := http.ListenAndServe(cfg.ProxyConfig.HTTPPort, handler); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()
}

func startHTTPSServer(cfg *config.Config, handler http.Handler, logger *slog.Logger) {
	logger.Info("Starting HTTPS server", slog.String("port", cfg.ProxyConfig.HTTPSPort))
	if err := http.ListenAndServeTLS(cfg.ProxyConfig.HTTPSPort, cfg.ProxyConfig.CertFile, cfg.ProxyConfig.KeyFile, handler); err != nil {
		logger.Error("HTTPS server failed", "error", err)
	}
}

func handleSubdomain1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is subdomain1.wayofdt.com"))
}

func handleSubdomain2(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is subdomain2.wayofdt.com"))
}

func handleMainDomain(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is wayofdt.com"))
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
