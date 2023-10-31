package gcp

import (
	"io"
	"os"
)

func CreateServiceAccount(log, logerr io.Writer, name, account, project string) error {
	args := []string{
		"gcloud", "iam", "service-accounts", "create",
		name,
		"--account=" + account,
		"--project=" + project,
	}
	return run(os.Stdin, log, logerr, args...)
}
