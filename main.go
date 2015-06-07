package main

import (
	"fmt"
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/rodnover55/packer-provisioner-deployer/driver"
)

func main() {
	server, err := plugin.Server()

	if err != nil {
		panic(err)
	}

	server.RegisterProvisioner(new(deployer.Provisioner))
	server.Serve()

	fmt.Println("test")
}
