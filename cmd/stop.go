package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [NAME]",
	Short: "Stop a running cloud machine",
	Long:  "Stop a running cloud machine",
	Example: indentor.Indent("  ", `
# Stop the default machine
gmachine stop

# Stop a machine named 'machine2'
gmachine stop machine2
`),
	SilenceUsage: true,
	RunE:         stop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stop(cmd *cobra.Command, args []string) error {
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

	return gcp.StopInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), name, machine.Project, machine.Zone)
}
