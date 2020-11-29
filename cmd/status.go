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

	machines := []string{} // TODO better var name, maybe even 'names'
	if !all {
		name := cfg.GetDefault()
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			return errors.New("Must specify machine or set a default machine with 'set-default'")
		}
		machines = append(machines, name)
	} else {
		for _, m := range cfg.Machines {
			machines = append(machines, m.Name)
		}
	}

	// initialize output table writer
	table := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	print := func(values ...string) {
		fmt.Fprintln(table, strings.Join(values, "\t"))
	}
	print("NAME", "ZONE", "PROJECT", "MACHINE_TYPE", "PREEMPTIBLE", "INTERNAL_IP", "EXTERNAL_IP", "STATUS")

	// describe each
	for _, name := range machines {
		machine, err := cfg.Get(name)
		if err != nil {
			return err
		}
		meta, err := gcp.DescribeInstance(name, machine.Project, machine.Zone)
		if err != nil {
			cmd.PrintErr(err)
			continue
		}

		print(
			name,
			path.Base(meta.Zone),
			machine.Project,
			path.Base(meta.MachineType),
			fmt.Sprintf("%t", meta.Scheduling.Preemptible),
			internalIP(meta.NetworkInterfaces),
			externalIP(meta.NetworkInterfaces),
			meta.Status,
		)

	}
	err = table.Flush()
	if err != nil {
		return err
	}

	// instances = []
	// for i in cfg.Machines
	//   m, err = gcp.Describe(i)
	//   instances = append(m)
	//
	// print table header
	// for i in instances
	//    print table row, instance data
	//
	// TODO: fanout 'describe' calls to limited size worker pool
	//       fanin results to table printer.. maybe vault-token-helper has a reusable pattern

	// return gcp.StatusInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), name, machine.Project, machine.Zone)
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
	// iface := interfaces[0]
	if len(interfaces[0].AccessConfigs) == 0 {
		return ""
	}
	return interfaces[0].AccessConfigs[0].NatIP
}
