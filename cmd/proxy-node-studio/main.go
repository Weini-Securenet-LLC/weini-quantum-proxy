package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"hermes-agent/proxy-node-studio/internal/proxynode"
)

//go:embed web/*
var webFS embed.FS

type fetchRequest struct {
	URL       string   `json:"url"`
	Protocols []string `json:"protocols"`
	Timeout   float64  `json:"timeout"`
}

func main() {
	port := flag.Int("port", 0, "local UI port (0 = random)")
	noBrowser := flag.Bool("no-browser", false, "do not auto-open the browser")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/web/", http.StatusTemporaryRedirect)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/fetch", handleFetch)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "default_url": proxynode.DefaultURL, "protocols": proxynode.SupportedProtocols})
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	uiURL := fmt.Sprintf("http://%s/web/", listener.Addr().String())
	webSub, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.FS(webSub))))
	server := &http.Server{Handler: logging(mux), ReadHeaderTimeout: 5 * time.Second}

	fmt.Println("Proxy Node Studio UI:", uiURL)
	if !*noBrowser {
		go openBrowser(uiURL)
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func handleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var req fetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(req.URL) == "" {
		req.URL = proxynode.DefaultURL
	}
	if req.Timeout <= 0 {
		req.Timeout = 20
	}
	if len(req.Protocols) == 0 {
		req.Protocols = proxynode.SupportedProtocols
	}
	out, err := proxynode.FetchAndNormalize(req.URL, req.Timeout, req.Protocols)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
