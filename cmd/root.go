package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	// config file location, order of preference: (1) -config flag > (2) environment > (3) default config dir

	// Default config dir depends on the platform: https://golang.org/pkg/os/#UserConfigDir
	cfgDir, _ := os.UserConfigDir()
	cfgFile = filepath.Join(cfgDir, "gmachine", "gmachine.yaml")

	if v := viper.GetString("GMACHINE_CONFIG"); v != "" {
		cfgFile = v
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "Config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	rootCmd.AddCommand(versionCmd)
}
