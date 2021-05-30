package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	TraceIDHeader = "X-B3-TraceId"
)

// LogMiddleware is the logging middleware
func LogMiddleware(logger *log.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process Request
		c.Next()

		// Stop timer
		duration := GetDurationInMillseconds(start)

		traceID := c.Request.Header.Get(TraceIDHeader)

		entry := logger.WithFields(log.Fields{
			"type":         "router",
			"client_ip":    GetClientIP(c),
			"duration(ms)": duration,
			"method":       c.Request.Method,
			"path":         c.Request.RequestURI,
			"status":       c.Writer.Status(),
			"referrer":     c.Request.Referer(),
			"traceID":      traceID,
		})

		if c.Writer.Status() >= 500 {
			entry.Error(c.Errors.String())
		} else {
			entry.Info("")
		}
	}
}

// GetClientIP gets the correct IP for the end client instead of the proxy
func GetClientIP(c *gin.Context) string {
	// first check the X-Forwarded-For header
	requester := c.Request.Header.Get("X-Forwarded-For")
	// if empty, check the Real-IP header
	if len(requester) == 0 {
		requester = c.Request.Header.Get("X-Real-IP")
	}
	// if the requester is still empty, use the hard-coded address from the socket
	if len(requester) == 0 {
		requester = c.Request.RemoteAddr
	}

	// if requester is a comma delimited list, take the first one
	// (this happens when proxied via elastic load balancer then again through nginx)
	if strings.Contains(requester, ",") {
		requester = strings.Split(requester, ",")[0]
	}

	return requester
}

// GetDurationInMillseconds takes a start time and returns a duration in milliseconds
func GetDurationInMillseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}
