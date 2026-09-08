package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/riedl/wishlist-api-service/internal/httputil"
)

// Health handles GET /health and reports basic service liveness. It is a
// minimal example of using the httputil response helpers.
func Health(c *gin.Context) {
	httputil.OK(c, gin.H{"status": "ok"})
}
