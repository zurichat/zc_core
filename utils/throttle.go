package utils

import (
	"errors"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// Map to hold the throttle rate limiters for each visitor
var zc_visitors = make(map[string]*rate.Limiter)
var mutex sync.Mutex

// Retrieve and return the rate limiter for the current visitor if it
// already exists. Else, create a new rate limiter and add it to the
// visitor map using the visitor's IP address as the key
func getVisitor(ip string) *rate.Limiter {
	mutex.Lock()
	defer mutex.Unlock()

	limiter, exists := zc_visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 2)
		zc_visitors[ip] = limiter
	}

	return limiter
}

// Throttling middleware
func throttle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the IP address for the current user
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			GetError(errors.New("error extracting ip address"), http.StatusInternalServerError, w)
			return
		}

		// Call the getVisitor function to retrieve the rate limiter for the
		// current user
		limiter := getVisitor(ip)
		if limiter.Allow() == false {
			GetError(err, http.StatusTooManyRequests, w)
			return
		}

		next.ServeHTTP(w, r)
	}

}
