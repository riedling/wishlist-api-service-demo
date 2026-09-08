package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/riedl/wishlist-api-service/internal/httputil"
)

// Item is an example domain model representing a wishlist entry. Replace or
// extend this with real domain models as the service grows.
type Item struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

// createItemRequest is the validated payload for creating an item.
type createItemRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=200"`
	Price int    `json:"price" binding:"required,min=0"`
}

// updateItemRequest is the validated payload for updating an item.
type updateItemRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=200"`
	Price int    `json:"price" binding:"required,min=0"`
}

// itemURI binds the ":id" path parameter shared by single-item routes.
type itemURI struct {
	ID int `uri:"id" binding:"required"`
}

// ItemsHandler groups the handlers for the /items resource. It holds its
// own dependencies (here, an in-memory store) so it's easy to swap in a
// real repository/service layer without touching routing code.
type ItemsHandler struct {
	mu     sync.Mutex
	nextID int
	items  map[int]Item
}

// NewItemsHandler constructs an ItemsHandler ready to be registered on a
// router group.
func NewItemsHandler() *ItemsHandler {
	return &ItemsHandler{
		nextID: 1,
		items:  make(map[int]Item),
	}
}

// List handles GET /items, demonstrating pagination via httputil.
func (h *ItemsHandler) List(c *gin.Context) {
	page, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	all := make([]Item, 0, len(h.items))
	for _, item := range h.items {
		all = append(all, item)
	}

	start := page.Offset()
	end := start + page.PerPage
	if start > len(all) {
		start = len(all)
	}
	if end > len(all) {
		end = len(all)
	}

	httputil.OKWithMeta(c, all[start:end], page.ToMeta(len(all)))
}

// Get handles GET /items/:id.
func (h *ItemsHandler) Get(c *gin.Context) {
	var uri itemURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	h.mu.Lock()
	item, found := h.items[uri.ID]
	h.mu.Unlock()

	if !found {
		httputil.NotFound(c, "item not found")
		return
	}

	httputil.OK(c, item)
}

// Create handles POST /items.
func (h *ItemsHandler) Create(c *gin.Context) {
	var req createItemRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	h.mu.Lock()
	item := Item{ID: h.nextID, Name: req.Name, Price: req.Price}
	h.items[item.ID] = item
	h.nextID++
	h.mu.Unlock()

	httputil.Created(c, item)
}

// Update handles PUT /items/:id.
func (h *ItemsHandler) Update(c *gin.Context) {
	var uri itemURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	var req updateItemRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, found := h.items[uri.ID]; !found {
		httputil.NotFound(c, "item not found")
		return
	}

	item := Item{ID: uri.ID, Name: req.Name, Price: req.Price}
	h.items[uri.ID] = item

	httputil.OK(c, item)
}

// Delete handles DELETE /items/:id.
func (h *ItemsHandler) Delete(c *gin.Context) {
	var uri itemURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, found := h.items[uri.ID]; !found {
		httputil.NotFound(c, "item not found")
		return
	}

	delete(h.items, uri.ID)
	c.Status(http.StatusNoContent)
}
