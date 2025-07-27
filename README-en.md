# 🧠 Caching Proxy

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Caching HTTP proxy server written in Go**
A simple caching proxy server that forwards requests to a remote server and caches the result.

> It's perfect for reducing the load on external APIs by caching responses with the ability for flexible configuration via a YAML file.

---

## 📌 Features

*   Transparent caching of HTTP requests
*   Support for Redis as a cache store
*   Flexible configuration via a YAML file
*   Ability to specify a blacklist of URL paths that should not be cached
*   Cache clearing via a secret key

## ⚙️ Installation and Running

### 1. Local Run

```bash
go run cmd/main.go
```

Make sure the configuration file `config.yaml` is located in the project root and is properly configured.

### 2. Running via Docker

Build the image (if needed):

```bash
docker build -t caching-proxy .
```

Run the container:

```bash
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/app.log:/app/app.log caching-proxy
```

### 3. Running via Docker Compose

You can use the example `docker-compose` from this repository:

```bash
docker-compose up --build
```

## 📄 Configuration

Configuration is done via the `config.yaml` file.

**Example configuration:**

```yaml
server:
  origin: https://httpbin.org/ # Original URL
  # port: 1323
  secret: pls_delete_cache_maboy # Secret for cache clearing
  timeout: 10 # Timeout to origin, in seconds
  ratelimit: # Rate limiting is performed per IP
    rate: 20
    duration: 60 # in seconds

redis:
  addr: redis                    # Redis address
  port: 6379                     # Redis port
  password:                      # Password (if used)
  db:                            # Database number
  TTL: 5                         # Cache Time To Live in minutes

blacklist:
  - /uuid                        # These paths will not be cached
  - /delay/(.+)                  # Supports regexp

logger:
  level: DEBUG                   # Currently supports only DEBUG, INFO, ERROR
  file: app.log                  # Leave empty for stdout output
```

## 🧪 Usage

After starting, the service is available at:

`http://localhost:8080/`

All requests are forwarded to the specified `origin` (`https://httpbin.org/` in this case), and the results are cached.

**Example usage:**

```bash
curl http://localhost:8080/ip
```

### Cache Clearing

To clear the cache, send a POST request with the secret key:

```bash
curl -H "Authorization: Bearer pls_delete_cache_maboy" -X POST http://localhost:8080/clear
```

## 📁 Blacklist Format

Simple strings and regular expressions are supported:

```yaml
blacklist:
  - /uuid
  - /delay/(.+)
```

## 🏷 Cache Status (`X-Cache`)

A `X-Cache` header is added to every response, indicating the caching status:

| Value   | Description                                        |
| :------ | :------------------------------------------------- |
| `MISS`  | Data was not cached, request executed directly     |
| `HIT`   | Response retrieved from cache                      |
| `BYPASS`| Request was excluded from caching (via blacklist)  |

## 📦 Technologies

*   Golang
*   Redis
*   YAML for configuration
*   Docker

## 🧾 License

MIT License – see [LICENSE](LICENSE) for details.

---
