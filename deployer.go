package deployer

import (
	"github.com/mitchellh/packer/packer/plugin"
)

func (p *Provisioner) Prepare(raws ...interface{}) error {

}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
