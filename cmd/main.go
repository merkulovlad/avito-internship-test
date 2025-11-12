package main

import (
	"log"
	
	"github.com/merkulovlad/avito-internship-test/internal/config"
	"github.com/merkulovlad/avito-internship-test/internal/logger"
)

func main() {
	cfg := config.MustLoad()
	options := &logger.Options{
		Level:      cfg.Log.Level,
		ToConsole:  cfg.Log.ToConsole,
		Filename:   cfg.Log.Filename,
	}
	logger, err := logger.NewLogger(options)
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()
}
