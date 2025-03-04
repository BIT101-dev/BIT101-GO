package service

import (
	"BIT101-GO/config"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Service struct {
	app *gin.Engine
	cfg *config.Config
}

func NewService(app *gin.Engine, cfg *config.Config) *Service {
	return &Service{
		app: app,
		cfg: cfg,
	}
}

func (s *Service) Run() {
	fmt.Println("BIT101-GO will run on port " + s.cfg.Port)
	s.app.Run(":" + s.cfg.Port)
}
