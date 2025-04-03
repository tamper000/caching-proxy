package proxy

import (
	"bufio"
	"bytes"
	"caching-proxy/internal/cache"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type ProxyServer struct {
	Origin string
	TTL    int64
	Port   int64
}

func NewProxy(origin string, ttl int64, port int64) ProxyServer {
	origin = strings.TrimSuffix(origin, "/")
	return ProxyServer{
		Origin: origin,
		TTL:    ttl,
		Port:   port,
	}
}

func (p ProxyServer) Start() {
	http.HandleFunc("/", p.handleProxy)

	fmt.Printf("Startring on :%d port\n", p.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", p.Port), nil)
}

func (p ProxyServer) handleProxy(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.RequestURI
	key := p.Origin + ":" + method + ":" + path

	if value, err := cache.GetCache(key); err == nil {
		buffer := bytes.NewBuffer(value)
		reader := bufio.NewReader(buffer)
		resp, err := http.ReadResponse(reader, nil)
		if err == nil {
			sendFinal(w, resp, "HIT")
			return
		}
	}

	client := http.Client{}
	client.Timeout = 5 * time.Second
	req, err := http.NewRequest(method, p.Origin+path, r.Body)
	if err != nil {
		http.Error(w, "Error forward request", http.StatusInternalServerError)
		return
	}
	for key, value := range r.Header {
		req.Header.Set(key, value[0])
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err == nil {
		err := cache.SetCache(key, respBytes, p.TTL)
		if err != nil {
			fmt.Println("Error set cache:", err)
		}
	}

	sendFinal(w, resp, "MISS")
}

func sendFinal(w http.ResponseWriter, r *http.Response, header string) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-Cache", header)
	w.WriteHeader(r.StatusCode)
	for key, value := range r.Header {
		w.Header().Set(key, value[0])
	}
	w.Write(bytes)
}
