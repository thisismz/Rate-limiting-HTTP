package main

import (
	ratelimit "RateLimiter/src"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	forwardedFor = "X-Forwarded-For"
)

type handlerWrapper struct {
	handler http.Handler
}

func (h *handlerWrapper) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	h.handler.ServeHTTP(writer, request)
}

var config = ratelimit.RateLimiterConfig{}

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()
	config = ratelimit.RateLimiterConfig{
		Extractor:   ratelimit.NewHTTPHeadersExtractor(forwardedFor),
		Strategy:    ratelimit.NewSortedSetCounterStrategy(client, time.Now().Local),
		Expiration:  time.Minute,
		MaxRequests: 10,
	}
}
func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	config = ratelimit.RateLimiterConfig{
		Extractor:   ratelimit.NewHTTPHeadersExtractor(forwardedFor),
		Strategy:    ratelimit.NewSortedSetCounterStrategy(client, time.Now().Local),
		Expiration:  time.Minute,
		MaxRequests: 10,
	}

	http.HandleFunc("/headers", headers)
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8090", limiter(http.DefaultServeMux))
}
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}
func limiter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := ratelimit.NewHTTPRateLimiterHandler(&handlerWrapper{handler: h}, &config)
		wrapper.ServeHTTP(w, r)
	})
}
