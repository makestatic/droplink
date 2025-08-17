package cli

import (
	"errors"
	"log"
	"time"

	"github.com/makestatic/droplink/internal/server"
)

func (cmd *Commands) Run() error {
	if Input == "" {
		return errors.New("input is required")
	}

	log.Println("Input: ", Input)

	if cmd.Password != "" {
		log.Println("Password: ", cmd.Password)
	}

	if cmd.Port != 8080 {
		log.Println("Port: ", cmd.Port)
	}

	if cmd.Global {
		log.Println("Global: ", cmd.Global)
	}

	if cmd.Zip {
		log.Println("Zip: ", cmd.Zip)
	}
	if cmd.Timeout > 0 {
		duration := time.Duration(cmd.Timeout) * time.Second
		log.Println("Timeout: ", duration)
	}

	server.Server(Input)

	return nil
}
