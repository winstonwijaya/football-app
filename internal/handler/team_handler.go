package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"football-app/internal/dto"
	"football-app/internal/service"
	"football-app/pkg/response"
)

type TeamHandler struct {
	service *service.TeamService
}

func NewTeamHandler(service *service.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

// Create handles POST /api/v1/teams
func (h *TeamHandler) Create(c *gin.Context) {
	var req dto.CreateTeamRequest

	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Create(c.Request.Context(), req, userID(c))
	if err != nil {
		c.Error(err)
		return
	}

	response.OK(c, http.StatusCreated, resp)
}

// List handles GET /api/v1/teams?city=&search=&page=&limit=
func (h *TeamHandler) List(c *gin.Context) {
	var query dto.ListTeamsQuery
	if err := bindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}
	query.Normalize()

	items, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.OKWithMeta(c, http.StatusOK, items, response.Meta{Page: query.Page, Limit: query.Limit, Total: total})
}

// Get handles GET /api/v1/teams/:id
func (h *TeamHandler) Get(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.OK(c, http.StatusOK, resp)
}

// Update handles PUT /api/v1/teams/:id
func (h *TeamHandler) Update(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateTeamRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Update(c.Request.Context(), id, req, userID(c))
	if err != nil {
		c.Error(err)
		return
	}

	response.OK(c, http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/teams/:id
func (h *TeamHandler) Delete(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	if err := h.service.Delete(c.Request.Context(), id, userID(c)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
