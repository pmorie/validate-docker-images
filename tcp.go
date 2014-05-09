package vdc

import (
	"errors"
	"net"

	"github.com/fsouza/go-dockerclient"
)

type ValidateTcpRequest struct {
	ValidateRequest

	Port string
}

func (req ValidateTcpRequest) Validate() error {
	if req.Port == "" {
		return errors.New("port must be provided")
	}

	return nil
}

func ValidateTcp(req ValidateTcpRequest) (*ValidateResult, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	result := &ValidateResult{Valid: true}

	dockerClient, err := docker.NewClient(req.DockerSocket)
	if err != nil {
		return nil, err
	}

	container, err := dockerClient.InspectContainer(req.ContainerID)
	if err != nil {
		return nil, err
	}

	mappedPort, err := determineMappedPort(req.Port, *container, req.Verbose)
	if err != nil {
		result.Valid = false
		result.Messages = append(result.Messages, err.Error())
		return result, nil
	}

	address := mappedPort.HostIp + ":" + mappedPort.HostPort
	_, err = net.Dial("tcp", address)
	if err != nil {
		result.Valid = false
		result.Messages = append(result.Messages, "Connection failed: "+err.Error())
	}

	return result, nil
}
