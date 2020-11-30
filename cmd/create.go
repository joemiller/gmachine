package cmd

import (
	"errors"
	"fmt"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/joemiller/gmachine/internal/indentor"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a cloud machine",
	Long:  "Create a cloud machine",
	Example: indentor.Indent("  ", `
# Create a new machine 'machine1' in the 'my-proj' project, zone 'us-west1-a'
gmachine create machine1 -p my-proj -z us-west1-a

# Encrypt the machine's root disk using a locally stored CSEK key
gmachine create machine1 -p my-proj -z us-west1-a --csek

# List all options
gmachine create -h
`),
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
	RunE:         create,
}

func init() {
	createCmd.Flags().StringP("project", "p", "", "The Google Cloud project to create the instance in")
	createCmd.Flags().StringP("zone", "z", "", "The Google Cloud zone to create the instance in")
	createCmd.Flags().String("disk-size", "10GB", "Size of the boot disk. Valid units: KB, MB, GB, TB")
	createCmd.Flags().String("disk-type", "pd-standard", "The type of the boot disk. Run 'gcloud compute disk-types list' for valid types")
	createCmd.Flags().String("image-project", "ubuntu-os-cloud", "The Google Cloud project against which all image and image family references will be resolved")
	createCmd.Flags().String("image-family", "ubuntu-2010", "The image family for the operating system that the boot disk will be initialized with")
	createCmd.Flags().Bool("csek", false, "Encrypt the boot disk with a customer-supplied-encryption-key. A key will be generated and stored in the local config file")
	createCmd.Flags().String("machine-type", "f1-micro", "Specifies the machine type used for the instances. To get a list of available machine types, run 'gcloud compute machine-types list'")
	createCmd.Flags().Bool("disable-ssh-project-keys", true, "Disable automatically adding project SSH key users to the instance")
	createCmd.Flags().Bool("set-default", false, "Set this instance as the default. The first instance will always be set as default")

	// TODO: there are so many more options that we might support over time, some examples:
	//   * --image (if specified, --image-family can't be used)
	//   * --service-account
	//   * --startup-script / --startup-script-url
	//   * --network / --subnet
	//   * --preemptible

	rootCmd.AddCommand(createCmd)
}

func create(cmd *cobra.Command, args []string) error {
	name := args[0] // guaranteed not nil due to cobra.ExactArgs(1)

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}
	zone, err := cmd.Flags().GetString("zone")
	if err != nil {
		return err
	}
	diskSize, err := cmd.Flags().GetString("disk-size")
	if err != nil {
		return err
	}
	diskType, err := cmd.Flags().GetString("disk-type")
	if err != nil {
		return err
	}
	imageProject, err := cmd.Flags().GetString("image-project")
	if err != nil {
		return err
	}
	imageFamily, err := cmd.Flags().GetString("image-family")
	if err != nil {
		return err
	}
	encrypt, err := cmd.Flags().GetBool("csek")
	if err != nil {
		return err
	}
	machineType, err := cmd.Flags().GetString("machine-type")
	if err != nil {
		return err
	}
	disableProjectSSHKeys, err := cmd.Flags().GetBool("disable-ssh-project-keys")
	if err != nil {
		return err
	}
	setAsDefault, err := cmd.Flags().GetBool("set-default")
	if err != nil {
		return err
	}

	if name == "" || project == "" || zone == "" {
		return errors.New("Missing required arguments: name, project, zone. Use -h for help.")
	}

	cfg, err := config.LoadFile(cfgFile)
	if err != nil {
		return err
	}

	// check if a matching machine is already in the config
	if cfg.Exists(name) {
		return fmt.Errorf("machine '%s' already exists in the config file", name)
	}

	// generate new csek key if requested
	var csekBundle gcp.CSEKBundle
	if encrypt {
		csekBundle, err = gcp.CreateCSEK(gcp.DiskURI(project, zone, name))
		if err != nil {
			return fmt.Errorf("Failed generating CSEK Key: %w", err)
		}
	}

	req := gcp.CreateRequest{
		Name:         name,
		Project:      project,
		Zone:         zone,
		MachineType:  machineType,
		BootDiskSize: diskSize,
		BootDiskType: diskType,
		ImageProject: imageProject,
		ImageFamily:  imageFamily,
		CSEK:         csekBundle,
	}
	if disableProjectSSHKeys {
		req.AddMetadata("block-project-ssh-keys", "true")
	}

	cmd.Println("Creating...")
	err = gcp.CreateInstance(cmd.OutOrStdout(), cmd.OutOrStderr(), req)
	if err != nil {
		return err
	}

	// add machine to config file
	err = cfg.Add(name, project, zone, csekBundle)
	if err != nil {
		return err
	}

	if setAsDefault {
		if err = cfg.SetDefault(name); err != nil {
			return err
		}
	}
	cmd.Println("Success")
	return nil
}
