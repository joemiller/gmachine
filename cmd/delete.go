package cmd

import (
	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a cloud machine",
	Long:  "Delete a cloud machine",
	Example: indentor.Indent("  ", `
# delete the machine named 'machine1'
gcloud delete machine1
`),
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         delete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func delete(cmd *cobra.Command, args []string) error {
	name := args[0] // guaranteed not nil due to cobra.ExactArgs(1)

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	cmd.Println("Deleting...")

	err = gcp.DeleteInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), machine.Name, machine.Project, machine.Zone)
	if err != nil {
		return err
	}

	// add machine to config file
	err = cfg.Delete(name)
	if err != nil {
		return err
	}

	cmd.Println("Success")
	return nil
}
