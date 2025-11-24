/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	rootCmd "k8s-local-bench/cmd/root"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd.Execute()
}
