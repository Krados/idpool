package controller

import (
	"errors"
	"net/http"

	"github.com/Krados/idpool/internal/service/idpool"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IDController struct {
	serv   *idpool.IDPoolServ
	logger *zap.SugaredLogger
}

func NewIDController(logger *zap.SugaredLogger, serv *idpool.IDPoolServ) *IDController {
	return &IDController{
		serv:   serv,
		logger: logger,
	}
}

func (s *IDController) NewID(c *gin.Context) {
	id, err := s.serv.Take()
	if err != nil {
		s.logger.Errorf("%+v", err)
		if errors.Is(err, idpool.NoIDAvailable{}) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": http.StatusInternalServerError,
				"msg":    err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"msg":    "InternalServerError",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"id":     id,
	})
}
