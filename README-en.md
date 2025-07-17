# 🧠 Caching Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Caching HTTP proxy server written in Go**
A simple caching proxy server that redirects requests to a remote server and caches the results.

> Perfectly suitable for reducing the load on external APIs by caching responses with flexible configuration via YAML.

---

## 📌 Features

- Transparent caching of HTTP requests
- Redis support as a cache storage
- Flexible configuration through a YAML file
- Ability to specify blacklisted URL paths not to be cached
- Cache clearing via a secret key

---

## ⚙️ Installation and Launch

### 1. Local Run

```bash
go run cmd/main.go
```

Make sure the `config.yaml` configuration file is located in the root directory of the project and has the correct settings.

### 2. Running via Docker

Build the image (if needed):

```bash
docker build -t caching-proxy .
```

Run the container:

```bash
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml caching-proxy
```

---

## 📄 Configuration

Configuration is done via the `config.yaml` file.

### Example configuration:

```yaml
server:
  origin: https://httpbin.org/     # Base URL to redirect requests to
  # port: 1323                    # Port (default: 8080)
  secret: pls_delete_cache_maboy  # Secret for cache clearing

redis:
  addr: localhost                 # Redis address
  port: 6379                      # Redis port
  password:                       # Password (if used)
  db:                             # Database number
  TTL: 5                          # Cache lifetime in minutes

blacklist:
  - /uuid                         # These paths will not be cached
  - /delay/(.+)                   # Supports regex

logger:
  level: DEBUG                     # Currently supports only DEBUG, INFO, ERROR
  file: app.log
```

---

## 🧪 Usage

After launch, the service is available at:

```
http://localhost:8080/
```

All requests are redirected to the specified `origin` (`https://httpbin.org/` in this case), and the results are cached.

### Example usage:

```bash
curl http://localhost:8080/ip
```

### Clearing the cache

To clear the cache, send a POST request with the secret key:

```bash
curl -H "Authorization: Bearer pls_delete_cache_maboy" -X POST http://localhost:8080/clear
```

---

## 📁 Blacklist Format

Simple strings and regular expressions are supported:

```yaml
blacklist:
  - /uuid
  - /delay/(.+)
```

---

## 🏷 Caching Status (X-Cache)

Each response includes an `X-Cache` header indicating the caching status:

| Value     | Description                                      |
|-----------|--------------------------------------------------|
| `MISS`    | Data was not cached, request processed directly  |
| `HIT`     | Response retrieved from cache                    |
| `BYPASS`  | Request bypassed caching (via blacklist)         |

---

## 📦 Technologies

- Golang
- Redis
- YAML for configuration
- Docker

---

## 🧾 License

MIT License — see [LICENSE](LICENSE) for details.
