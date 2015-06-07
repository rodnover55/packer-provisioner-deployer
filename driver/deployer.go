package deployer

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"os"
	"path/filepath"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	BeforeInstall   []string `mapstructure:"before_install"`
	DeployerCommand string   `mapstructure:"bin"`
	SkipInstall     bool     `mapstructure:"skip_install"`
	RemoteUrl       string   `mapstructure:"url"`
	File            string   `mapstructure:"file"`
	Task            string   `mapstructure:"task"`
	StagingDir      string   `mapstructure:"staging_directory"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) download(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Installing deployer...")

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

	if p.config.File == "" {
		p.config.File = "deploy.php"
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-deployer"
	}

	if p.config.Task == "" {
		p.config.Task = "deploy"
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

	ui.Say("Provisioning with deployer")

	path := filepath.Dir(p.config.File)

	err = p.uploadDirectory(ui, comm, p.config.StagingDir, path)

	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("cd %s && %s %s", p.config.StagingDir, p.config.DeployerCommand, p.config.Task)}

	err = cmd.StartWithUi(comm, ui)

	if err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf(
			"Command '%s' exited with non-zero exit status %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) uploadDirectory(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	// Make sure there is a trailing "/" so that the directory isn't
	// created on the other side.
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	return comm.UploadDir(dst, src, nil)
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
