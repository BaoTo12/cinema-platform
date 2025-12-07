package handler

import (
	"net/http"

	cinemaapp "cinemaos-backend/internal/application/cinema"
	"cinemaos-backend/internal/pkg/response"
	"cinemaos-backend/internal/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CinemaHandler handles cinema HTTP requests
type CinemaHandler struct {
	cinemaService *cinemaapp.Service
	validator     *validator.Validator
}

// NewCinemaHandler creates a new cinema handler
func NewCinemaHandler(cinemaService *cinemaapp.Service, validator *validator.Validator) *CinemaHandler {
	return &CinemaHandler{
		cinemaService: cinemaService,
		validator:     validator,
	}
}

// Create godoc
// @Summary Create cinema
// @Description Create a new cinema
// @Tags cinemas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body cinemaapp.CreateCinemaRequest true "Cinema details"
// @Success 201 {object} response.Response{data=cinemaapp.CinemaResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /cinemas [post]
func (h *CinemaHandler) Create(c *gin.Context) {
	var req cinemaapp.CreateCinemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.cinemaService.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Cinema created successfully",
		Data:    result,
	})
}

// GetByID godoc
// @Summary Get cinema by ID
// @Description Get a cinema by its ID
// @Tags cinemas
// @Produce json
// @Param id path string true "Cinema ID"
// @Success 200 {object} response.Response{data=cinemaapp.CinemaResponse}
// @Failure 404 {object} response.Response
// @Router /cinemas/{id} [get]
func (h *CinemaHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cinema ID")
		return
	}

	result, err := h.cinemaService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// List godoc
// @Summary List cinemas
// @Description List cinemas with filters and pagination
// @Tags cinemas
// @Produce json
// @Param params query cinemaapp.CinemaListParams false "Filter params"
// @Success 200 {object} response.Response{data=[]cinemaapp.CinemaResponse}
// @Router /cinemas [get]
func (h *CinemaHandler) List(c *gin.Context) {
	var params cinemaapp.CinemaListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	pagination := response.GetPagination(c)
	params.Page = pagination.Page
	params.Limit = pagination.Limit

	result, total, err := h.cinemaService.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, result, pagination, total)
}

// AddScreen godoc
// @Summary Add screen to cinema
// @Description Add a new screen to a cinema
// @Tags cinemas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Cinema ID"
// @Param request body cinemaapp.CreateScreenRequest true "Screen details"
// @Success 201 {object} response.Response{data=cinemaapp.ScreenResponse}
// @Failure 404 {object} response.Response
// @Router /cinemas/{id}/screens [post]
func (h *CinemaHandler) AddScreen(c *gin.Context) {
	cinemaID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid cinema ID")
		return
	}

	var req cinemaapp.CreateScreenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.cinemaService.AddScreen(c.Request.Context(), cinemaID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "Screen added successfully",
		Data:    result,
	})
}
