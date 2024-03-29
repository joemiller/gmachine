package gcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/api/compute/v1"
)

// CreateRequest represents a configuration for creating a new instance with the
// CreateInstance() func.
type CreateRequest struct {
	Name             string
	Account          string
	Project          string
	Zone             string
	MachineType      string
	BootDiskSize     string
	BootDiskType     string
	ImageProject     string
	ImageFamily      string
	Metadata         map[string]string
	CSEK             CSEKBundle
	ServiceAccount   string
	NoServiceAccount bool
	StartupScript    string
	StartupScriptURL string
}

// AddMetadata adds a key=value pair to the instance's metadata.
func (r *CreateRequest) AddMetadata(key, val string) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]string)
	}
	r.Metadata[key] = val
}

// TODO doc
func CreateInstance(log, logerr io.Writer, req CreateRequest) error {
	var err error

	args := []string{
		"gcloud", "beta", "compute", "instances", "create",
		req.Name,
		"--account=" + req.Account,
		"--project=" + req.Project,
		"--zone=" + req.Zone,
		"--machine-type=" + req.MachineType,
		"--boot-disk-size=" + req.BootDiskSize,
		"--boot-disk-type=" + req.BootDiskType,
		"--image-project=" + req.ImageProject,
		"--image-family=" + req.ImageFamily,
	}

	if !req.NoServiceAccount && req.ServiceAccount != "" {
		args = append(args, "--service-account="+req.ServiceAccount)
	}
	if req.NoServiceAccount {
		args = append(args, "--no-service-account", "--no-scopes")
	}

	// [--metadata=KEY=VALUE,[KEY=VALUE,...]]
	metadata := []string{}
	if len(req.Metadata) > 0 {
		for k, v := range req.Metadata {
			metadata = append(metadata, fmt.Sprintf("%s=%s", k, v))
		}
	}
	// --metadata=startup-script-url=URL
	if req.StartupScriptURL != "" {
		metadata = append(metadata, "startup-script-url="+req.StartupScriptURL)
	}
	if len(metadata) > 0 {
		args = append(args, "--metadata="+strings.Join(metadata, ","))
	}

	// startup-script-url uses `--metadata-from-file=``
	if req.StartupScript != "" {
		args = append(args, "--metadata-from-file=startup-script="+req.StartupScript)
	}

	// marshal CSEK to json and pass into gcloud via stdin
	var stdin []byte
	if len(req.CSEK) > 0 {
		stdin, err = req.CSEK.Marshal()
		if err != nil {
			return err
		}
		args = append(args, "--csek-key-file=-")
	}

	return run(bytes.NewReader(stdin), log, logerr, args...)
}

// TODO doc
func DeleteInstance(log, logerr io.Writer, name, account, project, zone string) error {
	args := []string{
		"gcloud", "compute", "instances", "delete",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
		"-q",
	}
	return run(nil, log, logerr, args...)
}

// TODO doc
// TODO support more gcloud-ssh flags like iap-tunnel. maybe make this a struct like SSHInput{}
func SSHInstance(name, account, project, zone string, extra string) error {
	args := []string{
		"gcloud", "compute", "ssh",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
	}
	if extra != "" {
		f := strings.Fields(extra)
		args = append(args, "--")
		args = append(args, f...)
	}
	return execve(args)
}

// TODO doc
func StopInstance(log, logerr io.Writer, name, account, project, zone string) error {
	args := []string{
		"gcloud", "compute", "instances", "stop",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
	}
	return run(os.Stdin, log, logerr, args...)
}

// TODO doc
func StartInstance(log, logerr io.Writer, name, account, project, zone string, csek CSEKBundle) error {
	var err error
	var stdin []byte

	args := []string{
		"gcloud", "beta", "compute", "instances", "start",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
	}

	// marshal CSEK to json and pass into gcloud via stdin
	if len(csek) > 0 {
		stdin, err = csek.Marshal()
		if err != nil {
			return err
		}
		args = append(args, "--csek-key-file=-")
	}

	// return run(bytes.NewReader(stdin), log, logerr, args...)
	return run(bytes.NewReader(stdin), os.Stdout, os.Stderr, args...)
}

// TODO doc
func SuspendInstance(log, logerr io.Writer, name, account, project, zone string) error {
	args := []string{
		"gcloud", "beta", "compute", "instances", "suspend",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
	}
	return run(os.Stdin, log, logerr, args...)
}

// TODO doc
func ResumeInstance(log, logerr io.Writer, name, account, project, zone string, csek CSEKBundle) error {
	var err error
	var stdin []byte

	args := []string{
		"gcloud", "beta", "compute", "instances", "resume",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
	}

	// marshal CSEK to json and pass into gcloud via stdin
	if len(csek) > 0 {
		stdin, err = csek.Marshal()
		if err != nil {
			return err
		}
		args = append(args, "--csek-key-file=-")
	}

	return run(bytes.NewReader(stdin), log, logerr, args...)
}

// From google-cloud-sdk/lib/googlecloudsdk/command_lib/compute/instances/flags.py
// the default table format for 'gcloud compute instances list' command:
const statusTable = `
table(
	name,
	zone.basename(),
	machineType.machine_type().basename(),
	scheduling.preemptible.yesno(yes=true, no=''),
	networkInterfaces[].networkIP.notnull().list():label=INTERNAL_IP,
	networkInterfaces[].accessConfigs[0].natIP.notnull().list():label=EXTERNAL_IP,
	status)
`

// TODO doc
func StatusInstance(log, logerr io.Writer, name, account, project, zone string) error {
	args := []string{
		"gcloud", "beta", "compute", "instances", "describe",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
		"--format=" + statusTable,
	}
	return run(os.Stdin, log, logerr, args...)
}

// TODO doc
func PrintIP(log, logerr io.Writer, name, account, project, zone string) error {
	args := []string{
		"gcloud", "beta", "compute", "instances", "describe",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
		"--format=get(networkInterfaces[0].accessConfigs[0].natIP)",
	}
	return run(os.Stdin, log, logerr, args...)
}

// TODO doc
func DescribeInstance(name, account, project, zone string) (compute.Instance, error) {
	var instance compute.Instance

	args := []string{
		"gcloud", "beta", "compute", "instances", "describe",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
		"--format=json",
	}

	b, err := output(args...)
	if err != nil {
		return instance, fmt.Errorf("(%s) %s", err, b)
	}

	err = json.Unmarshal(b, &instance)
	if err != nil {
		return instance, err
	}

	return instance, nil
}

// TODO doc
func ResizeInstance(log, logerr io.Writer, name, account, project, zone, size string) error {
	args := []string{
		"gcloud", "compute", "instances", "set-machine-type",
		name,
		"--account=" + account,
		"--project=" + project,
		"--zone=" + zone,
		"--machine-type=" + size,
	}
	return run(os.Stdin, log, logerr, args...)
}

// TODO doc
func OpenConsole(log, logerr io.Writer, name, project, zone, authuser string) error {
	// https://console.cloud.google.com/compute/instancesDetail/zones/us-west2-a/instances/joe-amd-dev1?authuser=1&project=planetscale-development
	url := fmt.Sprintf("https://console.cloud.google.com/compute/instancesDetail/zones/%s/instances/%s?project=%s",
		zone, name, project,
	)
	if authuser != "" {
		url += "&authuser=" + authuser
	}

	args := []string{
		"open", url,
	}
	return run(os.Stdin, log, logerr, args...)
}
