package middleware

import (
	"github.com/gin-gonic/gin"
)

const (
	HeaderMarketPayMode = "X-MarketPay-Mode"
	ContextDemoModeKey  = "demo_mode"
)

// DemoMode reads X-MarketPay-Mode (demo|live) and stores bool in context.
// Default is live mode (false) when header absent.
func DemoMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		mode := c.GetHeader(HeaderMarketPayMode)
		isDemo := mode == "demo"
		c.Set(ContextDemoModeKey, isDemo)
		c.Next()
	}
}

// IsDemoMode returns whether the request is in demo mode.
func IsDemoMode(c *gin.Context) bool {
	v, ok := c.Get(ContextDemoModeKey)
	if !ok {
		return false
	}
	isDemo, _ := v.(bool)
	return isDemo
}
