package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh [NAME]",
	Short: "Spawn 'gcloud compute ssh' to connect to a machine",
	Long:  "Spawn 'gcloud compute ssh' to connect to a machine",
	Example: indentor.Indent("  ", `
# open a shell via ssh on the default machine
gmachine ssh

# open a shell via ssh on the machine named 'machine2'
gmachine ssh machine2
	`),
	SilenceUsage: true,
	RunE:         ssh,
}

func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.Flags().String("ssh-args", "", "Additional ssh args to pass to ssh (example '-A -C'). Overrides default_ssh_args from config file if set.'")
	sshCmd.Flags().BoolP("agent-forward", "A", false, "Enable SSH Agent forwarding")
}

func ssh(cmd *cobra.Command, args []string) error {
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

	sshArgs := machine.DefaultSSHArgs
	if args, err := cmd.Flags().GetString("ssh-args"); err == nil && args != "" {
		sshArgs = args
	}

	if forward, _ := cmd.Flags().GetBool("agent-forward"); forward {
		sshArgs = sshArgs + " -A"
	}

	return gcp.SSHInstance(name,
		machine.Account,
		machine.Project,
		machine.Zone,
		sshArgs,
	)
}
