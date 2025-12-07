package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "cinemaos-backend/internal/pkg/errors"
	"cinemaos-backend/internal/pkg/validator"
)

// Response represents a standard API response
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta holds pagination or other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Pagination holds pagination parameters
type Pagination struct {
	Page  int
	Limit int
}

// GetPagination extracts pagination from query params with defaults
func GetPagination(c *gin.Context) Pagination {
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed := parseInt(p); parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return Pagination{Page: page, Limit: limit}
}

// Offset calculates the offset for database queries
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		result = result*10 + int(c-'0')
	}
	return result
}

// Success sends a success response
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage sends a success response with a message
func SuccessWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Message: "Created successfully",
		Data:    data,
	})
}

// NoContent sends a 204 no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data any, pagination Pagination, total int64) {
	totalPages := int(total) / pagination.Limit
	if int(total)%pagination.Limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       pagination.Page,
			Limit:      pagination.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	switch e := err.(type) {
	case *apperrors.AppError:
		appErr = e
	default:
		appErr = apperrors.ErrInternal(err.Error())
	}

	c.JSON(appErr.HTTPStatus, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		},
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, errors []validator.ValidationError) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeValidation),
			Message: "Validation failed",
			Details: errors,
		},
	})
}

// BadRequest sends a 400 bad request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeBadRequest),
			Message: message,
		},
	})
}

// Unauthorized sends a 401 unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeUnauthorized),
			Message: message,
		},
	})
}

// Forbidden sends a 403 forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeForbidden),
			Message: message,
		},
	})
}

// NotFound sends a 404 not found response
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeNotFound),
			Message: resource + " not found",
		},
	})
}

// InternalError sends a 500 internal server error response
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorResponse{
			Code:    string(apperrors.CodeInternal),
			Message: "An internal error occurred",
		},
	})
}
