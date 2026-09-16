package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"football-app/internal/dto"
	"football-app/internal/service"
	"football-app/pkg/response"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	response.OK(c, http.StatusOK, resp)
}
