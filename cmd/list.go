package cmd

import (
	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cloud machines in the config file",
	Long:  "List all cloud machines in the config file",
	Example: indentor.Indent("  ", `
# list
gmachine list
`),
	SilenceUsage: true,
	RunE:         list,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, _ []string) error {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	for _, m := range cfg.Machines {
		encrypted := false
		if m.CSEK != nil && len(m.CSEK) > 0 {
			encrypted = true
		}
		// TODO print out some kind of indication next to the default machine
		cmd.Printf("%s (%s, %s, %s, encrypted: %t)\n", m.Name, m.Account, m.Project, m.Zone, encrypted)
	}
	return nil
}

// TODO: maybe remove the list command since 'status' is similar?
