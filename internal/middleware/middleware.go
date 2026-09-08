// Package middleware contains reusable Gin middleware applied to all
// (or subsets of) routes: request IDs, structured logging, panic recovery,
// and CORS handling.
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/riedl/wishlist-api-service/internal/httputil"
)

// RequestIDHeader is the HTTP header used to propagate/return the request ID.
const RequestIDHeader = "X-Request-Id"

// RequestIDKey is the gin.Context key the request ID is stored under.
const RequestIDKey = "request_id"

// RequestID assigns a unique ID to every request, reusing an incoming
// X-Request-Id header if the client already supplied one. The ID is stored
// in the context and echoed back in the response headers, which makes it
// easy to correlate logs across services.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// Logger logs each request's method, path, status, latency, and request ID.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get(RequestIDKey)

		log.Printf("[%s] %s %s %d %s", requestID, c.Request.Method, path, status, latency)
	}
}

// Recovery converts panics into a consistent 500 JSON error response instead
// of letting Gin's default recovery write a plain-text body.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				httputil.AbortWithError(c, http.StatusInternalServerError, "internal_error", "an unexpected error occurred")
			}
		}()
		c.Next()
	}
}

// CORS returns permissive CORS middleware suitable for development. Adjust
// AllowOrigins/AllowHeaders for production usage.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+RequestIDHeader)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
