package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"football-app/internal/dto"
	"football-app/internal/service"
	"football-app/pkg/response"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(service *service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// MatchReport handles GET /api/v1/reports/matches/:id
func (h *ReportHandler) MatchReport(c *gin.Context) {
	id, err := idParam(c)
	if err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.MatchReport(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	response.OK(c, http.StatusOK, resp)
}

// ListMatchReports handles GET /api/v1/reports/matches?page=&limit=
func (h *ReportHandler) ListMatchReports(c *gin.Context) {
	var query dto.ListMatchReportsQuery
	if err := bindQuery(c, &query); err != nil {
		c.Error(err)
		return
	}
	query.Normalize()

	items, total, err := h.service.ListMatchReports(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}
	response.OKWithMeta(c, http.StatusOK, items, response.Meta{Page: query.Page, Limit: query.Limit, Total: total})
}
