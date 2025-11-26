/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	rootCmd "localplane/cmd/root"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd.Execute()
}
