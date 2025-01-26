package cmd

import (
	"os"
	"runtime/debug"

	"github.com/SourceSign/CloudPrism/common"
	"github.com/apex/log"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName    = "cloudprism"
	appVersion = "dev"
)

// nolint: gochecknoglobals
var (
	appConfigFile string
	appDebug      bool

	Revision = func() string {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					return setting.Value
				}
			}
		}

		return common.GetDeploymentTargetName(".")
	}()
)

// nolint: gochecknoglobals
// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     appName,
	Long:    banner(),
	Version: appVersion,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.OnInitialize(setLogLevel, initConfig)

	rootCmd.Version = appVersion + " (rev: " + Revision + ")"
	rootCmd.Flags().BoolP("version", "v", false, "Show current '"+appName+"' version")

	rootCmd.PersistentFlags().BoolVarP(&appDebug, "debug", "d", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&appConfigFile, "config", "f", "rapid.yaml", "config file to use")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if appConfigFile != "" {
		viper.SetConfigFile(appConfigFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("rapid")
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		log.WithField("file", viper.ConfigFileUsed()).Debug("Using config file")
	} else {
		log.WithError(err).Debug("Using config file")
	}
}

// a fancy global ascii-art banner.
func banner() string {
	ascii := figure.NewFigure(appName, "chunky", true).String()
	team := "Proudly brought to you by three OPaaS from CCoE\n"
	revision := "Revision: " + Revision + "\n"

	return ascii + "\n" + "🚀 Rapid Application Platform CLI\n" + team + revision
}

// this function sets the correct loglevel globally, depending on command-line parameters set.
func setLogLevel() {
	if appDebug {
		log.WithField("appDebug", appDebug).Debug("Setting log level to DEBUG")
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
