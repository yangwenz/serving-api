package api

import (
	"errors"
	"fmt"
	"github.com/HyperGAI/serving-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	userIDKey = "UID"
)

func authenticateRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userHeader := ctx.GetHeader(userIDKey)

		if len(userHeader) == 0 {
			err := errors.New("user-id header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(userHeader)
		if len(fields) != 1 {
			err := errors.New("invalid user-id header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.Request.Header.Set("UID", fields[0])
		ctx.Next()
	}
}

func traceRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		beforeRequest(ctx)
		// Call the next middleware or endpoint handler
		ctx.Next()
		// Do some request tracing after request is processed
		afterRequest(ctx)
	}
}

func beforeRequest(ctx *gin.Context) {
	start := time.Now()
	// Log the request start time
	userID := ctx.Request.Header.Get("UID")
	log.Info().Msgf("user-id %s, started %s %s",
		userID, ctx.Request.Method, ctx.Request.URL.Path)
	// Add start time to the request context
	ctx.Set("startTime", start)
}

func afterRequest(ctx *gin.Context) {
	// Get the start time from the request context
	startTime, exists := ctx.Get("startTime")
	if !exists {
		startTime = time.Now()
	}
	duration := time.Since(startTime.(time.Time))
	// Log the request completion time and duration
	userID := ctx.Request.Header.Get("UID")
	log.Info().Msgf("user-id %s, completed %s %s in %v",
		userID, ctx.Request.Method, ctx.Request.URL.Path, duration)
}

func setTrueClientIP() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := utils.GetTrueClientIP(ctx)
		ctx.Header("True-Client-IP", ip)
		ctx.Request.Header.Set("True-Client-IP", ip)
		log.Info().Msgf("X-Forwarded-For: %s", ctx.GetHeader("X-Forwarded-For"))
		log.Info().Msgf("client IP: %s", ip)
		ctx.Next()
	}
}

func rateLimitByUser(config utils.Config, formattedRate string, prefix string) gin.HandlerFunc {
	if config.RedisAddress != "" {
		middleware, err := utils.NewRateLimiterMiddleware(
			formattedRate,
			config.RedisAddress,
			fmt.Sprintf("rate_limiter_%s", prefix),
			"UID",
		)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create rate-limit middleware")
		}
		return middleware
	} else {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.FullPath()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		// Call the next middleware or endpoint handler
		ctx.Next()
		// Update metrics
		totalRequests.WithLabelValues(path).Inc()
		responseStatus.WithLabelValues(path, strconv.Itoa(ctx.Writer.Status())).Inc()
		timer.ObserveDuration()
	}
}
