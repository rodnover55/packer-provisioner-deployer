package deployer

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"os"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	BeforeInstall   []string `mapstructure:"before_install"`
	DeployerCommand string   `mapstructure:"bin"`
	SkipInstall     bool     `mapstructure:"skip_install"`
	RemoteUrl       string   `mapstructure:"url"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) download(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Installing Chef...")

	commands := append(p.config.BeforeInstall,
		fmt.Sprintf("curl -L %s > %s", p.config.RemoteUrl, p.config.DeployerCommand),
		fmt.Sprintf("chmod +x %s", p.config.DeployerCommand))

	for _, command := range commands {
		cmd := &packer.RemoteCmd{Command: command}
		err := cmd.StartWithUi(comm, ui)

		if err != nil {
			return err
		}

		if cmd.ExitStatus != 0 {
			return fmt.Errorf(
				"Command '%s' exited with non-zero exit status %d", cmd.ExitStatus)
		}
	}

	return nil
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	fmt.Println(raws)
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		}}, raws...)

	if err != nil {
		return err
	}

	if p.config.DeployerCommand == "" {
		p.config.DeployerCommand = "/usr/local/bin/dep"
	}

	if p.config.RemoteUrl == "" {
		p.config.RemoteUrl = "http://deployer.org/deployer.phar"
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	_, err := os.Stat(p.config.DeployerCommand)

	if !p.config.SkipInstall && os.IsNotExist(err) {
		err := p.download(ui, comm)

		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
