package cmd

import (
	"errors"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// resizeCmd represents the resize command
var resizeCmd = &cobra.Command{
	Use:   "resize NAME",
	Short: "Resize a cloud machine",
	Long:  "Resize a cloud machine",
	Example: indentor.Indent("  ", `
# resize the machine named 'machine1' to a pre-set machine-type
gmachine resize machine1 --type n2d-standard-32

# resize the machine named 'machine1' to a custom size
gmachine resize machine1 --type n2-custom-8-8192
`),
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         resize,
}

func init() {
	resizeCmd.Flags().StringP("type", "t", "", "Resize the machine to the specified machine-type")

	rootCmd.AddCommand(resizeCmd)
}

func resize(cmd *cobra.Command, args []string) error {
	name := args[0] // guaranteed not nil due to cobra.ExactArgs(1)

	size, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	if size == "" {
		return errors.New("--type/-t not specified")
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	machine, err := cfg.Get(name)
	if err != nil {
		return err
	}

	cmd.Printf("Resizing %s to %s...\n", machine.Name, size)
	err = gcp.ResizeInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), machine.Name, machine.Project, machine.Zone, size)
	if err != nil {
		return err
	}

	cmd.Println("Success")
	return nil
}
