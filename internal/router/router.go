// Package router wires up all route groups and middleware onto a single
// gin.Engine. Adding a new endpoint group should only require a new
// register* function and a call to it from New().
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/riedl/wishlist-api-service/internal/handlers"
	"github.com/riedl/wishlist-api-service/internal/httputil"
	"github.com/riedl/wishlist-api-service/internal/middleware"
)

// New builds and returns a fully configured gin.Engine, including global
// middleware, health checks, and versioned API route groups.
func New() *gin.Engine {
	engine := gin.New()

	engine.Use(
		middleware.RequestID(),
		middleware.Recovery(),
		middleware.Logger(),
		middleware.CORS(),
	)

	engine.NoRoute(func(c *gin.Context) {
		httputil.NotFound(c, "resource not found")
	})

	engine.GET("/health", handlers.Health)

	v1 := engine.Group("/api/v1")
	registerItemRoutes(v1)
	registerWishlistRoutes(v1)

	return engine
}

// registerItemRoutes mounts the /items resource. This is the pattern to
// follow when adding new resources: construct a handler, group its routes,
// and call the register function from New().
func registerItemRoutes(rg *gin.RouterGroup) {
	h := handlers.NewItemsHandler()

	items := rg.Group("/items")
	items.GET("", h.List)
	items.POST("", h.Create)
	items.GET("/:id", h.Get)
	items.PUT("/:id", h.Update)
	items.DELETE("/:id", h.Delete)
}

// registerWishlistRoutes mounts the /wishlists resource.
func registerWishlistRoutes(rg *gin.RouterGroup) {
	h := handlers.NewWishlistsHandler()

	wishlists := rg.Group("/wishlists")
	wishlists.GET("/public", h.ListPublicForUser)
	wishlists.POST("", h.Create)
	wishlists.DELETE("/:id", h.Delete)
}
