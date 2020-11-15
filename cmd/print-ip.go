package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// printIPCmd represents the printIP command
var printIPCmd = &cobra.Command{
	Use:   "print-ip [NAME]",
	Short: "Print a machine's public IP if it is RUNNING",
	Long:  "Print a machine's public IP if it is RUNNING",
	Example: indentor.Indent("  ", `
# print public IP of the default machine
gcloud print-ip

# print public IP of a machine named 'machine2'
gcloud print-ip machine2
`),
	SilenceUsage: true,
	RunE:         printIP,
}

func init() {
	rootCmd.AddCommand(printIPCmd)
}

func printIP(cmd *cobra.Command, args []string) error {
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

	return gcp.PrintIP(cmd.OutOrStdout(), cmd.OutOrStderr(), name, machine.Project, machine.Zone)
}
