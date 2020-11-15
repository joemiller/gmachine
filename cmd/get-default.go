package cmd

import (
	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// getDefaultCmd represents the get-default command
var getDefaultCmd = &cobra.Command{
	Use:   "get-default",
	Short: "Print the name of the default machine in the config file, if set",
	Long:  "Print the name of the default machine in the config file, if set",
	Example: indentor.Indent("  ", `
# print current default machine
gcloud get-default
`),
	SilenceUsage: true,
	RunE:         getDefault,
}

func init() {
	rootCmd.AddCommand(getDefaultCmd)
}

func getDefault(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	cmd.Println(cfg.GetDefault())
	return nil
}
