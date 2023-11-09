package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [NAME]",
	Short: "Start a stopped cloud machine",
	Long:  "Start a stopped cloud machine",
	Example: indentor.Indent("  ", `
# Start the default machine
gmachine start

# Start a machine named 'machine2'
gmachine start machine2
`),
	SilenceUsage: true,
	RunE:         start,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) error {
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

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	return gcp.StartInstance(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		name,
		machine.Account,
		machine.Project,
		machine.Zone,
		machine.CSEK,
	)
}
