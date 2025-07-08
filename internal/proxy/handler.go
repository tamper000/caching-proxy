package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"
)

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.RequestURI
	key := p.Config.Origin + ":" + method + ":" + path

	if value, err := p.Redis.GetCache(key); err == nil {
		buffer := bytes.NewBuffer(value)
		reader := bufio.NewReader(buffer)
		resp, err := http.ReadResponse(reader, nil)
		if err == nil {
			sendFinal(w, resp, "HIT")
			return
		}
	}

	req, err := http.NewRequest(method, p.Config.Origin+path, r.Body)
	if err != nil {
		http.Error(w, "Error forward request", http.StatusInternalServerError)
		return
	}
	for key, value := range r.Header {
		req.Header.Set(key, value[0])
	}

	resp, err := p.HttpClient.Do(req)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	respBytes, err := httputil.DumpResponse(resp, true)
	if err == nil {
		err := p.Redis.SetCache(key, respBytes)
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

func (p *Proxy) ClearHandler(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	secret := values.Get("secret")

	if p.Config.Secret != secret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := p.Redis.Client.FlushDB(ctx).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
