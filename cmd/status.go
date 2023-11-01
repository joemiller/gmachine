package cmd

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
	rootCmd.AddCommand(statusCmd)
}

func status(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	names := []string{}
	for _, m := range cfg.Machines {
		names = append(names, m.Name)
	}

	// initialize output table writer
	table := tabwriter.NewWriter(cmd.OutOrStdout(), 5, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	print := func(values ...string) {
		fmt.Fprintln(table, strings.Join(values, "\t"))
	}
	print("NAME", "ACCOUNT", "PROJECT", "ZONE", "MACHINE_TYPE", "PREEMPTIBLE", "ENCRYPTION", "SERVICE_ACCOUNT", "INTERNAL_IP", "EXTERNAL_IP", "STATUS", "DEFAULT")

	eg := errgroup.Group{}
	eg.SetLimit(8)

	printerWg := sync.WaitGroup{}
	outputCh := make(chan []string)

	printerWg.Add(1)
	go func() {
		defer printerWg.Done()
		for row := range outputCh {
			print(row...)
		}

		if err := table.Flush(); err != nil {
			cmd.PrintErr(err)
		}
	}()

	// status row for each machine
	for _, name := range names {
		name := name

		eg.Go(func() error {
			machine, err := cfg.Get(name)
			if err != nil {
				return err
			}
			meta, err := gcp.DescribeInstance(name, machine.Account, machine.Project, machine.Zone)
			if err != nil {
				cmd.PrintErr(err)
				return nil
			}

			encryptStatus := ""
			if len(machine.CSEK) > 0 {
				encryptStatus = "CSEK"
			}

			gsa := ""
			if meta.ServiceAccounts != nil && len(meta.ServiceAccounts) > 0 {
				// XXX: just the first one. I am not sure you can assign multiple to a VM? if so, probably uncommon
				gsa = meta.ServiceAccounts[0].Email
			}

			// TODO: also handle CMEK encryption some day

			outputCh <- []string{
				name,
				machine.Account,
				machine.Project,
				path.Base(meta.Zone),
				path.Base(meta.MachineType),
				fmt.Sprintf("%t", meta.Scheduling.Preemptible),
				encryptStatus,
				gsa,
				internalIP(meta.NetworkInterfaces),
				externalIP(meta.NetworkInterfaces),
				meta.Status,
				defaultStr(cfg.GetDefault(), name),
			}
			return nil
		})
	}

	// wait for the `gcloud` describers to finish
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("gcloud error: %v", err)
	}
	close(outputCh)

	// wait for the table printer go routine to finish:
	printerWg.Wait()

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
