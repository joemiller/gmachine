package gcp

import "fmt"

func DiskURI(project, zone, disk string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/disks/%s", project, zone, disk)
}
