package handler

import (


	movieapp "cinemaos-backend/internal/application/movie"
	"cinemaos-backend/internal/pkg/response"
	"cinemaos-backend/internal/pkg/validator"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MovieHandler handles movie HTTP requests
type MovieHandler struct {
	movieService *movieapp.Service
	validator    *validator.Validator
}

// NewMovieHandler creates a new movie handler
func NewMovieHandler(movieService *movieapp.Service, validator *validator.Validator) *MovieHandler {
	return &MovieHandler{
		movieService: movieService,
		validator:    validator,
	}
}

// Create godoc
// @Summary Create movie
// @Description Create a new movie
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body movieapp.CreateMovieRequest true "Movie details"
// @Success 201 {object} response.Response{data=movieapp.MovieResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /movies [post]
func (h *MovieHandler) Create(c *gin.Context) {
	var req movieapp.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.movieService.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// GetByID godoc
// @Summary Get movie by ID
// @Description Get a movie by its ID
// @Tags movies
// @Produce json
// @Param id path string true "Movie ID"
// @Success 200 {object} response.Response{data=movieapp.MovieResponse}
// @Failure 404 {object} response.Response
// @Router /movies/{id} [get]
func (h *MovieHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid movie ID")
		return
	}

	result, err := h.movieService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Update godoc
// @Summary Update movie
// @Description Update an existing movie
// @Tags movies
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Movie ID"
// @Param request body movieapp.UpdateMovieRequest true "Movie updates"
// @Success 200 {object} response.Response{data=movieapp.MovieResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /movies/{id} [put]
func (h *MovieHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid movie ID")
		return
	}

	var req movieapp.UpdateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if errors := h.validator.Validate(req); errors != nil {
		response.ValidationError(c, errors)
		return
	}

	result, err := h.movieService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Delete godoc
// @Summary Delete movie
// @Description Soft delete a movie
// @Tags movies
// @Security BearerAuth
// @Param id path string true "Movie ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /movies/{id} [delete]
func (h *MovieHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid movie ID")
		return
	}

	if err := h.movieService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "Movie deleted successfully", nil)
}

// List godoc
// @Summary List movies
// @Description List movies with filters and pagination
// @Tags movies
// @Produce json
// @Param params query movieapp.MovieListParams false "Filter params"
// @Success 200 {object} response.Response{data=[]movieapp.MovieResponse}
// @Router /movies [get]
func (h *MovieHandler) List(c *gin.Context) {
	var params movieapp.MovieListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// Set defaults
	pagination := response.GetPagination(c)
	params.Page = pagination.Page
	params.Limit = pagination.Limit

	result, total, err := h.movieService.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, result, pagination, total)
}

// GetNowShowing godoc
// @Summary Get now showing movies
// @Description Get movies currently showing
// @Tags movies
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.Response{data=[]movieapp.MovieResponse}
// @Router /movies/now-showing [get]
func (h *MovieHandler) GetNowShowing(c *gin.Context) {
	pagination := response.GetPagination(c)

	result, total, err := h.movieService.GetNowShowing(c.Request.Context(), pagination.Page, pagination.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, result, pagination, total)
}

// GetComingSoon godoc
// @Summary Get coming soon movies
// @Description Get upcoming movies
// @Tags movies
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} response.Response{data=[]movieapp.MovieResponse}
// @Router /movies/coming-soon [get]
func (h *MovieHandler) GetComingSoon(c *gin.Context) {
	pagination := response.GetPagination(c)

	result, total, err := h.movieService.GetComingSoon(c.Request.Context(), pagination.Page, pagination.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Paginated(c, result, pagination, total)
}
