package httpserver

import (
	"crypto/tls"
	"fmt"
	"github.com/Netcracker/disaster-recovery-daemon/config"
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
