package cmd

import (
	"fmt"

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

# force deletion on errors such as 'machine not found'. This will remove the machine from the config file.
gcloud delete machine1 -f
`),
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         delete,
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Delete the machine from the config file even if an error occurs deleting from GCP")

	rootCmd.AddCommand(deleteCmd)
}

func delete(cmd *cobra.Command, args []string) error {
	name := args[0] // guaranteed not nil due to cobra.ExactArgs(1)

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	cmd.Printf("Deleting %s...\n", machine.Name)

	err = gcp.DeleteInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), machine.Name, machine.Project, machine.Zone)
	if err != nil && !force {
		return fmt.Errorf("Delete failed: %v. (re-run with '-f' to delete %s from the config file)", err, machine.Name)
	}

	// remove machine from config file
	err = cfg.Delete(name)
	if err != nil {
		return err
	}

	cmd.Println("Success")
	return nil
}
