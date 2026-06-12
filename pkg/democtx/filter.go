package democtx

import "github.com/gin-gonic/gin"

// FromGin returns whether the request is in demo mode.
func FromGin(c *gin.Context) bool {
	v, ok := c.Get("demo_mode")
	if !ok {
		return false
	}
	isDemo, _ := v.(bool)
	return isDemo
}
