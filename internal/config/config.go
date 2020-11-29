package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/joemiller/gmachine/internal/gcp"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v2"
)

type config struct {
	Version  int       `mapstructure:"version"`
	Default  string    `mapstructure:"default"`
	Machines []machine `mapstructure:"machines"`
	filename string
	mu       sync.RWMutex
}

type machine struct {
	Name    string         `mapstructure:"name"`
	Project string         `mapstructure:"project"`
	Zone    string         `mapstructure:"zone"`
	CSEK    gcp.CSEKBundle `mapstructure:"csek"`
}

func newConfig() *config {
	return &config{Version: 1}
}

// TODO document
func LoadFile(file string) (*config, error) {
	cfg := newConfig()
	cfg.filename = file

	path, err := homedir.Expand(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file path %s: %w", file, err)
	}

	// no config file yet, return an empty config{}
	if !fileExists(path) {
		return cfg, nil
	}

	// nsure the file is writable
	if !writable(path) {
		return cfg, fmt.Errorf("config file %s is not writable", path)
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %w", path, err)
	}
	return cfg, nil
}

// TODO document
// TODO do we need Exist()? Could we just use Get instead?
func (c *config) Exists(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.Machines {
		if m.Name == name {
			return true
		}
	}
	return false
}

// TODO document
// TODO tests
func (c *config) Get(name string) (machine, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.Machines {
		if m.Name == name {
			return m, nil
		}
	}
	return machine{}, errors.New("machine not found")
}

// TODO make a thread-safe getter for config.Machines

// TODO document
func (c *config) Add(name, project, zone string, csek gcp.CSEKBundle) error {
	// fail if already exists
	if c.Exists(name) {
		return fmt.Errorf("machine '%s' already exists", name)
	}

	if csek == nil {
		csek = gcp.CSEKBundle{}
	}

	// add to config.Machines array
	m := machine{
		Name:    name,
		Project: project,
		Zone:    zone,
		CSEK:    csek,
	}
	c.mu.Lock()
	c.Machines = append(c.Machines, m)
	// If this is the only machine in the database, mark it as the new default
	if len(c.Machines) == 1 {
		c.Default = name
	}
	c.mu.Unlock()

	return c.save()
}

// TODO document
func (c *config) Delete(name string) error {
	if !c.Exists(name) {
		return fmt.Errorf("machine '%s' does not exist", name)
	}

	c.mu.Lock()
	for i, m := range c.Machines {
		if m.Name == name {
			c.Machines = append(c.Machines[:i], c.Machines[i+1:]...)
			break
		}
	}
	// if the deleted machine was the default, unset the default machine
	if c.Default == name {
		c.Default = ""
	}
	c.mu.Unlock()

	return c.save()
}

// TODO document
func (c *config) SetDefault(name string) error {
	if name != "" && !c.Exists(name) {
		return fmt.Errorf("machine '%s' does not exist", name)
	}
	c.mu.Lock()
	c.Default = name
	c.mu.Unlock()
	return c.save()
}

// TODO document
func (c *config) GetDefault() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Default
}

// TODO document
func (c *config) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.Machines)
}

// save persist the control cluster cache to a file in JSON format
func (c *config) save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	yamlBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.filename, yamlBytes, 0600)
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// writable returns true if the specified path is writable by the current process
func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}
