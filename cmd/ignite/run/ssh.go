package run

import (
	"fmt"
	"path"

	"github.com/weaveworks/ignite/pkg/constants"
	"github.com/weaveworks/ignite/pkg/metadata/vmmd"
	"github.com/weaveworks/ignite/pkg/util"
)

type SSHFlags struct {
	IdentityFile string
}

type sshOptions struct {
	*SSHFlags
	vm *vmmd.VM
}

func (sf *SSHFlags) NewSSHOptions(vmMatch string) (so *sshOptions, err error) {
	so = &sshOptions{SSHFlags: sf}
	so.vm, err = getVMForMatch(vmMatch)
	return
}

func SSH(so *sshOptions) error {
	// Check if the VM is running
	if !so.vm.Running() {
		return fmt.Errorf("VM %q is not running", so.vm.GetUID())
	}

	ipAddrs := so.vm.Status.IPAddresses
	if len(ipAddrs) == 0 {
		return fmt.Errorf("VM %q has no usable IP addresses", so.vm.GetUID())
	}

	// Auto-accept the "The authenticity of host **** can't be established" warning with
	// -o StrictHostKeyChecking=no, we're dealing with local VMs in a local subnet which is trusted.
	sshArgs := append(make([]string, 0, 3),
		fmt.Sprintf("root@%s", ipAddrs[0]),
		"-o",
		"StrictHostKeyChecking=no",
		"-i")

	// If an external identity file is specified, use it instead of the internal one
	if len(so.IdentityFile) > 0 {
		sshArgs = append(sshArgs, so.IdentityFile)
	} else {
		privKeyFile := path.Join(so.vm.ObjectPath(), fmt.Sprintf(constants.VM_SSH_KEY_TEMPLATE, so.vm.GetUID()))
		if !util.FileExists(privKeyFile) {
			return fmt.Errorf("no private key found for VM %q", so.vm.GetUID())
		}

		sshArgs = append(sshArgs, privKeyFile)
	}

	// SSH into the vm
	if _, err := util.ExecForeground("ssh", sshArgs...); err != nil {
		return fmt.Errorf("SSH into VM %q failed: %v", so.vm.GetUID(), err)
	}
	return nil
}
