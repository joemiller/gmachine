package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// suspendCmd represents the suspend command
var suspendCmd = &cobra.Command{
	Use:   "suspend [NAME]",
	Short: "Suspend a running machine",
	Long:  "Suspend a running machine",
	Example: indentor.Indent("  ", `
# Suspend the default machine
gmachine suspend

# Suspend a machine named 'machine2'
gmachine suspend machine2
`),
	SilenceUsage: true,
	RunE:         suspend,
}

func init() {
	rootCmd.AddCommand(suspendCmd)
}

func suspend(cmd *cobra.Command, args []string) error {
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

	return gcp.SuspendInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), name, machine.Project, machine.Zone)
}
