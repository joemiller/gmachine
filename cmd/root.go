package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const defaultConfigFile = "~/.gmachine.yaml"

var (
	version = "development"
	verbose = false

	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gmachine",
	Short: "Manage cloud machines on Google Cloud Platform",
	Long:  "Manage cloud machines on Google Cloud Platform",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv()
	// config file location, order of preference: flag > environment > default

	cfgFile = defaultConfigFile
	if v := viper.GetString("GMACHINE_CONFIG"); v != "" {
		cfgFile = v
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, fmt.Sprintf("config file (default is %s)", defaultConfigFile))
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	rootCmd.AddCommand(versionCmd)
}
