package main

import (
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/config"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/server"
	"git.netcracker.com/PROD.Platform.Streaming/disaster-recovery-daemon/utils"
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
