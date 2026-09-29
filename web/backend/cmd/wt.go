package cmd

import (
	"backend/config"
	"backend/internal"
	"backend/logger"
	"os"
	"os/signal"
	"syscall"

	loggergoUtil "github.com/Alonza0314/logger-go/v2/util"
	"github.com/free-ran-ue/util"
	"github.com/spf13/cobra"
)

var wtCmd = &cobra.Command{
	Use: "wt",
	Run: wtFunc,
}

func init() {
	wtCmd.Flags().StringP("config", "c", "config.yaml", "Path to the configuration file")
	if err := wtCmd.MarkFlagRequired("config"); err != nil {
		panic(err)
	}
}

func wtFunc(cmd *cobra.Command, args []string) {
	wtConfigFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}

	wtConfig := config.Config{}
	if err := util.LoadFromYaml(wtConfigFilePath, &wtConfig); err != nil {
		panic(err)
	}

	logger := logger.NewBackendLogger(loggergoUtil.LogLevelString(wtConfig.Logger.Level), "", true)

	wt := internal.NewBackend(&wtConfig, logger)
	if wt == nil {
		panic("failed to initialize the backend")
	}

	wt.Start()
	defer wt.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
}

func Execute() {
	if err := wtCmd.Execute(); err != nil {
		panic(err)
	}
}
