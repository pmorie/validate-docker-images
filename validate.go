package vdc

type ValidateRequest struct {
	DockerSocket string
	Verbose      bool
	ContainerID  string
}

type ValidateResult struct {
	Valid    bool
	Messages []string
}
