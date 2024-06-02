package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
	"strconv"
	"strings"
)

type Middleware struct {
	Limiter        *limiter.Limiter
	OnError        mgin.ErrorHandler
	OnLimitReached mgin.LimitReachedHandler
	KeyGetter      mgin.KeyGetter
	ExcludedKey    func(string) bool
}

func NewMiddleware(limiter *limiter.Limiter) gin.HandlerFunc {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        mgin.DefaultErrorHandler,
		OnLimitReached: mgin.DefaultLimitReachedHandler,
		KeyGetter:      DefaultKeyGetter(limiter),
		ExcludedKey:    nil,
	}
	return func(ctx *gin.Context) {
		middleware.Handle(ctx)
	}
}

// Handle gin request.
func (middleware *Middleware) Handle(c *gin.Context) {
	key := middleware.KeyGetter(c)
	if middleware.ExcludedKey != nil && middleware.ExcludedKey(key) {
		c.Next()
		return
	}

	context, err := middleware.Limiter.Get(c, key)
	if err != nil {
		middleware.OnError(c, err)
		c.Abort()
		return
	}

	c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

	if context.Reached {
		middleware.OnLimitReached(c)
		c.Abort()
		return
	}
	c.Next()
}

// DefaultKeyGetter is the default KeyGetter used by a new Middleware.
// It returns the Client IP address.
func DefaultKeyGetter(limiter *limiter.Limiter) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		key := c.GetHeader(limiter.Options.ClientIPHeader)
		if key != "" {
			return key
		}
		return c.ClientIP()
	}
}

func NewRateLimiterMiddleware(
	formattedRate string,
	redisAddress string,
	prefix string,
	clientHeader string,
) (gin.HandlerFunc, error) {
	// See: https://github.com/ulule/limiter-examples/blob/master/gin/main.go
	// Use the simplified format "<limit>-<period>"", with the given
	// periods:
	// * "S": second
	// * "M": minute
	// * "H": hour
	// * "D": day
	//
	// Examples:
	// * 5 reqs/second: "5-S"
	// * 10 reqs/minute: "10-M"
	// * 1000 reqs/hour: "1000-H"
	// * 2000 reqs/day: "2000-D"
	//
	// Usage:
	// router := gin.Default()
	// router.ForwardedByClientIP = true
	// router.Use(middleware)
	// router.GET("/", index)

	// Define a limit rate
	rate, err := limiter.NewRateFromFormatted(formattedRate)
	if err != nil {
		return nil, err
	}

	// Create a redis client
	client, err := NewRedisClient(redisAddress)
	if err != nil {
		return nil, err
	}

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	// Create a new middleware with the limiter instance.
	if clientHeader == "" {
		middleware := NewMiddleware(limiter.New(store, rate))
		return middleware, nil
	} else {
		middleware := NewMiddleware(limiter.New(store, rate, limiter.WithClientIPHeader(clientHeader)))
		return middleware, nil
	}
}

func GetTrueClientIP(c *gin.Context) string {
	// Get client IP from X-Forwarded-For
	// https://cloud.google.com/load-balancing/docs/https#x-forwarded-for_header
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor == "" {
		return c.ClientIP()
	}
	IPs := strings.Split(forwardedFor, ",")
	numIPs := len(IPs)
	if numIPs < 2 {
		log.Warn().Msgf("less than two IPs in X-Forwarded-For: %v", IPs)
		return c.ClientIP()
	}
	clientIP := strings.TrimSpace(IPs[numIPs-2])
	return clientIP
}
