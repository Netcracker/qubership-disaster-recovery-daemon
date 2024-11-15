package main

import (
	"github.com/Netcracker/disaster-recovery-daemon/config"
	"github.com/Netcracker/disaster-recovery-daemon/server"
	"github.com/Netcracker/disaster-recovery-daemon/utils"
	"log"
	"log/slog"
	"os"
)

func main() {
	cfgLoader := config.GetDefaultEnvConfigLoader()
	cfg, err := config.NewConfig(cfgLoader)
	handler := utils.NewCustomLogHandler(os.Stdout)
	logger := slog.New(handler)
	slog.SetDefault(logger)
	if err != nil {
		log.Fatalln(err.Error())
	}
	server.NewServer(cfg).Run()
}
