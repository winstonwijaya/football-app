package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"football-app/internal/dto"
	"football-app/internal/service"
	"football-app/pkg/response"
)

type MatchHandler struct {
	service *service.MatchService
}

func NewMatchHandler(service *service.MatchService) *MatchHandler {
	return &MatchHandler{service: service}
}

// Create handles POST /api/v1/matches
func (h *MatchHandler) Create(c *gin.Context) {
	var req dto.CreateMatchRequest
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

// List handles GET /api/v1/matches?status=&team_id=&from=&to=&page=&limit=
func (h *MatchHandler) List(c *gin.Context) {
	var query dto.ListMatchesQuery
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

// Get handles GET /api/v1/matches/:id
func (h *MatchHandler) Get(c *gin.Context) {
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

// Update handles PUT /api/v1/matches/:id
func (h *MatchHandler) Update(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateMatchRequest
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

// Delete handles DELETE /api/v1/matches/:id
func (h *MatchHandler) Delete(c *gin.Context) {
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
