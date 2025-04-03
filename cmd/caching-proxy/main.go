package main

import (
	"caching-proxy/internal/cache"
	"caching-proxy/internal/proxy"
	"flag"
	"fmt"
)

func main() {
	// Parsing flags
	port := flag.Int64("port", 8080, "caching proxy server port")
	origin := flag.String("origin", "", "root URL to cache")
	ttl := flag.Int64("ttl", 0, "cache ttl in minutes")
	rPort := flag.Int64("rport", 6379, "redis port")
	rAddr := flag.String("raddr", "localhost", "redis address")
	rPass := flag.String("rpass", "", "redis password")
	clear := flag.Bool("clear", false, "clear the cache")
	flag.Parse()

	if cache.ConnectDB(*rAddr, *rPass, *rPort) != nil {
		fmt.Println("Please set correct values for connecting to Redis.\n\n--rport <PORT> --raddr <ADDR> --rpass <PASSWORD>")
		return
	}

	if *clear {
		err := cache.ClearCache()
		if err != nil {
			fmt.Println("Error clearing cache:", err)
		} else {
			fmt.Println("Cache successfully reset")
		}

		return
	}

	if *origin != "" && *port > 0 {
		proxy := proxy.NewProxy(*origin, *ttl, *port)
		proxy.Start()
	} else {
		fmt.Println("None of the parameters are specified.\n\n--clear To clear the cache\n--origin <URL> --port <PORT> To start the server")
	}
}
