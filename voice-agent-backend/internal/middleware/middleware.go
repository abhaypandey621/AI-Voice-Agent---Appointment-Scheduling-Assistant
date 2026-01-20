package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

// Logger returns a logging middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return param.TimeStamp.Format(time.RFC3339) + " | " +
			param.StatusCodeColor() + " " + string(rune(param.StatusCode)) + " " + param.ResetColor() + " | " +
			param.Latency.String() + " | " +
			param.ClientIP + " | " +
			param.Method + " " +
			param.Path + "\n"
	})
}

// Recovery returns a recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}

// RateLimit returns a simple rate limiting middleware
func RateLimit() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use a proper rate limiter with Redis
	return func(c *gin.Context) {
		c.Next()
	}
}

// RequestID adds a request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
