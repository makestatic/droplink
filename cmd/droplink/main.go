package main

import (
	"github.com/makestatic/droplink/internal/cli"
	logger "github.com/makestatic/droplink/internal/log"
)

func main() {
	initLogger()
	cli.Parser()
}

func initLogger() {
	_ = logger.Init(logger.Options{
		Level:      logger.LevelInfo,
		JSON:       true,
		AddSource:  true,
		OutputPath: "logs/droplink.log",
	})
}
