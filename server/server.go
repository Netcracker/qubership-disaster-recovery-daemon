package server

import (
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/api/entity"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/internal/app"
	"log"
)

type Server struct {
	config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{config: config}
}

func (srv *Server) WithHealthFunc(healthFunc func(request entity.HealthRequest) (entity.HealthResponse, error), fullHealth bool) *Server {
	srv.config.HealthConfig.AdditionalHealthStatusConfig.HealthFunc = healthFunc
	srv.config.HealthConfig.AdditionalHealthStatusConfig.FullHealthEnabled = fullHealth
	return srv
}

func (srv *Server) Run() {
	log.Println("DR server started")
	app.Run(srv.config)
}
