package vdc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type AllowedHttpResponses []int

func (r AllowedHttpResponses) Contains(code int) bool {
	candidates := ([]int)(r)

	for i := range candidates {
		if candidates[i] == code {
			return true
		}
	}

	return false
}

type ValidateHttpRequest struct {
	DockerSocket string
	Verbose      bool

	ContainerID string
	Port        string
	Path        string
	Responses   AllowedHttpResponses
	Title       string
}

type ValidateResult struct {
	Valid    bool
	Messages []string
}

var htmlTitleExp = regexp.MustCompile(`<title>([^<]+)</title>`)

func (req ValidateHttpRequest) validate() error {
	if req.Port == "" {
		return errors.New("port must be provided")
	}
	if len(req.Responses) == 0 {
		return errors.New("allowed http responses must be provided")
	}

	return nil
}

func (req ValidateHttpRequest) log() {
	if req.Verbose {
		log.Printf("Validating container %s for http:\n", req.ContainerID)
		log.Printf("Port: %s\n", req.Port)
		log.Printf("Path: %s\n", req.Path)
		log.Printf("Title: %s\n", req.Title)
		log.Printf("Allowed responses: %+v\n", req.Responses)
	}
}

func ValidateHttp(req ValidateHttpRequest) (*ValidateResult, error) {
	err := req.validate()
	if err != nil {
		return nil, err
	}

	req.log()
	result := &ValidateResult{Valid: true}

	dockerClient, err := docker.NewClient(req.DockerSocket)
	if err != nil {
		return nil, err
	}

	container, err := dockerClient.InspectContainer(req.ContainerID)
	if err != nil {
		return nil, err
	}

	mappedPort, err := determineMappedPort(req, *container)
	if err != nil {
		result.Valid = false
		result.Messages = append(result.Messages, err.Error())
		return result, nil
	}

	url := requestUrl(req, *mappedPort)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkHttpResponse(req, *resp, result)

	return result, nil
}

func ValidateHttps(req ValidateHttpRequest) (*ValidateResult, error) {
	return nil, nil
}

func determineMappedPort(req ValidateHttpRequest, container docker.Container) (*docker.PortBinding, error) {
	mappedPort, ok := container.NetworkSettings.Ports[docker.Port(req.Port)]
	if ok == false {
		return nil, errors.New("Container " + container.ID + " did not expose port " + req.Port)
	}

	if req.Verbose {
		log.Printf("Container: %+v\n", container)
		log.Printf("NetworkSettings: %+v\n", container.NetworkSettings)
	}

	return &mappedPort[0], nil
}

func requestUrl(req ValidateHttpRequest, binding docker.PortBinding) string {
	url := "http://" + binding.HostIp + ":" + binding.HostPort
	if req.Path != "" {
		if strings.HasPrefix(req.Path, "/") {
			url += req.Path
		} else {
			url += "/" + req.Path
		}
	}

	return url
}

func checkHttpResponse(req ValidateHttpRequest, resp http.Response, result *ValidateResult) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if req.Verbose {
		log.Printf("Http response status code: %s\n", resp.Status)
	}

	if !req.Responses.Contains(resp.StatusCode) {
		result.Valid = false
		message := fmt.Sprintf("Invalid response status code: %d\n", resp.StatusCode)
		result.Messages = append(result.Messages, message)
	}

	if req.Title != "" {
		matches := htmlTitleExp.FindAllStringSubmatch(string(body), -1)
		if len(matches) == 0 {
			if req.Verbose {
				log.Println("Response did not contain a title")
			}
		}

		responseTitle := matches[0][1]
		if responseTitle != req.Title {
			result.Valid = false
			result.Messages = append(result.Messages, "Title did not match")
		}
	}

	return nil
}
