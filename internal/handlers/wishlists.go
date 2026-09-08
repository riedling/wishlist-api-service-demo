package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/riedl/wishlist-api-service/internal/httputil"
)

// Visibility represents the privacy setting of a wishlist.
type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityUnlisted Visibility = "unlisted"
)

// Wishlist is an example domain model representing a named collection of
// items owned by a user. Replace or extend this with a real repository-backed
// model as the service grows.
type Wishlist struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	IsPublic   bool       `json:"is_public"`
	Visibility Visibility `json:"visibility"`
	Items      []Item     `json:"items"`
}

// publicWishlistsQuery is the validated query-string payload for fetching a
// user's public wishlists.
type publicWishlistsQuery struct {
	UserID string `form:"user_id" binding:"required"`
}

// userWishlistsURI binds the ":user_id" path parameter shared by routes that
// operate on all of a user's wishlists.
type userWishlistsURI struct {
	UserID string `uri:"user_id" binding:"required"`
}

// visibilityQuery is the validated query-string payload for filtering a
// user's wishlists by privacy setting.
type visibilityQuery struct {
	Visibility string `form:"visibility" binding:"required,oneof=public private unlisted"`
}

// createWishlistRequest is the validated payload for creating a wishlist.
type createWishlistRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Name       string `json:"name" binding:"required,min=1,max=200"`
	IsPublic   bool   `json:"is_public"`
	Visibility string `json:"visibility" binding:"omitempty,oneof=public private unlisted"`
}

// updateWishlistRequest is the validated payload for updating a wishlist.
// Every field is optional (pointer) so callers can supply any subset of
// fields that exist on the wishlist; only fields present in the request
// body are applied.
type updateWishlistRequest struct {
	UserID     *string `json:"user_id" binding:"omitempty,min=1"`
	Name       *string `json:"name" binding:"omitempty,min=1,max=200"`
	IsPublic   *bool   `json:"is_public"`
	Visibility *string `json:"visibility" binding:"omitempty,oneof=public private unlisted"`
	Items      *[]Item `json:"items"`
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

// Create handles POST /wishlist, allowing a user to create a new wishlist.
func (h *WishlistsHandler) Create(c *gin.Context) {
	var req createWishlistRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	wl := Wishlist{
		ID:         uuid.NewString(),
		UserID:     req.UserID,
		Name:       req.Name,
		IsPublic:   req.IsPublic,
		Visibility: resolveVisibility(req.Visibility, req.IsPublic),
		Items:      []Item{},
	}

	h.mu.Lock()
	h.wishlists[wl.ID] = wl
	h.mu.Unlock()

	httputil.Created(c, wl)
}

// Update handles PUT /wishlists/:id, applying any subset of updatable
// fields supplied in the request body to the existing wishlist.
func (h *WishlistsHandler) Update(c *gin.Context) {
	var uri wishlistURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	var req updateWishlistRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	wl, found := h.wishlists[uri.ID]
	if !found {
		httputil.NotFound(c, "wishlist not found")
		return
	}

	if req.UserID != nil {
		wl.UserID = *req.UserID
	}
	if req.Name != nil {
		wl.Name = *req.Name
	}
	if req.IsPublic != nil {
		wl.IsPublic = *req.IsPublic
	}
	if req.Visibility != nil {
		wl.Visibility = Visibility(*req.Visibility)
	}
	if req.Items != nil {
		wl.Items = *req.Items
	}

	h.wishlists[uri.ID] = wl

	httputil.OK(c, wl)
}

// Delete handles DELETE /wishlist/:id, removing the wishlist with the given
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

// wishlistPatch describes a single wishlist to create or update as part of
// a bulk modification request. If ID is empty, a new wishlist is created;
// otherwise the existing wishlist owned by the user is updated in place.
type wishlistPatch struct {
	ID         string  `json:"id"`
	Name       *string `json:"name" binding:"omitempty,min=1,max=200"`
	IsPublic   *bool   `json:"is_public"`
	Visibility *string `json:"visibility" binding:"omitempty,oneof=public private unlisted"`
	Items      *[]Item `json:"items"`
}

// modifyWishlistsRequest is the validated payload for bulk-modifying the
// wishlists belonging to a user.
type modifyWishlistsRequest struct {
	Wishlists []wishlistPatch `json:"wishlists" binding:"required,min=1,dive"`
}

// resolveVisibility determines the Visibility to store for a wishlist,
// preferring an explicit visibility value and falling back to the legacy
// is_public boolean for backward compatibility.
func resolveVisibility(visibility string, isPublic bool) Visibility {
	if visibility != "" {
		return Visibility(visibility)
	}
	if isPublic {
		return VisibilityPublic
	}
	return VisibilityPrivate
}

// GetAllForUser handles GET /wishlist/user/:user_id?visibility=..., returning
// every wishlist owned by the given user that matches the requested privacy
// setting (public, private, or unlisted).
func (h *WishlistsHandler) GetAllForUser(c *gin.Context) {
	var uri userWishlistsURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	var query visibilityQuery
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
		if wl.UserID == uri.UserID && string(wl.Visibility) == query.Visibility {
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

// DeleteAllForUser handles DELETE /wishlist/user/:user_id, removing every
// wishlist owned by the given user.
func (h *WishlistsHandler) DeleteAllForUser(c *gin.Context) {
	var uri userWishlistsURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	for id, wl := range h.wishlists {
		if wl.UserID == uri.UserID {
			delete(h.wishlists, id)
		}
	}

	c.Status(http.StatusNoContent)
}

// ModifyForUser handles POST /wishlist/user/:user_id, allowing a user to
// create and/or update multiple wishlists in a single request. Each entry
// in the request body is either applied as an update (if it references an
// existing wishlist ID owned by the user) or created as a new wishlist (if
// no ID is supplied).
func (h *WishlistsHandler) ModifyForUser(c *gin.Context) {
	var uri userWishlistsURI
	if !httputil.BindURI(c, &uri) {
		return
	}

	var req modifyWishlistsRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	results := make([]Wishlist, 0, len(req.Wishlists))
	for _, patch := range req.Wishlists {
		if patch.ID == "" {
			wl := Wishlist{
				ID:         uuid.NewString(),
				UserID:     uri.UserID,
				Visibility: VisibilityPrivate,
				Items:      []Item{},
			}
			if patch.Name != nil {
				wl.Name = *patch.Name
			}
			if patch.IsPublic != nil {
				wl.IsPublic = *patch.IsPublic
			}
			if patch.Visibility != nil {
				wl.Visibility = Visibility(*patch.Visibility)
			}
			if patch.Items != nil {
				wl.Items = *patch.Items
			}
			h.wishlists[wl.ID] = wl
			results = append(results, wl)
			continue
		}

		wl, found := h.wishlists[patch.ID]
		if !found || wl.UserID != uri.UserID {
			httputil.NotFound(c, "wishlist not found: "+patch.ID)
			return
		}

		if patch.Name != nil {
			wl.Name = *patch.Name
		}
		if patch.IsPublic != nil {
			wl.IsPublic = *patch.IsPublic
		}
		if patch.Visibility != nil {
			wl.Visibility = Visibility(*patch.Visibility)
		}
		if patch.Items != nil {
			wl.Items = *patch.Items
		}
		h.wishlists[patch.ID] = wl
		results = append(results, wl)
	}

	httputil.OK(c, results)
}
