package service

import (
	"DBProject/internal/db/service"
	"DBProject/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"net/http"
)

type ServiceHandler struct {
	Service *service.ServiceController
}

func New(pool *pgx.ConnPool) *ServiceHandler {
	return &ServiceHandler{Service: service.New(pool)}
}

func (handler *ServiceHandler) Status(c *gin.Context) {
	status, err := handler.Service.Status()

	if err != nil {
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (handler *ServiceHandler) Clear(c *gin.Context) {
	err := handler.Service.Clear()

	if err != nil {
		utils.WriteError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
