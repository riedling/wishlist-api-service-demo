package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/riedl/wishlist-api-service/internal/httputil"
)

// Wishlist is an example domain model representing a named collection of
// items owned by a user. Replace or extend this with a real repository-backed
// model as the service grows.
type Wishlist struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
	Items    []Item `json:"items"`
}

// publicWishlistsQuery is the validated query-string payload for fetching a
// user's public wishlists.
type publicWishlistsQuery struct {
	UserID string `form:"user_id" binding:"required"`
}

// createWishlistRequest is the validated payload for creating a wishlist.
type createWishlistRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Name     string `json:"name" binding:"required,min=1,max=200"`
	IsPublic bool   `json:"is_public"`
}

// wishlistURI binds the ":id" path parameter shared by single-wishlist
// routes.
type wishlistURI struct {
	ID string `uri:"id" binding:"required"`
}

// WishlistsHandler groups the handlers for the /wishlists resource. It holds
// its own dependencies (here, an in-memory store) so it's easy to swap in a
// real repository/service layer without touching routing code.
type WishlistsHandler struct {
	mu        sync.Mutex
	wishlists map[string]Wishlist
}

// NewWishlistsHandler constructs a WishlistsHandler ready to be registered on
// a router group.
func NewWishlistsHandler() *WishlistsHandler {
	return &WishlistsHandler{
		wishlists: make(map[string]Wishlist),
	}
}

// ListPublicForUser handles GET /wishlists/public?user_id=..., returning
// every wishlist owned by the given user that is marked public. Results are
// paginated using the shared httputil pagination helper.
func (h *WishlistsHandler) ListPublicForUser(c *gin.Context) {
	var query publicWishlistsQuery
	if !httputil.BindQuery(c, &query) {
		return
	}

	page, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	matches := make([]Wishlist, 0)
	for _, wl := range h.wishlists {
		if wl.UserID == query.UserID && wl.IsPublic {
			matches = append(matches, wl)
		}
	}

	start := page.Offset()
	end := start + page.PerPage
	if start > len(matches) {
		start = len(matches)
	}
	if end > len(matches) {
		end = len(matches)
	}

	httputil.OKWithMeta(c, matches[start:end], page.ToMeta(len(matches)))
}

// Create handles POST /wishlists, allowing a user to create a new wishlist.
func (h *WishlistsHandler) Create(c *gin.Context) {
	var req createWishlistRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	wl := Wishlist{
		ID:       uuid.NewString(),
		UserID:   req.UserID,
		Name:     req.Name,
		IsPublic: req.IsPublic,
		Items:    []Item{},
	}

	h.mu.Lock()
	h.wishlists[wl.ID] = wl
	h.mu.Unlock()

	httputil.Created(c, wl)
}

// Delete handles DELETE /wishlists/:id, removing the wishlist with the given
// ID.
func (h *WishlistsHandler) Delete(c *gin.Context) {
	var uri wishlistURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, found := h.wishlists[uri.ID]; !found {
		httputil.NotFound(c, "wishlist not found")
		return
	}

	delete(h.wishlists, uri.ID)
	c.Status(http.StatusNoContent)
}
