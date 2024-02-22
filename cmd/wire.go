//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Krados/idpool/internal/conf"
	"github.com/Krados/idpool/internal/controller"
	"github.com/Krados/idpool/internal/data"
	"github.com/Krados/idpool/internal/server"
	"github.com/Krados/idpool/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"go.uber.org/zap"
)

func initApp(cfg *conf.Config, sugar *zap.SugaredLogger) (*gin.Engine, error) {
	panic(wire.Build(controller.ProviderSet, server.ProviderSet, service.ProviderSet, data.ProviderSet))
}
