// Package httputil provides extensible, reusable utilities for handling
// HTTP requests and responses across all API endpoints. Handlers should
// use these helpers instead of calling gin.Context methods directly so
// that response shape, error handling, and request parsing stay
// consistent throughout the service.
package httputil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Meta carries optional metadata about a response, such as pagination info.
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalItems int `json:"total_items,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// envelope is the consistent JSON shape returned by every endpoint.
type envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorBody describes an error returned to API clients.
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// OK writes a 200 response with the given data payload.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, envelope{Success: true, Data: data})
}

// OKWithMeta writes a 200 response with data and metadata (e.g. pagination).
func OKWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, envelope{Success: true, Data: data, Meta: meta})
}

// Created writes a 201 response with the given data payload.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, envelope{Success: true, Data: data})
}

// NoContent writes a 204 response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Fail writes an error response with the given HTTP status code, machine
// readable error code, and human readable message. Optional details (e.g.
// validation errors) can be attached.
func Fail(c *gin.Context, status int, code, message string, details ...interface{}) {
	body := &ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details[0]
	}
	c.JSON(status, envelope{Success: false, Error: body})
}

// AbortWithError writes an error response and stops further middleware and
// handler execution via c.Abort(). Prefer this inside middleware.
func AbortWithError(c *gin.Context, status int, code, message string, details ...interface{}) {
	body := &ErrorBody{Code: code, Message: message}
	if len(details) > 0 {
		body.Details = details[0]
	}
	c.AbortWithStatusJSON(status, envelope{Success: false, Error: body})
}

// Common convenience wrappers for typical HTTP error responses.

// BadRequest writes a 400 response.
func BadRequest(c *gin.Context, message string, details ...interface{}) {
	Fail(c, http.StatusBadRequest, "bad_request", message, details...)
}

// Unauthorized writes a 401 response.
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, "unauthorized", message)
}

// Forbidden writes a 403 response.
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, "forbidden", message)
}

// NotFound writes a 404 response.
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, "not_found", message)
}

// Conflict writes a 409 response.
func Conflict(c *gin.Context, message string) {
	Fail(c, http.StatusConflict, "conflict", message)
}

// UnprocessableEntity writes a 422 response, typically used for validation
// failures. Details commonly holds field-level validation errors.
func UnprocessableEntity(c *gin.Context, message string, details ...interface{}) {
	Fail(c, http.StatusUnprocessableEntity, "validation_failed", message, details...)
}

// InternalError writes a 500 response. The underlying error is intentionally
// not exposed to the client; log it separately before calling this helper.
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "an unexpected error occurred"
	}
	Fail(c, http.StatusInternalServerError, "internal_error", message)
}

// BindJSON parses and validates the request body into dst (which should be a
// pointer to a struct with `binding` tags). On failure it writes a 400
// response with field-level validation details and returns false so the
// caller can early-return.
func BindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		BadRequest(c, "invalid request body", ValidationDetails(err))
		return false
	}
	return true
}

// BindQuery parses and validates query string parameters into dst. On
// failure it writes a 400 response and returns false.
func BindQuery(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		BadRequest(c, "invalid query parameters", ValidationDetails(err))
		return false
	}
	return true
}

// BindURI parses and validates URI path parameters into dst. On failure it
// writes a 400 response and returns false.
func BindURI(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindUri(dst); err != nil {
		BadRequest(c, "invalid path parameters", ValidationDetails(err))
		return false
	}
	return true
}

// ValidationDetails converts a binding/validation error into a friendly,
// serializable representation. It understands go-playground/validator
// errors (used internally by Gin) and falls back to the raw error string.
func ValidationDetails(err error) interface{} {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[fe.Field()] = fieldErrorMessage(fe)
		}
		return fields
	}
	return err.Error()
}

func fieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "value is below the minimum of " + fe.Param()
	case "max":
		return "value exceeds the maximum of " + fe.Param()
	default:
		return "failed validation: " + fe.Tag()
	}
}
