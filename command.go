package microservice

import (
	"fmt"
	"os"
	"strings"

	"github.com/ory/viper"
	"github.com/setare/microservice/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var RootCmd = &cobra.Command{
	Use: os.Args[0],
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			logging.Logger.Error("failed getting current directory", zap.Error(err))
			os.Exit(1)
		}
		viper.AddConfigPath(cwd)
		viper.SetConfigName(".microservice")
	}

	// Set replacer...
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logging.Logger.Info("initialized", zap.String("config_file", viper.ConfigFileUsed()))
	} else {
		logging.Logger.Info("initialized without config file")
	}
}

// Execute starts the microservice based on its configuration.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(EC_ROOTCMD_FAILED)
	}
}

// AddCommand is a helper for `Rootcmd.AddCommand`.
func AddCommand(cmds ...*cobra.Command) {
	RootCmd.AddCommand(cmds...)
}
