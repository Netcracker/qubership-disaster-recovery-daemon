// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"github.com/Netcracker/disaster-recovery-daemon/api/entity"
	"github.com/Netcracker/disaster-recovery-daemon/config"
	"github.com/Netcracker/disaster-recovery-daemon/internal/app"
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
