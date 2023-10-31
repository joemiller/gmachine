package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joemiller/gmachine/internal/config"
	"github.com/stretchr/testify/assert"
)

func tempFile(t *testing.T, contents string) string {
	tmpfile := filepath.Join(t.TempDir(), "temp.yaml")
	err := os.WriteFile(tmpfile, []byte(contents), 0o600)
	assert.NoError(t, err)
	return tmpfile
}

func TestLoadFile_valid(t *testing.T) {
	contents := `
---
version: 1
default: foo
machines:
  - name: foo
    account: my-account
    project: my-proj
    zone: us-central1-a
    csek:
      - uri: "https://www.googleapis.com/compute/beta/projects/pantheon-sandbox/zones/us-central1-b/disks/joem-buildbox"
        key: "acXTX3rxrKAFTF0tYVLvydU1riRZTvUNC4g5I11NY+c="
        key-type: "raw"
`
	tmpfile := tempFile(t, contents)

	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)

	// globals
	assert.Equal(t, "foo", cfg.Default)

	// machines:
	machine := cfg.Machines[0]
	assert.Equal(t, "foo", machine.Name)
	assert.Equal(t, "my-account", machine.Account)
	assert.Equal(t, "my-proj", machine.Project)
	assert.Equal(t, "us-central1-a", machine.Zone)

	// verify CSEK details loaded correctly:
	csek := machine.CSEK[0]
	assert.Equal(t, "https://www.googleapis.com/compute/beta/projects/pantheon-sandbox/zones/us-central1-b/disks/joem-buildbox", csek.URI)
	assert.Equal(t, "acXTX3rxrKAFTF0tYVLvydU1riRZTvUNC4g5I11NY+c=", csek.Key)
	assert.Equal(t, "raw", csek.KeyType)
}

func TestLoadFile_file_does_not_exist(t *testing.T) {
	cfg, err := config.LoadFile("/no/such/file")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestLoadFile_corrupt_file(t *testing.T) {
	contents := `asdfasdfasdfkjaskdfjaslfdf`
	tmpfile := tempFile(t, contents)

	cfg, err := config.LoadFile(tmpfile)
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadFile_empty_file(t *testing.T) {
	tmpfile := tempFile(t, "")

	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestAdd(t *testing.T) {
	tmpfile := tempFile(t, "")

	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	err = cfg.Add("foo-machine", "my-account", "my-proj", "us-west1-a", nil)
	assert.NoError(t, err)
	// there should be 1 machine in the config now
	assert.Equal(t, 1, cfg.Count())
	// the first machine added should also be marked as the default
	assert.Equal(t, "foo-machine", cfg.GetDefault())

	// adding a machine that already exists in the config should error
	err = cfg.Add("foo-machine", "my-account", "my-proj", "us-west1-a", nil)
	assert.Error(t, err)

	// read in the saved config file, it should contain the added machine
	cfg2, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)
}

func TestGet(t *testing.T) {
	// create new empty config
	tmpfile := tempFile(t, "")
	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Add node
	err = cfg.Add("foo", "my-account", "my-proj", "zone1", nil)
	assert.NoError(t, err)

	// read it back and verify
	m, err := cfg.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foo", m.Name)
	assert.Equal(t, "my-account", m.Account)
	assert.Equal(t, "my-proj", m.Project)
	assert.Equal(t, "zone1", m.Zone)

	// get on a non-existent node should return error
	_, err = cfg.Get("no-such-node")
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	// create new empty config
	tmpfile := tempFile(t, "")
	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Add 2 nodes
	err = cfg.Add("foo", "my-account", "my-proj", "zone1", nil)
	assert.NoError(t, err)
	err = cfg.Add("bar", "my-account", "my-proj", "zone1", nil)
	assert.NoError(t, err)

	// should have 2 nodes
	assert.Equal(t, 2, cfg.Count())

	// delete node
	err = cfg.Delete("foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.Count())

	// deleting a non-existent node should error
	err = cfg.Delete("no-such-node")
	assert.Error(t, err)

	// read in the saved config file, it should only contain 1 node
	cfg2, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg2.Count())
}

func TestGetSetDefault(t *testing.T) {
	// create new empty config
	tmpfile := tempFile(t, "")
	cfg, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Add 1st node, it should be set as the default automatically
	err = cfg.Add("foo", "my-account", "my-proj", "zone1", nil)
	assert.NoError(t, err)
	assert.Equal(t, "foo", cfg.GetDefault())

	// add another node and set it as the new default
	err = cfg.Add("bar", "my-account", "my-proj", "zone1", nil)
	assert.NoError(t, err)
	err = cfg.SetDefault("bar")
	assert.NoError(t, err)
	assert.Equal(t, "bar", cfg.GetDefault())

	// attempting to set an unknown machine as the default should error
	err = cfg.SetDefault("no-such-machine")
	assert.Error(t, err)

	// read in the saved config file, ensure the last successful call to SetDefault was persisted
	cfg2, err := config.LoadFile(tmpfile)
	assert.NoError(t, err)
	assert.Equal(t, "bar", cfg2.GetDefault())
}
