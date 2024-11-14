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

package httpserver

import (
	"crypto/tls"
	"fmt"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
	"net/http"
)

func StartServer(handler http.Handler, config config.ServerConfig) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: handler,
	}
	if config.TLSEnabled {
		server.TLSConfig = &tls.Config{CipherSuites: config.Suites}
		return server.ListenAndServeTLS(fmt.Sprintf("%s/tls.crt", config.CertsPath), fmt.Sprintf("%s/tls.key", config.CertsPath))
	}
	return server.ListenAndServe()
}
