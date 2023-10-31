# gmachine

Command-line utility for managing cloud workstations on Google Compute Engine.

Name: .. gmachines :shrug: I couldn't think of a better name (yet).

## Install

**Pre-requisites:**

* Install and configure `gcloud` CLI: https://cloud.google.com/sdk/docs/install

**Install:**

* macOS (Linuxbrew might work too): `brew install joemiller/taps/gmachine`
* Binaries for all platforms (macOS, Linux, *BSD) on [GitHub Releases](https://github.com/joemiller/gmachine/releases)

## Configuration

### `gmachine.yaml` config file

State is stored in the `gmachine.yaml` file. Normally you do not need to modify this file (although you can). The `gmachine` util
will update the file as needed.

VM's created and managed by `gmachine` are tracked in the `gmachine.yaml` file. This file will be located at the system defined
[`os.UserConfigDir`](https://pkg.go.dev/os#UserConfigDir). Typically `~/Library/Application Support/gmachine/gmachine.yaml` on macOS
and `~/.config/gmachine/gmachine.yaml` on Linux.

Set the `GMACHINE_CONFIG_DIR` environment variable to override the default location.

The `gmachine.yaml` file will contain private key material (CSEK) if you've created a VM with the `gmachine create --csek` flag. Protect
it with `0600` permissions.

## Usage

> :construction: TODO/WIP... for now run `gmachine` with no arguments for list of commands. Some commands are documented below:


### `gmachine create`

Create a new VM in `my-project`, `us-west2-a`.

```console
gmachine create my-workstation \
  -p my-project \
  -z us-west2-a \
  --disk-size 100GB \
  --disk-type pd-ssd \
  --image-family ubuntu-2204-lts \
  --machine-type n2d-standard-32
```

Optionally, add the `--csek` flag to create a CSEK (Customer-Supplied Encryption Key) encrypted disk.
The private key is stored in the `gmachine.yaml` config file. The key is necessary to decrypt the root disk and start the VM.
This makes your VM more secure by requiring the presence of the CSEK key at boot time. Keep in mind that users with
access to the GCP project may still be able to SSH into your VM while it is running if they have access to add SSH
keys to the instance and the local GCP guest tools are running.

CSEK keys also limit some functionality such as instance suspend which is not available with CSEK-encrypted VMs.

### `gmachine status`

Run `gmachine status -a` to list all VMs in your `gmachine.yaml` file.

## Recipes and Use Cases

### Cloud Workstation

Creating and managing one or more "cloud workstations" was the original motivation for writing this util. Here's what I do:

```console
gmachine create my-workstation \
  -p my-project \
  -z us-west2-a \
  --csek \
  --disk-size 100GB \
  --disk-type pd-ssd \
  --image-family ubuntu-2204-lts \
  --machine-type n2d-standard-8
```

The `--csek` flag ensures no one besides you (or someone with the private key from your `gmachine.yaml` file) can decrypt
the root disk and boot the VM.

If you share the GCP project with other users you may also consider adding `--disable-ssh-project-keys`.`

If you need a larger or smaller node use `gmachine resize` to change the machine-type.

> :construction: TODO: Add a flag to remove public IP. Then you could use `gmachine ssh` to SSH through the IAP proxy or via tailscale. For now you can remove the public IP manually.

### Tailscale Exit Nodes

Spin up a cheap tailscale exit node with `--startup-script`:

- `tailscale-exit-node.sh`
```sh
#!/usr/bin/env bash

set -eou pipefail

echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' | sudo tee -a /etc/sysctl.conf
sysctl -p /etc/sysctl.conf

curl -fsSL https://tailscale.com/install.sh | sh
tailscale up --advertise-exit-node --authkey <YOUR_TAILSCALE_AUTH_KEY_HERE>
```

```console
gmachine create ts-exit-europe-west1 \
  -p my-project \
  -z europe-west1-c \
  --disk-size 10GB \
  --disk-type pd-ssd \
  --image-family ubuntu-2204-lts \
  --machine-type f1-micro \
  --no-service-account \
  --startup-script tailscale-exit-node.sh
```

You may need to authorize the node to allow it to be an exit-node, unless you've setup [auto-approval](https://tailscale.com/kb/1018/acls/#auto-approvers-for-routes-and-exit-nodes) in your Tailnet ACL.

## Releases

Releases are cut automatically on a successful main branch build. This project uses
[autotag](https://github.com/pantheon-systems/autotag) and [goreleaser](https://goreleaser.com/) to automate this process.

Semver (`vMajor.Minor.Patch`) is used for versioning and releases. By default, autotag will bump the patch version
on a successful master build, eg: `v1.0.0` -> `v1.0.1`.

To bump the major or minor release instead, include the text `[major]` or `[minor]` in the commit message.
See the autotag [docs](https://github.com/pantheon-systems/autotag#incrementing-major-and-minor-versions) for more details.

To prevent a new release being built, include `[ci skip]` in the commit message. Only use this for things like
documentation updates.