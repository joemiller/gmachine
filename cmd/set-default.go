package cmd

import (
	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// setDefaultCmd represents the set-default command
var setDefaultCmd = &cobra.Command{
	Use:   "set-default NAME",
	Short: "Set the default machine to use when no machine name is specified",
	Long:  "Set the default machine to use when no machine name is specified",
	Example: indentor.Indent("  ", `
# set machine2 as the default
gcloud set-default machine2
`),
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         setDefault,
}

func init() {
	rootCmd.AddCommand(setDefaultCmd)
}

func setDefault(cmd *cobra.Command, args []string) error {
	name := args[0] // guaranteed not nil due to cobra.ExactArgs(1)

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	// add machine to config file
	err = cfg.SetDefault(name)
	if err != nil {
		return err
	}

	cmd.Println("Success")
	return nil
}
