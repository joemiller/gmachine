package cmd

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [NAME]",
	Short: "Print the current status of a machine",
	Long:  "Print the current status of a machine",
	Example: indentor.Indent("  ", `
# Print the status of the default machine
gmachine status

# Print the status of a machine named 'machine2'
gmachine status machine2
`),
	SilenceUsage: true,
	RunE:         status,
}

func init() {
	statusCmd.Flags().BoolP("all", "a", false, "Print status of all instances")

	rootCmd.AddCommand(statusCmd)
}

func status(cmd *cobra.Command, args []string) error {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	names := []string{}
	if !all {
		name := cfg.GetDefault()
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			return errors.New("must specify machine or set a default machine with 'set-default'")
		}
		names = append(names, name)
	} else {
		for _, m := range cfg.Machines {
			names = append(names, m.Name)
		}
	}

	// initialize output table writer
	table := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	print := func(values ...string) {
		fmt.Fprintln(table, strings.Join(values, "\t"))
	}
	print("NAME", "ACCOUNT", "PROJECT", "ZONE", "MACHINE_TYPE", "PREEMPTIBLE", "INTERNAL_IP", "EXTERNAL_IP", "STATUS", "DEFAULT")

	// status row for each machine
	for _, name := range names {
		machine, err := cfg.Get(name)
		if err != nil {
			return err
		}
		meta, err := gcp.DescribeInstance(name, machine.Account, machine.Project, machine.Zone)
		if err != nil {
			cmd.PrintErr(err)
			continue
		}

		print(
			name,
			machine.Account,
			machine.Project,
			path.Base(meta.Zone),
			path.Base(meta.MachineType),
			fmt.Sprintf("%t", meta.Scheduling.Preemptible),
			internalIP(meta.NetworkInterfaces),
			externalIP(meta.NetworkInterfaces),
			meta.Status,
			defaultStr(cfg.GetDefault(), name),
		)

	}
	err = table.Flush()
	if err != nil {
		return err
	}

	// TODO: fanout 'describe' calls to a limited size worker pool
	//       fanin results to table printer.. errgroup

	return nil
}

// return internalIP  if set, else empty string.
// Similar in behavior to gcloud --format='table(networkInterfaces[].networkIP.notnull().list():label=INTERNAL_IP)'
func internalIP(interfaces []*compute.NetworkInterface) string {
	if len(interfaces) == 0 {
		return ""
	}
	return interfaces[0].NetworkIP
}

// return externalIP (natIP) if set, else empty string.
// Similar in behavior to gcloud --format='table(networkInterfaces[].accessConfigs[0].natIP.notnull().list():label=EXTERNAL_IP)'
func externalIP(interfaces []*compute.NetworkInterface) string {
	if len(interfaces) == 0 {
		return ""
	}
	if len(interfaces[0].AccessConfigs) == 0 {
		return ""
	}
	return interfaces[0].AccessConfigs[0].NatIP
}

func defaultStr(def, name string) string {
	if def == name {
		return "*"
	}
	return ""
}
