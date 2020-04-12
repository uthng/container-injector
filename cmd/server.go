/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	log "github.com/uthng/golog"

	"github.com/uthng/container-injector/server/http"
)

var (
	serverAddr     string
	serverCertFile string
	serverKeyFile  string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Init HTTP server to inject container sidecar.",
	Long:  `A HTTP Webhook Server listens on a given port from Kubernetes APIServer to inject container sidecar.`,
	Run: func(cmd *cobra.Command, args []string) {
		initServer(args)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	serverCmd.PersistentFlags().StringVar(&serverAddr, "addr", ":8443", "Server listening addr.")
	serverCmd.PersistentFlags().StringVar(&serverCertFile, "cert", "/etc/webhook/certs/cert.pem", "X.509 certificat for HTTPS")
	serverCmd.PersistentFlags().StringVar(&serverKeyFile, "key", "/etc/webhook/certs/key.pem", "X.509 Privaye Key for HTTPS")
}

func initServer(args []string) {
	errs := make(chan error)

	// Set default verbosity
	logger := log.NewLogger()
	logger.SetVerbosity(verbosity)
	logger.DisableColor()

	// Set gitllabl logger
	httpLogger := log.NewLogger()
	httpLogger.SetVerbosity(verbosity)
	httpLogger.DisableColor()

	// Initialize http server
	httpServer := http.NewServer(serverAddr, serverCertFile, serverKeyFile, httpLogger)

	// HTTP
	go func() {
		logger.Infow("HTTP server starts listening", "addr", serverAddr)
		errs <- httpServer.Serve()
	}()

	// Interuption
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Errorw("Exit", "err", <-errs)
}
