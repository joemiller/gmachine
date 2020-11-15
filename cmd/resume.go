package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// resumeCmd represents the resume command
var resumeCmd = &cobra.Command{
	Use:   "resume [NAME]",
	Short: "Resume a suspended machine",
	Long:  "Resume a suspended machine",
	Example: indentor.Indent("  ", `
# resume the default machine
gcloud machine resume

# resume a machine named 'machine2'
gcloud machine resume machine2
`),
	SilenceUsage: true,
	RunE:         resume,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func resume(cmd *cobra.Command, args []string) error {
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

	return gcp.ResumeInstance(
		cmd.OutOrStdout(),
		cmd.OutOrStderr(),
		name,
		machine.Project,
		machine.Zone,
		machine.CSEK,
	)
}
