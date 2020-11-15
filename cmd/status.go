package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [NAME]",
	Short: "Print the current status of a machine",
	Long:  "Print the current status of a machine",
	Example: indentor.Indent("  ", `
# Print the status of the default machine
gmachine status

# Print the status of a machine named 'machine2'
gmachine status machine2
`),
	SilenceUsage: true,
	RunE:         status,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func status(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	name := cfg.GetDefault()
	if len(args) > 0 {
		name = args[0]
	}
	if name == "" {
		return errors.New("Must specify machine or set a default machine with 'set-default'")
	}

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	return gcp.StatusInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), name, machine.Project, machine.Zone)
}
