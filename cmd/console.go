package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:   "console [NAME]",
	Short: "Opens the Google Cloud Console for the given machine",
	Long:  "Opens the Google Cloud Console for the given machine",
	Example: indentor.Indent("  ", `
# Open the Google Cloud Console for the default machine
gmachine console

# Open the Google Cloud Console for a machine named 'machine2'
gmachine console machine2

# Set "authuser=1" in the URL when opening the Google Cloud Console
gmachine console -u 1
`),
	SilenceUsage: true,
	RunE:         console,
}

func init() {
	consoleCmd.Flags().StringP("authuser", "u", "", "The 'authuser=' var to add to the console URL")

	rootCmd.AddCommand(consoleCmd)
}

func console(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	name := cfg.GetDefault()
	if len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		return errors.New("must specify machine or set a default machine with 'set-default'")
	}

	authuser, err := cmd.Flags().GetString("authuser")
	if err != nil {
		return err
	}

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	return gcp.OpenConsole(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		name,
		machine.Project,
		machine.Zone,
		authuser,
	)
}
