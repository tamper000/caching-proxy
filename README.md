
# Caching proxy

Lightweight proxy server for proxying requests. Redis is used for caching requests

## Install

```bash
git clone https://github.com/tamper000/caching-proxy
cd caching-proxy
go build -o proxy cmd/caching-proxy/main.go
```

## Usage

### Start the server:

```bash
./proxy --origin "https://httpbin.org/" -p 1234
```
### To clear the cache::

```bash
./proxy --clear
```
