package vdc

import (
	"errors"
	"log"

	"github.com/fsouza/go-dockerclient"
)

func determineMappedPort(port string, container docker.Container, verbose bool) (*docker.PortBinding, error) {
	mappedPort, ok := container.NetworkSettings.Ports[docker.Port(port)]
	if ok == false {
		return nil, errors.New("Container " + container.ID + " did not expose port " + port)
	}

	if verbose {
		log.Printf("Container: %+v\n", container)
		log.Printf("NetworkSettings: %+v\n", container.NetworkSettings)
	}

	return &mappedPort[0], nil
}
