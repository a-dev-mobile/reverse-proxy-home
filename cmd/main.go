package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/a-dev-mobile/reverse-proxy-home/internal/config"
	"github.com/a-dev-mobile/reverse-proxy-home/internal/logging"

	"golang.org/x/exp/slog"
)


func main() {

	cfg, lg := initializeApp()

	// Мультиплексор для маршрутизации
	mux := http.NewServeMux()
	// Обработчики для каждого домена
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		// Удаление порта из хоста, если он есть
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		switch host {
		// https://wayofdt.com:8444/
		// http://wayofdt.com:8090/
		case "wayofdt.com":
			handleMainDomain(w, r)
		// https://subdomain1.wayofdt.com:8444/
		// http://subdomain1.wayofdt.com:8090/
		case "subdomain1.wayofdt.com":
			handleSubdomain1(w, r)
		// https://subdomain2.wayofdt.com:8444/
		// http://subdomain2.wayofdt.com:8090/
		case "subdomain2.wayofdt.com":
			handleSubdomain2(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	// HTTP-сервер
	go func() {
		log.Println("Starting HTTP server on :8090")
		if err := http.ListenAndServe(":8090", mux); err != nil {
			log.Fatal("HTTP server failed: ", err)
		}
	}()

	// HTTPS-сервер
	log.Println("Starting HTTPS server on :8444")
	if err := http.ListenAndServeTLS(":8444", `c:\1\wayofdt.com.crt`, `c:\1\wayofdt.com.key`, mux); err != nil {
		log.Fatal("HTTPS server failed: ", err)
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

	cfg, err := config.LoadConfig("../config","config.yaml")

	if err != nil {
		log.Fatalf("Error loading config: %s", err)

	}

	return cfg
}
