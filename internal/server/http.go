package server

import (
	"github.com/Krados/idpool/internal/controller"

	"github.com/gin-gonic/gin"
)

func NewHTTPServer(idController *controller.IDController) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("api").Group("v1")
	{
		v1.POST("newID", idController.NewID)
	}

	return r
}
