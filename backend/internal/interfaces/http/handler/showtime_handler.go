package handler

import (
	"cinemaos-backend/internal/application/showtime"
	"cinemaos-backend/internal/pkg/response"
	"cinemaos-backend/internal/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ShowtimeHandler handles showtime HTTP requests
type ShowtimeHandler struct {
	service   *showtime.Service
	validator *validator.Validator
}

// NewShowtimeHandler creates a new showtime handler
func NewShowtimeHandler(service *showtime.Service, validator *validator.Validator) *ShowtimeHandler {
	return &ShowtimeHandler{
		service:   service,
		validator: validator,
	}
}

// Create creates a new showtime
func (h *ShowtimeHandler) Create(c *gin.Context) {
	var req showtime.CreateShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if validationErrors := h.validator.Validate(req); validationErrors != nil {
		response.ValidationError(c, validationErrors)
		return
	}

	res, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		// assuming service returns standard error, let response.Error handle it
		response.Error(c, err)
		return
	}

	response.Created(c, res)
}

// GetByID gets a showtime by ID
func (h *ShowtimeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid showtime ID")
		return
	}

	res, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		// response.Error handles mapping to 404 if it's a known error, 
		// otherwise might return 500. For now relying on it.
		// Detailed handling might require checking error type.
		response.Error(c, err)
		return
	}

	response.Success(c, res)
}

// List lists showtimes
func (h *ShowtimeHandler) List(c *gin.Context) {
	var params showtime.ShowtimeListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	res, err := h.service.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, res)
}

// Update updates a showtime
func (h *ShowtimeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid showtime ID")
		return
	}

	var req showtime.UpdateShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	if validationErrors := h.validator.Validate(req); validationErrors != nil {
		response.ValidationError(c, validationErrors)
		return
	}

	res, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, res)
}

// Delete deletes a showtime
func (h *ShowtimeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid showtime ID")
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Showtime deleted successfully", nil)
}
