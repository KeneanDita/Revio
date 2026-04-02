package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimit(rateStr string) gin.HandlerFunc {
	rate, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		panic("invalid rate format: " + err.Error())
	}
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	middleware := ginlimiter.NewMiddleware(instance)
	return middleware
}

func RateLimitStrict() gin.HandlerFunc {
	return RateLimit("10-M") // 10 requests per minute
}

func RateLimitNormal() gin.HandlerFunc {
	return RateLimit("60-M") // 60 requests per minute
}

func RateLimitAuth() gin.HandlerFunc {
	return RateLimit("5-M") // 5 login attempts per minute
}

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}
