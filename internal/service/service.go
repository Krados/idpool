package service

import (
	"github.com/Krados/idpool/internal/service/idpool"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	idpool.NewIDPoolServ,
	idpool.NewMysqlIDRepo,
)
